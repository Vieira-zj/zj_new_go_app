package pkg

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

//
// Diff a func of .go file.
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

// FuncProfileEntry .
type FuncProfileEntry struct {
	FuncInfo      *FuncInfo      `json:"func_info"`
	ProfileBlocks []ProfileBlock `json:"profile_blocks"`
}

// DiffEntry func diff result entry.
type DiffEntry struct {
	SrcFuncProfileEntry *FuncProfileEntry `json:"src_func_profile_entry,omitempty"`
	DstFuncProfileEntry *FuncProfileEntry `json:"dst_func_profile_entry,omitempty"`
	Result              string            `json:"result"`
}

// funcDiff compares func between src and dst .go files, and return diff, same
func funcDiff(srcPath, dstPath string, funcName string) (*DiffEntry, error) {
	srcFuncInfo, err := GetFuncInfo(srcPath, nil, funcName)
	if err != nil {
		return nil, err
	}
	dstFuncInfo, err := GetFuncInfo(dstPath, nil, funcName)
	if err != nil {
		return nil, err
	}

	result := diffFuncSrc(srcFuncInfo, dstFuncInfo)
	diffEntry := &DiffEntry{
		SrcFuncProfileEntry: &FuncProfileEntry{
			FuncInfo: srcFuncInfo,
		},
		DstFuncProfileEntry: &FuncProfileEntry{
			FuncInfo: dstFuncInfo,
		},
		Result: result,
	}
	return diffEntry, nil
}

func diffFuncSrc(srcFuncInfo, dstFuncInfo *FuncInfo) string {
	if srcFuncInfo.StmtCount == dstFuncInfo.StmtCount {
		src := deleteEmptyLinesInText([]byte(srcFuncInfo.Source))
		dst := deleteEmptyLinesInText([]byte(dstFuncInfo.Source))
		if len(src) == len(dst) && strings.EqualFold(src, dst) {
			return resultSame
		}
	}
	return resultDiff
}

func diffFuncSrcDeprecated(srcBody, dstBody []byte, srcFuncInfo, dstFuncInfo *FuncInfo) string {
	srcFuncBody := GetFuncSrc(srcBody, srcFuncInfo)
	dstFuncBody := GetFuncSrc(dstBody, dstFuncInfo)
	if len(srcFuncBody) == len(dstFuncBody) && bytes.Equal(srcFuncBody, dstFuncBody) {
		return resultSame
	}
	return resultDiff
}

//
// Diff funcs of .go file.
//

// DiffEntries list of DiffEntry.
type DiffEntries []*DiffEntry

func (e DiffEntries) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e DiffEntries) Len() int      { return len(e) }
func (e DiffEntries) Less(i, j int) bool {
	return e[i].Result[0] > e[j].Result[0]
}

// funcDiffForGoFiles compares func bewteen src and dst .go files, and returns func diff info.
func funcDiffForGoFiles(srcPath, dstPath string) (DiffEntries, error) {
	srcFuncInfos, err := GetFuncInfos(srcPath, nil)
	if err != nil {
		return nil, err
	}
	dstFuncInfos, err := GetFuncInfos(dstPath, nil)
	if err != nil {
		return nil, err
	}

	// 交集
	sameFuncInfos := getSameFuncInfos(srcFuncInfos, dstFuncInfos)

	retDiffEntries := make([]*DiffEntry, 0, len(srcFuncInfos))
	// 差集 get del funcs
	delDiffEntries := getDiffEntries(srcFuncInfos, sameFuncInfos, diffTypeRemove)
	retDiffEntries = append(retDiffEntries, delDiffEntries...)
	// 差集 get add funcs
	addDiffEntries := getDiffEntries(dstFuncInfos, sameFuncInfos, diffTypeAdd)
	retDiffEntries = append(retDiffEntries, addDiffEntries...)

	// 交集 get change funcs
	for _, dstFuncInfo := range dstFuncInfos {
		if dstFuncInfo.Name == "main" {
			continue
		}
		if srcFuncInfo, ok := sameFuncInfos[dstFuncInfo.Name]; ok {
			result := diffFuncSrc(srcFuncInfo, dstFuncInfo)
			diffEntry := &DiffEntry{
				SrcFuncProfileEntry: &FuncProfileEntry{
					FuncInfo: srcFuncInfo,
				},
				DstFuncProfileEntry: &FuncProfileEntry{
					FuncInfo: dstFuncInfo,
				},
				Result: diffResultAndTypeMap[result],
			}
			retDiffEntries = append(retDiffEntries, diffEntry)
		}
	}

	sort.Sort(DiffEntries(retDiffEntries))
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
				diffEntry.SrcFuncProfileEntry = &FuncProfileEntry{FuncInfo: funcInfo}
			} else {
				diffEntry.DstFuncProfileEntry = &FuncProfileEntry{FuncInfo: funcInfo}
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
	retLines := make([]string, 0, 16)
	if entry.SrcFuncProfileEntry != nil {
		lines := prettyFuncProfileEntry(entry.SrcFuncProfileEntry, "src")
		retLines = append(retLines, lines...)
	}
	if entry.DstFuncProfileEntry != nil {
		lines := prettyFuncProfileEntry(entry.DstFuncProfileEntry, "dst")
		retLines = append(retLines, lines...)
	}
	retLines = append(retLines, "diff: "+entry.Result+"\n")
	return strings.Join(retLines, "\n")
}

func prettyFuncProfileEntry(entry *FuncProfileEntry, tag string) []string {
	lines := make([]string, 0, 16)
	lines = append(lines, tag+": "+prettySprintFuncInfo(entry.FuncInfo))
	if entry.ProfileBlocks != nil {
		for _, block := range entry.ProfileBlocks {
			lines = append(lines, "\t"+fmt.Sprintf("%+v", block))
		}
	}
	return lines
}

func prettySprintFuncInfo(info *FuncInfo) string {
	return fmt.Sprintf("[%s:%s] [%d:%d,%d:%d]",
		info.Path, info.Name, info.StartLine, info.StartCol, info.EndLine, info.EndCol)
}
