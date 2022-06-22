package pkg

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"
)

func TestPrint(t *testing.T) {
	msg := "hello"
	num := 101
	char := 'a'
	fmt.Printf("print: msg=%q, num=%q, ch=%q\n", msg, num, char)
}

func TestParseProfileLine(t *testing.T) {
	line := "demo.hello/echoserver/handlers/hooks.go:12.65,13.36 1 0"
	fname, profile, err := parseProfileLine(line)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("porfile: [%s]:%+v\n", fname, profile)
}

func TestParseCovFile(t *testing.T) {
	// read cov
	fpath := "/tmp/test/profile.cov"
	results, err := parseCovFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	for fname, profile := range results {
		fmt.Printf("\nmode=%s, filename=%s\n", profile.Mode, fname)
		for _, block := range profile.Blocks {
			fmt.Printf("\tblock: %+v\n", block)
		}
	}

	// write cov
	outProfiles := make([]*Profile, 0, len(results))
	for _, profile := range results {
		outProfiles = append(outProfiles, profile)
	}

	outPath := "/tmp/test/out_profile.cov"
	writeCovFile(outPath, outProfiles)
}

func TestLinkProfileBlocksToFunc(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	funcInfos, err := GetFuncInfos(srcPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("funcs info:")
	for _, info := range funcInfos {
		fmt.Println("\t", info.Path, info.Name)
	}

	fmt.Println("\nprofile:")
	covPath := filepath.Join(filepath.Dir(srcPath), "profile.cov")
	fnProfile, err := parseCovFile(covPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range fnProfile {
		fmt.Println(profile.FileName)
		for _, block := range profile.Blocks {
			fmt.Printf("\t%+v\n", block)
		}
	}

	fmt.Println("\nlink profile blocks to func:")
	for _, fnInfo := range funcInfos {
		if profile, ok := fnProfile[fnInfo.Path]; ok {
			entry := &FuncProfileEntry{FuncInfo: fnInfo}
			linkProfileBlocksToFunc(entry, profile)
			fmt.Println(prettySprintFuncInfo(entry.FuncInfo))
			fmt.Println("blocks:")
			for _, block := range entry.ProfileBlocks {
				fmt.Printf("\t%+v\n", block)
			}
		}
	}

	fmt.Println("\nunlinked blocks:")
	for _, profile := range fnProfile {
		for _, block := range profile.Blocks {
			if !block.isLink {
				fmt.Printf("block: %+v\n", block)
			}
		}
	}
}

func TestMergeProfiles(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	dstPath := filepath.Join(testRootDir, "src2/main.go")

	// step1. parse profile
	covPath := filepath.Join(filepath.Dir(srcPath), "profile.cov")
	srcFnProfile, err := parseCovFile(covPath)
	if err != nil {
		t.Fatal(err)
	}
	covPath = filepath.Join(filepath.Dir(dstPath), "profile.cov")
	dstFnProfile, err := parseCovFile(covPath)
	if err != nil {
		t.Fatal(err)
	}

	_, srcProfileMode, err := getProfileNameAndMode(srcFnProfile)
	if err != nil {
		t.Fatal(err)
	}
	dstProfileName, dstProfileMode, err := getProfileNameAndMode(dstFnProfile)
	if err != nil {
		t.Fatal(err)
	}
	if srcProfileMode != dstProfileMode {
		t.Fatal(fmt.Errorf("profile mode is not inconsistent"))
	}

	// step2. diff func
	diffEntries, err := funcDiffForGoFiles(srcPath, dstPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\nfunc diff:")
	for _, entry := range diffEntries {
		fmt.Println(prettySprintDiffEntry(entry))
	}

	// step3. link profile blocks to func
	for _, entry := range diffEntries {
		if entry.Result == diffTypeAdd {
			fpath := entry.DstFuncProfileEntry.FuncInfo.Path
			profile := dstFnProfile[fpath]
			linkProfileBlocksToFunc(entry.DstFuncProfileEntry, profile)
		} else if entry.Result == diffTypeRemove {
			fpath := entry.SrcFuncProfileEntry.FuncInfo.Path
			profile := srcFnProfile[fpath]
			linkProfileBlocksToFunc(entry.SrcFuncProfileEntry, profile)
		} else if entry.Result == diffTypeSame || entry.Result == diffTypeChange {
			// src
			fpath := entry.SrcFuncProfileEntry.FuncInfo.Path
			srcProfile := srcFnProfile[fpath]
			linkProfileBlocksToFunc(entry.SrcFuncProfileEntry, srcProfile)
			// dst
			fpath = entry.DstFuncProfileEntry.FuncInfo.Path
			dstProfile := dstFnProfile[fpath]
			linkProfileBlocksToFunc(entry.DstFuncProfileEntry, dstProfile)
		} else {
			t.Fatal(fmt.Errorf("invalid entry diff result"))
		}
	}

	fmt.Println("\ndiff entries:")
	for _, entry := range diffEntries {
		fmt.Println(prettySprintDiffEntry(entry))
	}

	fmt.Println("\nunlinked blocks:")
	unLinkedBlocks := make([]ProfileBlock, 0, 8)
	for _, profile := range dstFnProfile {
		for _, block := range profile.Blocks {
			if !block.isLink {
				fmt.Printf("\t%+v\n", block)
				unLinkedBlocks = append(unLinkedBlocks, block)
			}
		}
	}

	// step4. merge profiles, and write cov file
	mergedBlocks, err := mergeProfileForDiffEntries(diffEntries)
	if err != nil {
		t.Fatal(err)
	}
	mergedBlocks = append(mergedBlocks, unLinkedBlocks...)
	sort.Sort(blocksByStartPos(mergedBlocks))

	profile := &Profile{
		FileName: dstProfileName,
		Mode:     dstProfileMode,
		Blocks:   mergedBlocks,
	}
	fmt.Println("write profile count:", len(profile.Blocks))

	mergedPath := filepath.Join(filepath.Dir(dstPath), "profile_merged.cov")
	if err := writeCovFile(mergedPath, []*Profile{profile}); err != nil {
		t.Fatal(err)
	}
}
