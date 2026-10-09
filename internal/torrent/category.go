package torrent

import "strings"

// Категории раздач. Ключ хранится в базе, подпись используется в путях и интерфейсе.
const (
	CategoryMovie = "movie"
	CategoryTV    = "tv"
	CategoryMusic = "music"
	CategoryOther = "other"
)

// CategoryKey приводит значение категории к известному ключу.
// Пустая строка означает отсутствие категории, неизвестное значение остается как есть.
func CategoryKey(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}

	switch strings.ToLower(v) {
	case "movie", "movies", "film", "films", "фильм", "фильмы":
		return CategoryMovie
	case "tv", "series", "serial", "сериал", "сериалы":
		return CategoryTV
	case "music", "музыка":
		return CategoryMusic
	case "other", "другое":
		return CategoryOther
	}
	return v
}

// CategoryLabel возвращает читаемое имя категории для путей и подписей
func CategoryLabel(key string) string {
	switch key {
	case CategoryMovie:
		return "Movies"
	case CategoryTV:
		return "Series"
	case CategoryMusic:
		return "Music"
	case CategoryOther:
		return "Other"
	}
	return key
}
