package torrshash

import (
	"bytes"
	"math/big"
	"regexp"
	"strings"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var base62Regex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func IsBase62(s string) bool {
	return base62Regex.MatchString(strings.TrimSpace(s))
}

func Encode62(b []byte) string {
	x := new(big.Int).SetBytes(b)
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var res []byte
	for x.Cmp(zero) > 0 {
		x.QuoRem(x, base, mod)
		res = append(res, base62Alphabet[mod.Int64()])
	}
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}
	return string(res)
}

func Decode62(s string) []byte {
	res := new(big.Int)
	base := big.NewInt(62)
	for _, char := range s {
		val := bytes.IndexByte([]byte(base62Alphabet), byte(char))
		res.Mul(res, base)
		res.Add(res, big.NewInt(int64(val)))
	}
	return res.Bytes()
}
