package torrsearch

import (
	"strings"

	snowballeng "github.com/kljensen/snowball/english"
	snowballru "github.com/kljensen/snowball/russian"
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

// stemmerFilter returns a slice of stemmed tokens.
func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		worden := snowballeng.Stem(token, false)
		wordru := snowballru.Stem(token, false)
		if wordru == "" || worden == "" {
			continue
		}
		if wordru != token {
			r[i] = wordru
		} else {
			r[i] = worden
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
	case "a", "about", "above", "after", "again", "against", "all", "am", "an",
		"and", "any", "are", "as", "at", "be", "because", "been", "before",
		"being", "below", "between", "both", "but", "by", "can", "did", "do",
		"does", "doing", "don", "down", "during", "each", "few", "for", "from",
		"further", "had", "has", "have", "having", "he", "her", "here", "hers",
		"herself", "him", "himself", "his", "how", "i", "if", "in", "into", "is",
		"it", "its", "itself", "just", "me", "more", "most", "my", "myself",
		"no", "nor", "not", "now", "of", "off", "on", "once", "only", "or",
		"other", "our", "ours", "ourselves", "out", "over", "own", "s", "same",
		"she", "should", "so", "some", "such", "t", "than", "that", "the", "their",
		"theirs", "them", "themselves", "then", "there", "these", "they",
		"this", "those", "through", "to", "too", "under", "until", "up",
		"very", "was", "we", "were", "what", "when", "where", "which", "while",
		"who", "whom", "why", "will", "with", "you", "your", "yours", "yourself",
		"yourselves", "и", "в", "во", "не", "что", "он", "на", "я", "с",
		"со", "как", "а", "то", "все", "она", "так", "его",
		"но", "да", "ты", "к", "у", "же", "вы", "за", "бы",
		"по", "только", "ее", "мне", "было", "вот", "от",
		"меня", "еще", "нет", "о", "из", "ему", "теперь",
		"когда", "даже", "ну", "вдруг", "ли", "если", "уже",
		"или", "ни", "быть", "был", "него", "до", "вас",
		"нибудь", "опять", "уж", "вам", "ведь", "там", "потом",
		"себя", "ничего", "ей", "может", "они", "тут", "где",
		"есть", "надо", "ней", "для", "мы", "тебя", "их",
		"чем", "была", "сам", "чтоб", "без", "будто", "чего",
		"раз", "тоже", "себе", "под", "будет", "ж", "тогда",
		"кто", "этот", "того", "потому", "этого", "какой",
		"совсем", "ним", "здесь", "этом", "один", "почти",
		"мой", "тем", "чтобы", "нее", "сейчас", "были", "куда",
		"зачем", "всех", "никогда", "можно", "при", "наконец",
		"два", "об", "другой", "хоть", "после", "над", "больше",
		"тот", "через", "эти", "нас", "про", "всего", "них",
		"какая", "много", "разве", "три", "эту", "моя",
		"впрочем", "хорошо", "свою", "этой", "перед", "иногда",
		"лучше", "чуть", "том", "нельзя", "такой", "им", "более",
		"всегда", "конечно", "всю", "между":
		return true
	}
	return false
}
