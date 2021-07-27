package dlna

import (
	"path/filepath"
)

func isHashPath(path string) bool {
	base := filepath.Base(path)
	if len(base) == 40 {
		data := []byte(base)
		for _, v := range data {
			if !(v >= 48 && v <= 57 || v >= 65 && v <= 70 || v >= 97 && v <= 102) {
				return false
			}
		}
		return true
	}
	return false
}
