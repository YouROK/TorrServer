package torrsearch

import (
	"strings"
)

// lowercaseFilter returns a slice of tokens normalized to lower case.
func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = replaceChars(strings.ToLower(token))
	}
	return r
}

// stopwordFilter returns a slice of tokens with stop words removed.
func stopwordFilter(tokens []string) []string {
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if !isStopWord(token) {
			r = append(r, token)
		}
	}
	return r
}

func replaceChars(word string) string {
	out := []rune(word)
	for i, r := range out {
		if r == 'ё' {
			out[i] = 'е'
		}
	}
	return string(out)
}

func isStopWord(word string) bool {
	switch word {
	case "a", "am", "an", "and", "are", "as", "at", "be",
		"by", "did", "do", "is", "of", "or", "s", "so", "t",
		"и", "в", "с", "со", "а", "но", "к", "у",
		"же", "бы", "по", "от", "о", "из", "ну",
		"ли", "ни", "нибудь", "уж", "ведь", "ж", "об":
		return true
	}
	return false
}
