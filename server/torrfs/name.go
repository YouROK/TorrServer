package torrfs

import (
	"strings"
)

// SanitizeName makes a torrent/category name usable as a filesystem node name.
func SanitizeName(name string) string {
	name = strings.Map(func(r rune) rune {
		switch {
		case r == '/' || r == '\\':
			return '_'
		case r < 0x20 || r == 0x7f:
			return -1
		default:
			return r
		}
	}, name)

	name = strings.TrimSpace(name)

	// "." and ".." are collapsed by path.Clean and break path resolution
	if name == "" || name == "." || name == ".." {
		return ""
	}
	return name
}
