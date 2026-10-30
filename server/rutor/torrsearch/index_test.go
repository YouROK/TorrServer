package torrsearch

import (
	"reflect"
	"testing"

	"server/rutor/models"
)

func docs(titles ...string) []*models.TorrentDetails {
	out := make([]*models.TorrentDetails, len(titles))
	for i, t := range titles {
		out[i] = &models.TorrentDetails{Title: t}
	}
	return out
}

func TestAnalyze(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"The Lord of the Rings: 2001", []string{"the", "lord", "the", "rings", "2001"}},
		{"Ёжик и Медвежонок", []string{"ежик", "медвежонок"}}, // ё→е, stop word "и"
		{"a an and of", []string{}},
		{"Spider-Man", []string{"spider", "man"}},
		{"", []string{}},
	}
	for _, c := range cases {
		got := analyze(c.in)
		if len(got) == 0 && len(c.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("analyze(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIntersection(t *testing.T) {
	cases := []struct{ a, b, want []int }{
		{[]int{1, 2, 3}, []int{2, 3, 4}, []int{2, 3}},
		{[]int{1, 3, 5}, []int{2, 4, 6}, []int{}},
		{nil, []int{1}, []int{}},
		{[]int{7}, []int{7}, []int{7}},
	}
	for _, c := range cases {
		if got := intersection(c.a, c.b); !reflect.DeepEqual(got, c.want) {
			t.Errorf("intersection(%v,%v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestIndexSearch(t *testing.T) {
	NewIndex(docs(
		"Матрица / The Matrix (1999) BDRip",           // 0
		"Матрица: Перезагрузка / The Matrix Reloaded", // 1
		"Ёлки 2 (2011)",        // 2
		"Matrix Matrix Matrix", // 3: repeated token, one posting
	))

	if got := Search("matrix"); !reflect.DeepEqual(got, []int{0, 1, 3}) {
		t.Errorf("matrix = %v", got)
	}
	// AND semantics, order-insensitive.
	if got := Search("reloaded matrix"); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("reloaded matrix = %v", got)
	}
	// ё in query and title are folded to е.
	if got := Search("елки"); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("елки = %v", got)
	}
	if got := Search("ЁЛКИ"); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("ЁЛКИ = %v", got)
	}
	// One unknown token empties the whole result.
	if got := Search("matrix nonexistent"); got != nil {
		t.Errorf("unknown token = %v, want nil", got)
	}
	// Only stop words: no tokens, no results.
	if got := Search("and of"); got != nil {
		t.Errorf("stop words only = %v, want nil", got)
	}
	if got := len(GetIDX()["matrix"]); got != 3 {
		t.Errorf("matrix postings = %d, want 3 (duplicate token in a title must not repeat the id)", got)
	}
}

// A multi-token query intersects into a new slice and must not rewrite the
// postings of the first token.
func TestIntersectDoesNotMutatePostings(t *testing.T) {
	NewIndex(docs("alpha beta", "alpha", "alpha beta gamma"))
	first := Search("alpha")
	_ = Search("alpha beta")
	if !reflect.DeepEqual(first, []int{0, 1, 2}) || !reflect.DeepEqual(GetIDX()["alpha"], []int{0, 1, 2}) {
		t.Fatalf("postings changed: %v / %v", first, GetIDX()["alpha"])
	}
}

func TestNewIndexNilResets(t *testing.T) {
	NewIndex(docs("something"))
	NewIndex(nil)
	if len(GetIDX()) != 0 || Search("something") != nil {
		t.Fatal("NewIndex(nil) must clear the index")
	}
}
