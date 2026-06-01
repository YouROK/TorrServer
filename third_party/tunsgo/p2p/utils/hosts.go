package utils

import "strings"

func MatchHost(patterns []string, target string) bool {
	target = strings.ToLower(target)
	for _, pattern := range patterns {
		pattern = strings.ToLower(pattern)

		if pattern == target || pattern == "*" {
			return true
		}

		if strings.HasPrefix(pattern, "*") {
			suffix := pattern[1:]
			if strings.HasPrefix(pattern, ".") {
				suffix = suffix[1:]
			}
			if strings.HasSuffix(target, suffix) {
				return true
			}
		}
	}
	return false
}
