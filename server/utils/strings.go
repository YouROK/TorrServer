package utils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

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

func CommonPrefix(first, second string) string {
	var result strings.Builder

	minLength := len(first)
	if len(second) < minLength {
		minLength = len(second)
	}

	for i := 0; i < minLength; i++ {
		if first[i] != second[i] {
			break
		}
		result.WriteByte(first[i])
	}

	return result.String()
}

func NumberPrefix(str string) (int, error) {
	var result strings.Builder

	for i := 0; i < len(str); i++ {
		if !unicode.IsDigit(rune(str[i])) {
			break
		}
		result.WriteByte(str[i])
	}

	return strconv.Atoi(result.String())
}

func CompareStrings(first, second string) bool {
	prefixLen := len(CommonPrefix(first, second))
	// start at the first digit of a number the common prefix ends in ("1" of "1" and "10")
	for prefixLen > 0 && unicode.IsDigit(rune(first[prefixLen-1])) {
		prefixLen--
	}
	resultStr1 := first[prefixLen:]
	resultStr2 := second[prefixLen:]
	num1, err1 := NumberPrefix(resultStr1)
	num2, err2 := NumberPrefix(resultStr2)

	if err1 == nil && err2 == nil {
		if num1 == num2 {
			return resultStr1 < resultStr2
		}
		return num1 < num2
	}
	if err1 == nil {
		return true
	} else if err2 == nil {
		return false
	}
	return resultStr1 < resultStr2
}
