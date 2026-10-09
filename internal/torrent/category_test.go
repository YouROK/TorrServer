package torrent

import "testing"

// TestCategoryKey проверяет приведение значений категории к известным ключам
func TestCategoryKey(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"   ", ""},
		{"movie", CategoryMovie},
		{"Movies", CategoryMovie},
		{"ФИЛЬМЫ", CategoryMovie},
		{"tv", CategoryTV},
		{"series", CategoryTV},
		{"Сериалы", CategoryTV},
		{"music", CategoryMusic},
		{"other", CategoryOther},
		{"Другое", CategoryOther},
		// Неизвестное значение остается как есть
		{"documentary", "documentary"},
	}

	for _, c := range cases {
		if got := CategoryKey(c.in); got != c.want {
			t.Errorf("CategoryKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestCategoryLabel проверяет читаемые имена категорий
func TestCategoryLabel(t *testing.T) {
	cases := map[string]string{
		CategoryMovie: "Movies",
		CategoryTV:    "Series",
		CategoryMusic: "Music",
		CategoryOther: "Other",
		"custom":      "custom",
	}

	for key, want := range cases {
		if got := CategoryLabel(key); got != want {
			t.Errorf("CategoryLabel(%q) = %q, want %q", key, got, want)
		}
	}
}
