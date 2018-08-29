package utils

import (
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

func CleanFName(file string) string {
	re := regexp.MustCompile(`[ !*'();:@&=+$,/?#\[\]~"]`)
	ret := re.ReplaceAllString(file, `_`)
	ret = strings.Replace(ret, "__", "_", -1)
	return ret
}

func FreeOSMem() {
	debug.FreeOSMemory()
}

func FreeOSMemGC() {
	runtime.GC()
	debug.FreeOSMemory()
}

const (
	_ = 1.0 << (10 * iota) // ignore first value by assigning to blank identifier
	KB
	MB
	GB
	TB
	PB
	EB
)

func Format(b float64) string {
	multiple := ""
	value := b

	switch {
	case b >= EB:
		value /= EB
		multiple = "EB"
	case b >= PB:
		value /= PB
		multiple = "PB"
	case b >= TB:
		value /= TB
		multiple = "TB"
	case b >= GB:
		value /= GB
		multiple = "GB"
	case b >= MB:
		value /= MB
		multiple = "MB"
	case b >= KB:
		value /= KB
		multiple = "KB"
	case b == 0:
		return "0"
	default:
		return strconv.FormatInt(int64(b), 10) + "B"
	}

	return fmt.Sprintf("%.2f%s", value, multiple)
}

//
//func IsCyrillic(str string) bool {
//	for _, r := range str {
//		if unicode.Is(unicode.Cyrillic, r) {
//			return true
//		}
//	}
//	return false
//}
