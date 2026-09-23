package torrfs

import (
	"strings"
	"unicode"
)

// sanitizeName заменяет недопустимые символы на '_',
func sanitizeName(s string) string {
	if s == "" {
		return "unnamed"
	}

	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if isBadRune(r) {
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
			continue
		}
		b.WriteRune(r)
		lastUnderscore = r == '_'
	}

	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "unnamed"
	}
	return out
}

func isBadRune(r rune) bool {
	switch r {
	case '/', '\\', '<', '>', '&', '"', ':', '*', '?', '|':
		return true
	}
	return unicode.IsControl(r)
}
