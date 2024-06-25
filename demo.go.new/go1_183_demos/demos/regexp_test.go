package demos

import (
	"regexp"
	"strings"
	"testing"
)

func TestRegExpSubmatch(t *testing.T) {
	pattern := `(\w+)\s(\d+)`
	regex := regexp.MustCompile(pattern)
	input := "John 25, Jane 30, Alice 35"

	t.Run("find submatch", func(t *testing.T) {
		indices := regex.FindStringSubmatchIndex(input)
		t.Log("indices:", indices)

		submatches := make([]string, 0)
		for i := 2; i < len(indices); i += 2 {
			start, end := indices[i], indices[i+1]
			submatches = append(submatches, input[start:end])
		}
		t.Log("sub matches string:", submatches)
	})

	t.Run("find all submatch", func(t *testing.T) {
		indices := regex.FindAllStringSubmatchIndex(input, -1)
		t.Log("indices:", indices)

		submatches := make([]string, 0)
		for _, idx := range indices {
			for i := 2; i < len(idx); i += 2 {
				start, end := idx[i], idx[i+1]
				submatches = append(submatches, input[start:end])
			}
		}
		t.Log("sub matches string:", submatches)
	})
}

func TestRegExpSubmatchAndReplace(t *testing.T) {
	pattern := `(\d{2})-(\d{2})-(\d{4})`
	regex := regexp.MustCompile(pattern)
	input := "Today's date is 25-06-2024, nice"

	indices := regex.FindStringSubmatchIndex(input)
	t.Log("indices:", indices)

	t.Run("get sub matches str", func(t *testing.T) {
		submatches := make([]string, 0, (len(indices)-2)/2)
		for i := 2; i < len(indices); i += 2 {
			start, end := indices[i], indices[i+1]
			submatches = append(submatches, input[start:end])
		}
		t.Log("sub matches string:", submatches)
	})

	t.Run("replace sub matches", func(t *testing.T) {
		sb := strings.Builder{}
		cursor, replaceIdx := 0, 0
		replaces := []string{"day", "month", "year"}

		for i := 2; i < len(indices); i += 2 {
			start, end := indices[i], indices[i+1]
			sb.WriteString(input[cursor:start])
			sb.WriteString(replaces[replaceIdx])
			replaceIdx += 1
			cursor = end
		}
		sb.WriteString(input[cursor:])

		t.Log("replace string:", sb.String())
	})
}
