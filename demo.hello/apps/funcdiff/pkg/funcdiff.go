package funcdiff

import (
	"log"
	"sort"
)

// DiffEntry func diff result entry.
type DiffEntry struct {
	FuncName string
	IsDiff   bool
}

// DiffEntries list of DiffEntry.
type DiffEntries []DiffEntry

// Swap implements sort interface.
func (entries DiffEntries) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

// Len implements sort interface.
func (entries DiffEntries) Len() int {
	return len(entries)
}

func (entries DiffEntries) Less(i, j int) bool {
	return entries[i].IsDiff
}

/*
Return modified funcs for current pr:
1. Get changed files from git commits
2. Compare funcs for diff version of changed file
*/

// FuncDiff compares same func from diff src files. Return true:diff, false:same
func FuncDiff(sPath, tPath string, funcName string) (bool, error) {
	sInfo, err := GetFuncInfo(sPath, funcName)
	if err != nil {
		return false, err
	}
	tInfo, err := GetFuncInfo(tPath, funcName)
	if err != nil {
		return false, err
	}
	if sInfo.StmtCount != tInfo.StmtCount {
		log.Printf("func total statements diff: %s[%d] and %s[%d]\n", sPath, sInfo.StmtCount, tPath, tInfo.StmtCount)
		return false, nil
	}
	return funcSrcDiff(sPath, tPath, sInfo, tInfo)
}

func funcSrcDiff(sPath, tPath string, sInfo, tInfo *FuncInfo) (bool, error) {
	sFuncSrc, err := GetFuncSrc(sPath, sInfo)
	if err != nil {
		return false, err
	}
	tFuncSrc, err := GetFuncSrc(tPath, tInfo)
	if err != nil {
		return false, err
	}
	return !(sFuncSrc == tFuncSrc), nil
	// TODO: handle for line with comments
}

// GoFileDiffByFunc compares diff go src file, and returns func diff info. Return true:diff, false:same
func GoFileDiffByFunc(sPath, tPath string) (DiffEntries, error) {
	sInfos, err := GetFileAllFuncInfos(sPath)
	if err != nil {
		return nil, err
	}
	tInfos, err := GetFileAllFuncInfos(tPath)
	if err != nil {
		return nil, err
	}

	if len(sInfos) < len(tInfos) {
		sPath, tPath = tPath, sPath
		sInfos, tInfos = tInfos, sInfos
	}

	ret := make(DiffEntries, len(sInfos))
	for idx, sInfo := range sInfos {
		diffResult := true
		for _, tInfo := range tInfos {
			if sInfo.FuncName == tInfo.FuncName && sInfo.StmtCount == tInfo.StmtCount {
				if diffResult, err = funcSrcDiff(sPath, tPath, sInfo, tInfo); err != nil {
					return nil, err
				}
			}
		}
		ret[idx] = DiffEntry{
			FuncName: sInfo.FuncName,
			IsDiff:   diffResult,
		}
	}
	sort.Sort(ret)
	return ret, nil
}

/*
func diff extend
*/

// DiffEntryExt func diff result entry. (TODO: use map instead of struct for high perf in iterator.)
type DiffEntryExt struct {
	FuncName string
	// DiffType: add, del, change and same
	DiffType string
}

// GoFileDiffByFuncExt compares diff go src file, and returns func diff info for add, del, change and same.
func GoFileDiffByFuncExt(sPath, tPath string) ([]DiffEntryExt, error) {
	sFuncInfos, err := GetFileAllFuncInfos(sPath)
	if err != nil {
		return nil, err
	}
	tFuncInfos, err := GetFileAllFuncInfos(tPath)
	if err != nil {
		return nil, err
	}

	// 交集
	sameFuncInfos := make([]*FuncInfo, 0, len(sFuncInfos))
	for _, sInfo := range sFuncInfos {
		for _, tInfo := range tFuncInfos {
			if sInfo.FuncName == tInfo.FuncName {
				sameFuncInfos = append(sameFuncInfos, sInfo)
				break
			}
		}
	}

	// 差集 get del funcs
	retDiffEntries := make([]DiffEntryExt, 0, len(tFuncInfos))
	delDiffEntries := getDiffEntriesOfSlice(sFuncInfos, sameFuncInfos, "del")
	retDiffEntries = append(retDiffEntries, delDiffEntries...)
	// 差集 get add funcs
	addDiffEntries := getDiffEntriesOfSlice(tFuncInfos, sameFuncInfos, "add")
	retDiffEntries = append(retDiffEntries, addDiffEntries...)

	// get change funcs
	for _, sInfo := range sameFuncInfos {
		diff := false
		for _, tInfo := range tFuncInfos {
			if sInfo.FuncName == tInfo.FuncName && sInfo.StmtCount == tInfo.StmtCount {
				if diff, err = funcSrcDiff(sPath, tPath, sInfo, tInfo); err != nil {
					return nil, err
				}
			}
		}
		if diff {
			retDiffEntries = append(retDiffEntries, DiffEntryExt{
				FuncName: sInfo.FuncName,
				DiffType: "change",
			})
		} else {
			retDiffEntries = append(retDiffEntries, DiffEntryExt{
				FuncName: sInfo.FuncName,
				DiffType: "same",
			})
		}
	}

	return retDiffEntries, nil
}

func getDiffEntriesOfSlice(sFuncInfos []*FuncInfo, baseFuncInfos []*FuncInfo, diffType string) []DiffEntryExt {
	retDiffEntries := make([]DiffEntryExt, 0)
	for _, sInfo := range sFuncInfos {
		found := false
		for _, info := range baseFuncInfos {
			if sInfo.FuncName == info.FuncName {
				found = true
				break
			}
		}
		if !found {
			retDiffEntries = append(retDiffEntries, DiffEntryExt{
				FuncName: sInfo.FuncName,
				DiffType: diffType,
			})
		}
	}

	return retDiffEntries
}
