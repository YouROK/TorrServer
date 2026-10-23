package dlna

import (
	"errors"
	"net"
	"os/user"
	"strings"
	"sync"
	"testing"

	"github.com/anacrolix/dms/upnpav"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr"
	torrState "server/torr/state"
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

func TestStartListenFailureDoesNotExitProcess(t *testing.T) {
	oldListen, oldExit := listenBind, exitFunc
	oldEmb, oldIPs, oldSets := settings.Embedded, settings.IPs, settings.BTsets
	settings.Embedded = false
	settings.IPs = []string{"127.0.0.1"}
	settings.BTsets = &settings.BTSets{}
	code := 0
	exitFunc = func(c int) { code = c }
	listenBind = func([]string, string) (net.Listener, error) {
		return nil, errors.New("bind failed")
	}
	t.Cleanup(func() {
		Stop()
		listenBind, exitFunc = oldListen, oldExit
		settings.Embedded, settings.IPs, settings.BTsets = oldEmb, oldIPs, oldSets
	})
	Start()
	if code != 1 {
		t.Fatalf("exitFunc code = %d", code)
	}
	if dmsServer != nil {
		t.Fatal("server left set after listen failure")
	}
}

func TestStartStopEmbedded(t *testing.T) {
	oldEmb, oldIPs, oldSets := settings.Embedded, settings.IPs, settings.BTsets
	settings.Embedded = true
	settings.IPs = []string{"127.0.0.1"}
	settings.BTsets = &settings.BTSets{FriendlyName: "DLNA Box"}
	t.Cleanup(func() {
		Stop()
		settings.Embedded, settings.IPs, settings.BTsets = oldEmb, oldIPs, oldSets
	})
	Start()
	if dmsServer == nil {
		t.Fatal("expected DLNA server")
	}
	Stop()
	if dmsServer != nil {
		t.Fatal("stop left the server")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); Start() }()
		go func() { defer wg.Done(); Stop() }()
	}
	wg.Wait()
	Stop()
}

func TestBrowseLibraryFromDB(t *testing.T) {
	dir := t.TempDir()
	oldPath, oldSets := settings.Path, settings.BTsets
	settings.Path = dir
	t.Cleanup(func() {
		settings.CloseDB()
		settings.Path, settings.BTsets = oldPath, oldSets
	})
	if err := settings.InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	torr.InitApiHelper(torr.NewBTS())
	hash := strings.Repeat("ab", 20)
	spec := &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex(hash), DisplayName: "Show"}
	settings.AddTorrent(&settings.TorrentDB{
		TorrentSpec: spec,
		Title:       "A/Show",
		Category:    " TV ",
		Poster:      "http://poster",
		Timestamp:   50,
	})
	spec2 := &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex(strings.Repeat("cd", 20)), DisplayName: "M"}
	settings.AddTorrent(&settings.TorrentDB{TorrentSpec: spec2, Title: "Movie", Category: "movie", Timestamp: 10})

	cats := getTorrentCategories()
	if len(cats) < 2 {
		t.Fatalf("categories %#v", cats)
	}
	byCat := getTorrentsByCategory("/TR/tv")
	if len(byCat) != 1 {
		t.Fatalf("tv %#v", byCat)
	}
	cnt := byCat[0].(upnpav.Container)
	if cnt.Title != "A|Show" || cnt.ChildCount != 1 {
		t.Fatalf("container %#v", cnt)
	}
	items, err := onBrowse("/"+hash, "", "127.0.0.1:1900", "ua")
	if err != nil || len(items) != 1 {
		t.Fatalf("hash browse %v %v", items, err)
	}
	meta, err := onBrowseMeta("/"+hash, "", "127.0.0.1", "ua")
	if err != nil || meta == nil {
		t.Fatal(err)
	}
	if _, err := onBrowseMeta("/TR/tv", "", "127.0.0.1", "ua"); err != nil {
		t.Fatal(err)
	}
	if _, err := onBrowseMeta("/TR", "", "127.0.0.1", "ua"); err != nil {
		t.Fatal(err)
	}
	if meta := getTorrentMeta("/"+hash+"/LD", "127.0.0.1"); meta == nil {
		t.Fatal("load meta")
	}
	oldUser, oldHost := currentUser, hostname
	t.Cleanup(func() { currentUser, hostname = oldUser, oldHost })
	settings.BTsets.FriendlyName = ""
	currentUser = func() (*user.User, error) { return &user.User{Name: "box"}, nil }
	hostname = func() (string, error) { return "box", nil }
	if name := getDefaultFriendlyName(); name != "TorrServer: box" {
		t.Fatalf("same name %q", name)
	}
	hostname = func() (string, error) { return "localhost", nil }
	if name := getDefaultFriendlyName(); !strings.Contains(name, "box") {
		t.Fatalf("localhost %q", name)
	}
	currentUser = func() (*user.User, error) { return nil, errors.New("nouser") }
	hostname = func() (string, error) { return "", errors.New("nohost") }
	if name := getDefaultFriendlyName(); name != "TorrServer" {
		t.Fatalf("fallback %q", name)
	}
	hostname = func() (string, error) { return "host", nil }
	if name := getDefaultFriendlyName(); !strings.Contains(name, "host") {
		t.Fatalf("host only %q", name)
	}
	if root, err := onBrowse("/", "", "127.0.0.1", "ua"); err != nil || len(root) == 0 {
		t.Fatalf("root %v %v", root, err)
	}
	if cats, err := onBrowse("/TR", "", "127.0.0.1", "ua"); err != nil || len(cats) == 0 {
		t.Fatalf("cats %v %v", cats, err)
	}
	missing := strings.Repeat("11", 20)
	if _, err := onBrowse("/"+missing+"/LD", "", "127.0.0.1:9080", "ua"); err != nil {
		t.Fatal(err)
	}
	if _, err := onBrowse("/short/LD", "", "127.0.0.1", "ua"); err != nil {
		t.Fatal(err)
	}
	if meta := getTorrentMeta("/"+hash+"/movie.mkv", "127.0.0.1:9080"); meta == nil {
		t.Fatal("file meta")
	}
	settings.Port = "8090"
	settings.BTsets.EnableDebug = true
	obj := getObjFromTorrent("/"+hash, "%2F"+hash, "127.0.0.1:9080", &torr.Torrent{
		TorrentSpec: spec,
		Title:       "Show",
	}, &torrState.TorrentFileStat{Path: "notes.txt", Id: 1})
	if obj != nil {
		t.Fatal("non-media file")
	}
}
