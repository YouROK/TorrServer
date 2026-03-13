package tgbot

func tr(userID int64, key string) string {
	lang := getUserLang(userID)
	if lang == LangEN {
		if s, ok := msgEN[key]; ok {
			return s
		}
	}
	if s, ok := msgRU[key]; ok {
		return s
	}
	return key
}
