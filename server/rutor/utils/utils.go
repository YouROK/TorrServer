package utils

import "strings"

func ClearStr(str string) string {
	ret := ""
	str = strings.ToLower(str)
	for _, r := range str {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'а' && r <= 'я') || r == 'ё' {
			ret = ret + string(r)
		}
	}
	return ret
}
