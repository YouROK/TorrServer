package tgbot

func tr(userID int64, key string) string {
	return trLang(getUserLang(userID), key)
}

func trLang(lang, key string) string {
	if lang == LangEN {
		if s, ok := msgEN[key]; ok {
			return s
		}
	}
	if s, ok := msgRU[key]; ok {
		return s
	}
	if s, ok := msgEN[key]; ok {
		return s
	}
	return key
}
