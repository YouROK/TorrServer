package library

import "testing"

func TestNormalizeCategoryAliases(t *testing.T) {
	cases := map[string]string{
		"movie": "movie", " Movies ": "movie", "FILM": "movie", "films": "movie",
		"tv": "tv", "Series": "tv", "serial": "tv", "shows": "tv",
		"music": "music", "OTHER": "other",
		"uncategorized": "uncategorized", "none": "uncategorized", "clear": "uncategorized", "-": "uncategorized",
		"all": "all", "*": "all",
		"": "", "anime": "", "movie tv": "",
	}
	for in, want := range cases {
		if got := NormalizeCategory(in); got != want {
			t.Errorf("NormalizeCategory(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStandardCategoriesAreFixedPoints(t *testing.T) {
	for _, c := range StandardCategories {
		if !IsStandardCategory(c) || NormalizeCategory(c) != c || CategoryValue(c) != c {
			t.Errorf("%q is not a fixed point of the category helpers", c)
		}
	}
	// "all" and "uncategorized" are filters, never a stored category.
	for _, s := range []string{"all", "*", "uncategorized", "none", "", "bogus"} {
		if IsStandardCategory(s) {
			t.Errorf("IsStandardCategory(%q) = true", s)
		}
		if got := CategoryValue(NormalizeCategory(s)); got != "" {
			t.Errorf("CategoryValue(Normalize(%q)) = %q, want empty", s, got)
		}
	}
	// Aliases resolve to what gets stored.
	if got := CategoryValue(NormalizeCategory("Series")); got != "tv" {
		t.Errorf("Series stored as %q", got)
	}
}

func TestMatchCategory(t *testing.T) {
	cases := []struct {
		torrent, filter string
		want            bool
	}{
		{"movie", "", true},
		{"", "all", true},
		{"tv", " ALL ", true},
		{"", "uncategorized", true},
		{"  ", "uncategorized", true},
		{"movie", "uncategorized", false},
		{"Movie", "movie", true},
		{" movie ", "MOVIE", true},
		{"tv", "movie", false},
		{"", "movie", false},
	}
	for _, c := range cases {
		if got := MatchCategory(c.torrent, c.filter); got != c.want {
			t.Errorf("MatchCategory(%q,%q) = %v, want %v", c.torrent, c.filter, got, c.want)
		}
	}
}
