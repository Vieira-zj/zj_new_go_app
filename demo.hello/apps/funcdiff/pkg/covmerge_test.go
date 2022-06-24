package pkg

import (
	"fmt"
	"path/filepath"
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
	profiles, err := parseCovFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	for fname, profile := range profiles {
		fmt.Printf("\nmode=%s, filename=%s\n", profile.Mode, fname)
		for _, block := range profile.Blocks {
			fmt.Printf("\tblock: %+v\n", block)
		}
	}

	// write cov
	outProfiles := make([]*Profile, 0, len(profiles))
	for _, profile := range profiles {
		outProfiles = append(outProfiles, profile)
	}

	outPath := "/tmp/test/out_profile.cov"
	if err := writeProfilesToCovFile(outPath, outProfiles); err != nil {
		t.Fatal(err)
	}
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

func TestMergeProfiles01(t *testing.T) {
	// case: merge profile for a change file
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	dstPath := filepath.Join(testRootDir, "src2/main.go")

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

	profileMode, err := getProfileMode(dstFnProfile)
	if err != nil {
		t.Fatal(err)
	}

	dstFilePath, mergedBlocks, err := mergeProfiles(srcPath, dstPath, srcFnProfile, dstFnProfile)
	if err != nil {
		t.Fatal(err)
	}

	profiles := []*Profile{
		{
			FileName: dstFilePath,
			Mode:     profileMode,
			Blocks:   mergedBlocks,
		},
	}
	mergedPath := filepath.Join(filepath.Dir(dstPath), "profile_merged.cov")
	if err := writeProfilesToCovFile(mergedPath, profiles); err != nil {
		t.Fatal(err)
	}
}

func TestMergeProfiles02(t *testing.T) {
	// case: merge profile for same and change files
	getPath := func(key string) string {
		return filepath.Join(testRootDir, key)
	}

	covPath := filepath.Join(testRootDir, "src1/profile.cov")
	srcFnProfile, err := parseCovFile(covPath)
	if err != nil {
		t.Fatal(err)
	}
	covPath = filepath.Join(testRootDir, "src2/profile.cov")
	dstFnProfile, err := parseCovFile(covPath)
	if err != nil {
		t.Fatal(err)
	}

	dstProfileMode, err := getProfileMode(dstFnProfile)
	if err != nil {
		t.Fatal(err)
	}

	diffFiles := map[string]string{
		getPath("src1/main.go"):     getPath("src2/main.go"),
		getPath("src1/nochange.go"): getPath("src2/nochange.go"),
	}
	profiles := make([]*Profile, 0, len(diffFiles))
	for srcPath, dstPath := range diffFiles {
		dstFilePath, mergedBlocks, err := mergeProfiles(srcPath, dstPath, srcFnProfile, dstFnProfile)
		if err != nil {
			t.Fatal(err)
		}
		profiles = append(profiles, &Profile{
			FileName: dstFilePath,
			Mode:     dstProfileMode,
			Blocks:   mergedBlocks,
		},
		)
	}

	if err := writeProfilesToCovFile(getPath("src2/profile_merged.cov"), profiles); err != nil {
		t.Fatal(err)
	}
}
