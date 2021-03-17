package cover

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Profile represents the profiling data for a specific go source file in ".cov".
type Profile struct {
	FileName string
	Mode     string
	Blocks   []ProfileBlock
}

type profilesByFileName []*Profile

// implement sort interface.
func (p profilesByFileName) Len() int {
	return len(p)
}

func (p profilesByFileName) Less(i, j int) bool {
	return p[i].FileName < p[j].FileName
}

func (p profilesByFileName) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// ProfileBlock represents a single func block of profiling data.
type ProfileBlock struct {
	StartLine, StartCol int
	EndLine, EndCol     int
	NumStmt, Count      int
}

type blocksByStart []ProfileBlock

// implement sort interface.
func (b blocksByStart) Len() int {
	return len(b)
}

func (b blocksByStart) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b blocksByStart) Less(i, j int) bool {
	bi, bj := b[i], b[j]
	return bi.StartLine < bj.StartLine || bi.StartLine == bj.StartLine && bi.StartCol < bj.StartCol
}

// parseProfiles parses profile data in specified .cov file and returns a Profile for each source file.
// Example:
// demo.hello/echoserver/handlers/ping.go:13.41,15.2 1 2
// demo.hello/echoserver/handlers/ping.go:18.40,21.2 2 1
func parseProfiles(fileName string) ([]*Profile, error) {
	covFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	fileProfiles := make(map[string]*Profile)
	mode := ""

	buf := bufio.NewReader(covFile)
	s := bufio.NewScanner(buf)
	for s.Scan() {
		line := s.Text()
		if len(mode) == 0 {
			const p = "mode: "
			if !strings.HasPrefix(line, p) {
				return nil, fmt.Errorf("bad mode line: %v", line)
			}
			mode = line[len(p):]
			continue
		}

		fn, block, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %q doesn't match expected format: %v", line, err)
		}
		p, ok := fileProfiles[fn]
		if !ok {
			p = &Profile{
				FileName: fn,
				Mode:     mode,
			}
			fileProfiles[fn] = p
		}
		p.Blocks = append(p.Blocks, block)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	for _, p := range fileProfiles {
		sort.Sort(blocksByStart(p.Blocks))
	}

	retProfiles := make([]*Profile, 0, len(fileProfiles))
	for _, profile := range fileProfiles {
		retProfiles = append(retProfiles, profile)
	}
	sort.Sort(profilesByFileName(retProfiles))
	return retProfiles, nil
}

// parseLine parses a line from a coverage file.
// Columns: StartLine,StartCol,EndLine,EndCol,NumStmt,Count
// demo.hello/echoserver/handlers/hooks.go:12.65,13.36 1 0
func parseLine(l string) (fileName string, block ProfileBlock, err error) {
	end := len(l)
	b := ProfileBlock{}

	b.Count, end, err = seekBack(l, ' ', end, "StartLine")
	if err != nil {
		return "", b, err
	}
	b.NumStmt, end, err = seekBack(l, ' ', end, "NumStmt")
	if err != nil {
		return "", b, err
	}
	b.EndCol, end, err = seekBack(l, '.', end, "EndCol")
	if err != nil {
		return "", b, err
	}
	b.EndLine, end, err = seekBack(l, ',', end, "EndLine")
	if err != nil {
		return "", b, err
	}
	b.StartCol, end, err = seekBack(l, '.', end, "StartCol")
	if err != nil {
		return "", b, err
	}
	b.StartLine, end, err = seekBack(l, ':', end, "StartLine")
	if err != nil {
		return "", b, err
	}

	funcName := l[:end]
	if funcName == "" {
		return "", b, errors.New("a FileName cannot be blank")
	}
	return funcName, b, nil
}

// seekBack searches backwards from end to find sep in l, then returns the
// value between sep and end as an integer.
// example: seekBack(l, ' ', end, "Count")
func seekBack(l string, sep byte, end int, what string) (value int, nextSep int, err error) {
	for idx := end - 1; idx >= 0; idx-- {
		if l[idx] == sep {
			i, err := strconv.Atoi(l[idx+1 : end])
			if err != nil {
				return 0, 0, fmt.Errorf("couldn't parse %q: %v", what, err)
			}
			return i, idx, nil
		}
	}
	return 0, 0, fmt.Errorf("couldn't find a %s before %s", string(sep), what)
}
