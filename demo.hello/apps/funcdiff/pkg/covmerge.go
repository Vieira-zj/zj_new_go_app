package pkg

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

/* Read profile from .cov file. */

const profileModePrefix = "mode: "

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
			if mode == "set" {
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

/* Write profile to .cov file. */

func writeCovFile(filePath string, profiles []*Profile) error {
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
			linkedBlocks = append(linkedBlocks, block)
		}
	}
	entry.ProfileBlocks = linkedBlocks
}

/* Merge Func Cov Entries */

// mergeProfiles: 1.diff func; 2.link profile blocks to func; 3.merge profiles, and write cov file
func mergeProfiles(entry *FuncProfileEntry, covPath string) error {
	// TODO:
	return nil
}

func mergeProfileForDiffEntries(diffEntries []*DiffEntry) ([]ProfileBlock, error) {
	retBlocks := make([]ProfileBlock, 0, 16)
	for _, diffEntry := range diffEntries {
		if diffEntry.Result == diffTypeAdd || diffEntry.Result == diffTypeChange {
			retBlocks = append(retBlocks, diffEntry.DstFuncProfileEntry.ProfileBlocks...)
		} else if diffEntry.Result == diffTypeSame {
			blocks, err := mergeProfileBlocks(
				diffEntry.SrcFuncProfileEntry.ProfileBlocks, diffEntry.DstFuncProfileEntry.ProfileBlocks)
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

	retBlocks := make([]ProfileBlock, 0, len(srcBlocks))
	for i := 0; i < len(srcBlocks); i++ {
		block := dstBlocks[i]
		if srcBlocks[i].Count > dstBlocks[i].Count {
			block.Count = srcBlocks[i].Count
		}
		retBlocks = append(retBlocks, block)
	}
	return retBlocks, nil
}
