package dlna

import (
	"strings"
	"sync"
	"testing"

	"github.com/anacrolix/dms/upnpav"

	"server/settings"
	"server/torr"
)

func TestIsHashPath(t *testing.T) {
	hash := strings.Repeat("ab", 20)
	cases := []struct {
		path string
		want bool
	}{
		{hash, true},
		{strings.ToUpper(hash), true},
		{"/" + hash, true},
		{hash + "a", false},
		{hash[:39], false},
		{"", false},
		{strings.Repeat("g", 40), false},
		{strings.Repeat(" ", 40), false},
		{"/TR/" + hash + "/LD", false},
	}
	for _, tc := range cases {
		if got := isHashPath(tc.path); got != tc.want {
			t.Errorf("isHashPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestNormalizeCategoryAndTitle(t *testing.T) {
	cases := map[string][2]string{
		"":         {"uncategorized", "Uncategorized"},
		"  Movie ": {"movie", "Movies"},
		"TV":       {"tv", "TV Shows"},
		"music":    {"music", "Music"},
		"other":    {"other", "Other"},
		"Anime":    {"anime", "anime"},
	}
	for in, want := range cases {
		got := normalizeCategory(in)
		if got != want[0] || categoryTitle(got) != want[1] {
			t.Errorf("%q -> %q / %q, want %q / %q", in, got, categoryTitle(got), want[0], want[1])
		}
	}
}

func TestBrowseEmptyLibrary(t *testing.T) {
	torr.InitApiHelper(torr.NewBTS())
	old := settings.BTsets
	settings.BTsets = &settings.BTSets{FriendlyName: "Box"}
	t.Cleanup(func() { settings.BTsets = old })

	root := getRoot()
	if len(root) != 1 {
		t.Fatalf("root = %#v", root)
	}
	cnt, ok := root[0].(upnpav.Container)
	if !ok || cnt.Title != "Torrents" || cnt.ChildCount != 0 {
		t.Fatalf("root container = %#v", root[0])
	}

	cats := getTorrentCategories()
	if len(cats) != 1 {
		t.Fatalf("categories = %#v", cats)
	}
	empty, ok := cats[0].(upnpav.Container)
	if !ok || empty.Title != "No Torrents" {
		t.Fatalf("empty category = %#v", cats[0])
	}

	items, err := onBrowse("/", "", "127.0.0.1:1900", "test")
	if err != nil || len(items) != 1 {
		t.Fatalf("browse / = %v %v", items, err)
	}
	if _, err := onBrowseMeta(strings.Repeat("ab", 20), "", "127.0.0.1", "test"); err == nil {
		t.Fatal("missing hash meta should fail")
	}
	meta, err := onBrowseMeta("/", "", "127.0.0.1", "test")
	if err != nil || meta == nil {
		t.Fatal(err)
	}
	if name := getDefaultFriendlyName(); name != "Box" {
		t.Fatalf("friendly = %q", name)
	}
	byCat := getTorrentsByCategory("/TR/movie")
	if len(byCat) != 1 {
		t.Fatalf("by category = %#v", byCat)
	}
}

func TestBrowseConcurrent(t *testing.T) {
	torr.InitApiHelper(torr.NewBTS())
	old := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() {
		Stop()
		settings.BTsets = old
	})

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = onBrowse("/", "", "127.0.0.1", "ua")
			_, _ = onBrowse("/TR", "", "127.0.0.1", "ua")
			_, _ = onBrowseMeta("/TR", "", "127.0.0.1", "ua")
			Stop()
		}()
	}
	wg.Wait()
}
