package pkg

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

/* Read profile from .cov file. */

const (
	profileModePrefix = "mode: "
	// profile mode
	profileModeSet   = "set"
	profileModeCount = "count"
)

// ProfileBlock .
type ProfileBlock struct {
	StartLine, StartCol int
	EndLine, EndCol     int
	NumStmt, Count      int
	isLink              bool
}

// Profile .
type Profile struct {
	FileName string
	Mode     string
	Blocks   []ProfileBlock
}

type blocksByStartPos []ProfileBlock

func (b blocksByStartPos) Len() int      { return len(b) }
func (b blocksByStartPos) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b blocksByStartPos) Less(i, j int) bool {
	bi, bj := b[i], b[j]
	return bi.StartLine < bj.StartLine || bi.StartLine == bj.StartLine && bi.StartCol < bj.StartCol
}

// parseCovFile parses cov file and returns a profile for each .go source file.
func parseCovFile(fpath string) (map[string]*Profile, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mode := ""
	RetFnProfile := make(map[string]*Profile, 16)
	buf := bufio.NewReader(f)
	s := bufio.NewScanner(buf)
	for s.Scan() {
		line := strings.Trim(s.Text(), "\n")
		if len(line) == 0 {
			continue
		}
		if len(mode) == 0 {
			if mode, err = parseModeLine(line); err != nil {
				return nil, err
			}
			continue
		}
		filename, block, err := parseProfileLine(line)
		if err != nil {
			return nil, fmt.Errorf("line [%s] doesn't match expected format: %v", line, err)
		}
		profile, ok := RetFnProfile[filename]
		if !ok {
			profile = &Profile{
				FileName: filename,
				Mode:     mode,
			}
			RetFnProfile[filename] = profile
		}
		profile.Blocks = append(profile.Blocks, block)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	for _, profile := range RetFnProfile {
		mergedBlocks, err := mergeBlocksByStartPos(profile.Mode, profile.Blocks)
		if err != nil {
			return nil, err
		}
		profile.Blocks = mergedBlocks
	}

	return RetFnProfile, nil
}

func parseModeLine(line string) (string, error) {
	if !strings.HasPrefix(line, profileModePrefix) || line == profileModePrefix {
		return "", fmt.Errorf("invalid mode: %v", line)
	}
	return line[len(profileModePrefix):], nil
}

// line format: name.go:line.column,line.column numberOfStatements count
func parseProfileLine(line string) (string, ProfileBlock, error) {
	block := ProfileBlock{}
	fields := strings.Split(line, ":")
	fileName := fields[0]
	covInfo := fields[1]

	fields = strings.Split(covInfo, " ")
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return "", block, fmt.Errorf("parse [Count] error: %v", err)
	}
	block.Count = count

	numOfState, err := strconv.Atoi(fields[1])
	if err != nil {
		return "", block, fmt.Errorf("parse [NumberOfStatements] error: %v", err)
	}
	block.NumStmt = numOfState

	position := fields[0]
	fields = strings.Split(position, ",")
	startPos := fields[0]
	startLine, startCol, err := parsePosition(startPos, "Start")
	if err != nil {
		return "", block, err
	}
	block.StartLine = startLine
	block.StartCol = startCol

	endPos := fields[1]
	endLine, endCol, err := parsePosition(endPos, "End")
	if err != nil {
		return "", block, err
	}
	block.EndLine = endLine
	block.EndCol = endCol

	return fileName, block, nil
}

func parsePosition(position, key string) (int, int, error) {
	fields := strings.Split(position, ".")
	line, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse [%sLine] error: %v", key, err)
	}

	col, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse [%sCol] error: %v", key, err)
	}
	return line, col, nil
}

func mergeBlocksByStartPos(mode string, blocks []ProfileBlock) ([]ProfileBlock, error) {
	sort.Sort(blocksByStartPos(blocks))
	j := 1
	for i := 1; i < len(blocks); i++ {
		cur := blocks[i]
		prev := blocks[j-1]
		if cur.StartLine == prev.StartLine &&
			cur.StartCol == prev.StartCol &&
			cur.EndLine == prev.EndLine &&
			cur.EndCol == prev.EndCol {
			if cur.NumStmt != prev.NumStmt {
				return nil, fmt.Errorf("inconsistent NumStmt for block: %+v", cur)
			}
			if mode == profileModeSet {
				blocks[j-1].Count |= cur.Count
			} else {
				blocks[j-1].Count += cur.Count
			}
			continue
		}
		blocks[j] = cur
		j++
	}

	return blocks[:j], nil
}

func getProfileMode(fnProfile map[string]*Profile) (string, error) {
	if len(fnProfile) == 0 {
		return "", fmt.Errorf("profile is empty")
	}

	for _, profile := range fnProfile {
		return profile.Mode, nil
	}
	return "", fmt.Errorf("no profile mode found")
}

/* Write profile to .cov file. */

func writeProfilesToCovFile(filePath string, profiles []*Profile) error {
	outLines := make([]string, 0, 16)
	header := profileModePrefix + profiles[0].Mode
	outLines = append(outLines, header)

	for _, profile := range profiles {
		lines := buildFileProfileLines(profile)
		outLines = append(outLines, lines...)
	}

	content := strings.Join(outLines, "\n")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func buildFileProfileLines(profile *Profile) []string {
	retLines := make([]string, 0, len(profile.Blocks))
	for _, block := range profile.Blocks {
		line := buildProfileLine(profile.FileName, block)
		retLines = append(retLines, line)
	}
	return retLines
}

func buildProfileLine(fpath string, block ProfileBlock) string {
	return fmt.Sprintf("%s:%d.%d,%d.%d %d %d",
		fpath, block.StartLine, block.StartCol, block.EndLine, block.EndCol, block.NumStmt, block.Count)
}

/* Func Cov Struct */

func linkProfileBlocksToFunc(entry *FuncProfileEntry, profile *Profile) {
	linkedBlocks := make([]ProfileBlock, 0, 16)
	funcInfo := entry.FuncInfo
	for i := 0; i < len(profile.Blocks); i++ {
		block := profile.Blocks[i]
		if (block.StartLine > funcInfo.EndLine) ||
			(block.StartLine == funcInfo.EndLine && block.StartCol > funcInfo.EndCol) {
			continue
		}
		if (block.StartLine > funcInfo.StartLine) ||
			((block.StartLine == funcInfo.StartLine) && (block.StartCol > funcInfo.StartCol)) {
			profile.Blocks[i].isLink = true
			linkedBlocks = append(linkedBlocks, profile.Blocks[i])
		}
	}
	entry.ProfileBlocks = linkedBlocks
}

/* Merge Func Cov Entries */

// mergeProfiles: 1.diff func; 2.link profile blocks to func; 3.merge profiles
func mergeProfiles(srcPath, dstPath string, srcFnProfile, dstFnProfile map[string]*Profile) (string, []ProfileBlock, error) {
	// 1.diff func
	diffEntries, err := funcDiffForGoFiles(srcPath, dstPath)
	if err != nil {
		return "", nil, err
	}

	dstFilePath := ""
	for _, entry := range diffEntries {
		if entry.DstFuncProfileEntry != nil {
			dstFilePath = entry.DstFuncProfileEntry.FuncInfo.Path
			break
		}
	}

	// 2.link profile blocks to func
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
			return "", nil, fmt.Errorf("invalid entry diff result")
		}
	}

	log.Println("Diff entries:")
	for _, entry := range diffEntries {
		fmt.Println(prettySprintDiffEntry(entry))
	}

	unLinkedBlocks := make([]ProfileBlock, 0, 8)
	for _, block := range dstFnProfile[dstFilePath].Blocks {
		if !block.isLink {
			unLinkedBlocks = append(unLinkedBlocks, block)
		}
	}

	// 3.merge profiles
	mergedBlocks, err := mergeProfileForDiffEntries(diffEntries)
	if err != nil {
		return "", nil, err
	}
	mergedBlocks = append(mergedBlocks, unLinkedBlocks...)
	sort.Sort(blocksByStartPos(mergedBlocks))

	return dstFilePath, mergedBlocks, nil
}

func mergeProfileForDiffEntries(diffEntries []*DiffEntry) ([]ProfileBlock, error) {
	retBlocks := make([]ProfileBlock, 0, 16)
	for _, diffEntry := range diffEntries {
		if diffEntry.Result == diffTypeAdd || diffEntry.Result == diffTypeChange {
			retBlocks = append(retBlocks, diffEntry.DstFuncProfileEntry.ProfileBlocks...)
		} else if diffEntry.Result == diffTypeSame {
			srcBlocks := diffEntry.SrcFuncProfileEntry.ProfileBlocks
			dstBlocks := diffEntry.DstFuncProfileEntry.ProfileBlocks
			blocks, err := mergeProfileBlocks(srcBlocks, dstBlocks)
			if err != nil {
				return nil, err
			}
			retBlocks = append(retBlocks, blocks...)
		}
	}

	return retBlocks, nil
}

func mergeProfileBlocks(srcBlocks, dstBlocks []ProfileBlock) ([]ProfileBlock, error) {
	if len(srcBlocks) != len(dstBlocks) {
		return nil, fmt.Errorf("src and dst blocks numbers are not equal")
	}

	mergedBlocks := make([]ProfileBlock, 0, len(srcBlocks))
	for i := 0; i < len(srcBlocks); i++ {
		block := dstBlocks[i]
		if srcBlocks[i].Count > block.Count {
			block.Count = srcBlocks[i].Count
		}
		mergedBlocks = append(mergedBlocks, block)
	}
	return mergedBlocks, nil
}
