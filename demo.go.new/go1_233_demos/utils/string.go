package utils

import "strings"

func SplitIgnoreEmpty(s, sep string) []string {
	if len(s) == 0 {
		return nil
	}

	raw := strings.Split(s, sep)
	result := make([]string, 0, len(raw))
	for _, str := range raw {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
