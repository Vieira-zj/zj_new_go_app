package pkg

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
)

//
// Diff specified func of .go file.
//

const (
	resultDiff = "diff"
	resultSame = "same"

	diffTypeAdd    = "add"
	diffTypeChange = "change"
	diffTypeRemove = "remove"
	diffTypeSame   = "same"
)

var diffResultAndTypeMap = map[string]string{
	resultDiff: diffTypeChange,
	resultSame: diffTypeSame,
}

// DiffEntry func diff result entry.
type DiffEntry struct {
	SrcFnInfo *FuncInfo `json:"src_func_info,omitempty"`
	DstFnInfo *FuncInfo `json:"dst_func_info,omitempty"`
	Result    string    `json:"result"`
}

// FuncDiff compares func between src and dst .go files, and return diff, same
func FuncDiff(srcPath, dstPath string, funcName string) (*DiffEntry, error) {
	srcFuncInfo, err := GetFuncInfo(srcPath, funcName)
	if err != nil {
		return nil, err
	}
	dstFuncInfo, err := GetFuncInfo(dstPath, funcName)
	if err != nil {
		return nil, err
	}

	diffEntry := &DiffEntry{
		SrcFnInfo: srcFuncInfo,
		DstFnInfo: dstFuncInfo,
	}
	if srcFuncInfo.StmtCount != dstFuncInfo.StmtCount {
		diffEntry.Result = resultDiff
		return diffEntry, nil
	}

	srcBody, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}
	dstBody, err := os.ReadFile(dstPath)
	if err != nil {
		return nil, err
	}

	result, err := funcSrcDiff(srcBody, dstBody, srcFuncInfo, dstFuncInfo)
	if err != nil {
		return nil, err
	}
	diffEntry.Result = result
	return diffEntry, nil
}

func funcSrcDiff(srcBody, dstBody []byte, srcFuncInfo, dstFuncInfo *FuncInfo) (string, error) {
	srcFuncBody := GetFuncSrc(srcBody, srcFuncInfo)
	dstFuncBody := GetFuncSrc(dstBody, dstFuncInfo)
	if len(srcFuncBody) == len(dstFuncBody) && bytes.Equal(srcFuncBody, dstFuncBody) {
		return resultSame, nil
	}
	return resultDiff, nil
}

//
// Diff funcs of .go file.
//

// DiffEntries list of DiffEntry.
type DiffEntries []*DiffEntry

// Swap implements sort interface.
func (entries DiffEntries) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

// Len implements sort interface.
func (entries DiffEntries) Len() int {
	return len(entries)
}

func (entries DiffEntries) Less(i, j int) bool {
	return entries[i].Result[0] > entries[j].Result[0]
}

// GoFileDiffFunc compares func bewteen src and dst .go files, and returns func diff info.
func GoFileDiffFunc(srcPath, dstPath string) (DiffEntries, error) {
	srcFuncInfos, err := GetFuncInfos(srcPath)
	if err != nil {
		return nil, err
	}
	dstFuncInfos, err := GetFuncInfos(dstPath)
	if err != nil {
		return nil, err
	}

	// 交集
	sameFuncInfos := getSameFuncInfos(srcFuncInfos, dstFuncInfos)

	var retDiffEntries DiffEntries
	retDiffEntries = make([]*DiffEntry, 0, len(srcFuncInfos))
	// 差集 get del funcs
	delDiffEntries := getDiffEntries(srcFuncInfos, sameFuncInfos, diffTypeRemove)
	retDiffEntries = append(retDiffEntries, delDiffEntries...)
	// 差集 get add funcs
	addDiffEntries := getDiffEntries(dstFuncInfos, sameFuncInfos, diffTypeAdd)
	retDiffEntries = append(retDiffEntries, addDiffEntries...)

	srcBody, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}
	dstBody, err := os.ReadFile(dstPath)
	if err != nil {
		return nil, err
	}

	// 交集 get change funcs
	for _, dstFuncInfo := range dstFuncInfos {
		if dstFuncInfo.Name == "main" {
			continue
		}
		if srcFuncInfo, ok := sameFuncInfos[dstFuncInfo.Name]; ok {
			result := resultDiff
			if srcFuncInfo.StmtCount == dstFuncInfo.StmtCount {
				if result, err = funcSrcDiff(srcBody, dstBody, srcFuncInfo, dstFuncInfo); err != nil {
					return nil, err
				}
			}
			diffEntry := &DiffEntry{
				SrcFnInfo: srcFuncInfo,
				DstFnInfo: dstFuncInfo,
				Result:    diffResultAndTypeMap[result],
			}
			retDiffEntries = append(retDiffEntries, diffEntry)
		}
	}

	sort.Sort(retDiffEntries)
	return retDiffEntries, nil
}

func getSameFuncInfos(srcFuncInfos, dstFuncInfos []*FuncInfo) map[string]*FuncInfo {
	sameFuncInfos := make(map[string]*FuncInfo, len(srcFuncInfos))
	for _, srcFuncInfo := range srcFuncInfos {
		for _, dstFuncInfo := range dstFuncInfos {
			if srcFuncInfo.Name == dstFuncInfo.Name {
				sameFuncInfos[srcFuncInfo.Name] = srcFuncInfo
				break
			}
		}
	}
	return sameFuncInfos
}

func getDiffEntries(funcInfos []*FuncInfo, baseFuncInfos map[string]*FuncInfo, diffType string) DiffEntries {
	retDiffEntries := make([]*DiffEntry, 0, 16)
	for _, funcInfo := range funcInfos {
		if _, ok := baseFuncInfos[funcInfo.Name]; !ok {
			diffEntry := &DiffEntry{
				Result: diffType,
			}
			if diffType == diffTypeRemove {
				diffEntry.SrcFnInfo = funcInfo
			} else {
				diffEntry.DstFnInfo = funcInfo
			}
			retDiffEntries = append(retDiffEntries, diffEntry)
		}
	}
	return retDiffEntries
}

//
// Pretty print.
//

func prettySprintDiffEntry(entry *DiffEntry) string {
	lines := make([]string, 0, 3)
	if entry.SrcFnInfo != nil {
		lines = append(lines, "src: "+prettySprintFuncInfo(entry.SrcFnInfo))
	}
	if entry.DstFnInfo != nil {
		lines = append(lines, "dst: "+prettySprintFuncInfo(entry.DstFnInfo))
	}
	lines = append(lines, "diff: "+entry.Result)
	return strings.Join(lines, "\n")
}

func prettySprintFuncInfo(info *FuncInfo) string {
	return fmt.Sprintf("[%s:%s] [%d:%d,%d:%d]", info.Path, info.Name, info.StartLine, info.StartCol, info.EndLine, info.EndCol)
}
