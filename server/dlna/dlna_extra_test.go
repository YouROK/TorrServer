package dlna

import (
	"errors"
	"os/user"
	"strings"
	"testing"

	"github.com/anacrolix/dms/upnpav"
	"github.com/anacrolix/torrent"

	"server/settings"
	"server/torr"
	"server/torr/offlinetest"
	"server/torr/state"
)

func TestGetDefaultFriendlyName(t *testing.T) {
	oldSets, oldUser, oldHost := settings.BTsets, currentUser, hostname
	t.Cleanup(func() { settings.BTsets, currentUser, hostname = oldSets, oldUser, oldHost })
	settings.BTsets = &settings.BTSets{}

	set := func(name string, userErr error, host string, hostErr error) {
		currentUser = func() (*user.User, error) {
			if userErr != nil {
				return nil, userErr
			}
			return &user.User{Name: name}, nil
		}
		hostname = func() (string, error) { return host, hostErr }
	}
	fail := errors.New("boom")

	tests := []struct {
		name    string
		user    string
		userErr error
		host    string
		hostErr error
		want    string
	}{
		{"neither known", "", fail, "", fail, "TorrServer"},
		{"user equals host", "box", nil, "box", nil, "TorrServer: box"},
		{"user on host", "pavel", nil, "mac", nil, "TorrServer: pavel on mac"},
		{"host lookup fails", "pavel", nil, "", fail, "TorrServer: pavel@"},
		{"user lookup fails", "", fail, "mac", nil, "TorrServer: @mac"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set(tt.user, tt.userErr, tt.host, tt.hostErr)
			if got := getDefaultFriendlyName(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("localhost resolves to an address or falls back", func(t *testing.T) {
		set("", fail, "localhost", nil)
		got := getDefaultFriendlyName()
		if !strings.HasPrefix(got, "TorrServer ") && got != "TorrServer: @localhost" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("configured name wins", func(t *testing.T) {
		settings.BTsets = &settings.BTSets{FriendlyName: "Living room"}
		set("pavel", nil, "mac", nil)
		if got := getDefaultFriendlyName(); got != "Living room" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestLoadTorrentAndItems(t *testing.T) {
	offlinetest.Prepare(t)
	bt := torr.NewBTS()
	torr.InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)

	show := offlinetest.WriteShow(t)
	seed := offlinetest.Seed(t, show)
	tor, err := torr.AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil || !tor.GotInfo() {
		t.Fatal(err)
	}
	tor.AddPeers([]torrent.Peer{seed.Peer()})
	hash := tor.Hash().HexString()

	// Not a hash, and a well-formed hash nobody added: nothing to list.
	if got := loadTorrent("/not-a-hash", "127.0.0.1:1900"); got != nil {
		t.Fatalf("bad hash: %v", got)
	}
	if got := loadTorrent("/"+strings.Repeat("ab", 20), "127.0.0.1:1900"); got != nil {
		t.Fatalf("unknown torrent: %v", got)
	}

	items := loadTorrent("/"+hash, "192.168.1.10:1900")
	if len(items) == 0 {
		t.Fatal("no media items for the show")
	}
	var titles []string
	for _, it := range items {
		item, ok := it.(upnpav.Item)
		if !ok {
			t.Fatalf("item type %T", it)
		}
		titles = append(titles, item.Title)
		if len(item.Res) != 1 || !strings.Contains(item.Res[0].URL, "/play/"+hash+"/") ||
			!strings.HasPrefix(item.Res[0].URL, "http://192.168.1.10:") || item.Res[0].Size == 0 {
			t.Fatalf("resource %+v", item.Res)
		}
		if item.ParentID != "%2F"+hash || !strings.HasPrefix(item.Class, "object.item.") {
			t.Fatalf("object %+v", item.Object)
		}
	}
	joined := strings.Join(titles, "|")
	if !strings.Contains(joined, "episode1.mkv") || strings.Contains(joined, ".srt") {
		t.Fatalf("titles %q: want episode1.mkv and no subtitles", joined)
	}

	// The same path through the hash directory also resolves.
	if again := loadTorrent("/"+hash+"/Season 1", "192.168.1.10:1900"); len(again) != 0 && len(again) != len(items) {
		t.Fatalf("subpath items = %d, want 0 or %d", len(again), len(items))
	}

	// A file with no known MIME type is skipped.
	if got := getObjFromTorrent("/"+hash, "%2F"+hash, "h:1", tor, &state.TorrentFileStat{Id: 9, Path: "noext"}); got != nil {
		t.Fatalf("unknown mime: %v", got)
	}
}
