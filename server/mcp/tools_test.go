package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/anacrolix/torrent"

	"server/rutor/models"
	set "server/settings"
	"server/torr"
	"server/torr/offlinetest"
)

func TestToolsOffline(t *testing.T) {
	offlinetest.Prepare(t)
	bt := torr.NewBTS()
	torr.InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)

	show := offlinetest.WriteShow(t)
	seed := offlinetest.Seed(t, show)
	var meta []byte
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := show.Meta.Write(w); err != nil {
				t.Error(err)
			}
		}))
		defer srv.Close()
		meta = []byte(srv.URL)
	}
	ctx := context.Background()

	if _, _, err := addTorrent(ctx, nil, addTorrentIn{}); err == nil {
		t.Fatal("empty link")
	}
	set.ReadOnly = true
	if _, _, err := addTorrent(ctx, nil, addTorrentIn{Link: "http://127.0.0.1/x.torrent"}); err == nil {
		t.Fatal("readonly add")
	}
	set.ReadOnly = false

	if _, out, err := addTorrent(ctx, nil, addTorrentIn{Link: string(meta), Title: "Show", Category: "tv", Save: boolPtr(true)}); err != nil || out.Hash == "" || len(out.Files) != 4 {
		t.Fatalf("add %+v %v", out, err)
	}
	hash := show.Spec.InfoHash.HexString()
	tor := torr.GetTorrent(hash)
	if tor == nil || !tor.GotInfo() {
		t.Fatal("not loaded")
	}
	tor.AddPeers([]torrent.Peer{seed.Peer()})

	if _, out, err := getServerInfo(ctx, nil, emptyInput{}); err != nil || out.Name == "" {
		t.Fatal(err, out)
	}
	if _, out, err := listTorrents(ctx, nil, listTorrentsIn{Category: "tv", Search: "show"}); err != nil || out.Count < 1 {
		t.Fatal(err, out)
	}
	if _, out, err := getTorrent(ctx, nil, hashIn{Hash: hash}); err != nil || len(out.Files) != 4 {
		t.Fatalf("get %+v %v", out, err)
	}
	if _, _, err := getTorrent(ctx, nil, hashIn{Hash: "zz"}); err == nil {
		t.Fatal("bad hash")
	}
	if _, _, err := getTorrent(ctx, nil, hashIn{Hash: strings.Repeat("ab", 20)}); err == nil {
		t.Fatal("missing")
	}

	title := "Renamed"
	if _, out, err := updateTorrent(ctx, nil, updateTorrentIn{Hash: hash, Title: &title}); err != nil || out.Title != title {
		t.Fatal(err, out.Title)
	}
	if _, _, err := updateTorrent(ctx, nil, updateTorrentIn{Hash: "nope"}); err == nil {
		t.Fatal("update bad")
	}

	if _, out, err := getPlayURL(ctx, nil, playURLIn{Hash: hash, FileIndex: 1}); err != nil || out.PlayURL == "" {
		t.Fatal(err, out)
	}
	if _, _, err := getPlayURL(ctx, nil, playURLIn{Hash: hash, FileIndex: 99}); err == nil {
		t.Fatal("bad index")
	}
	if _, out, err := getPlaylistURL(ctx, nil, playlistURLIn{Hash: hash}); err != nil || out.PlaylistURL == "" {
		t.Fatal(err)
	}
	if _, out, err := getPlaylistURL(ctx, nil, playlistURLIn{Category: "tv"}); err != nil || out.PlaylistURL == "" {
		t.Fatal(err)
	}

	if _, _, err := markViewed(ctx, nil, markViewedIn{Hash: hash, FileIndex: 0}); err == nil {
		t.Fatal("index")
	}
	if _, out, err := markViewed(ctx, nil, markViewedIn{Hash: hash, FileIndex: 1, TimeCode: 3}); err != nil || !out.OK {
		t.Fatal(err)
	}
	if _, out, err := listViewed(ctx, nil, listViewedIn{Hash: hash}); err != nil || out.Count < 1 {
		t.Fatal(err, out)
	}
	if _, _, err := getNextUnwatched(ctx, nil, nextUnwatchedIn{Query: "Show", Hash: hash, Category: "tv"}); err != nil {
		t.Fatal(err)
	}
	idx := 1
	if _, out, err := unmarkViewed(ctx, nil, unmarkViewedIn{Hash: hash, FileIndex: &idx}); err != nil || !out.OK {
		t.Fatal(err)
	}
	if _, _, err := unmarkViewed(ctx, nil, unmarkViewedIn{Hash: hash}); err != nil {
		t.Fatal(err)
	}
	set.ReadOnly = true
	if _, _, err := markViewed(ctx, nil, markViewedIn{Hash: hash, FileIndex: 1}); err == nil {
		t.Fatal("readonly viewed")
	}
	if _, _, err := removeTorrent(ctx, nil, hashIn{Hash: hash}); err == nil {
		t.Fatal("readonly remove")
	}
	set.ReadOnly = false

	if _, _, err := searchTorrents(ctx, nil, searchTorrentsIn{}); err == nil {
		t.Fatal("empty query")
	}
	if _, _, err := searchTorrents(ctx, nil, searchTorrentsIn{Query: "x", Source: "nope"}); err == nil {
		t.Fatal("bad source")
	}
	if _, out, err := searchTorrents(ctx, nil, searchTorrentsIn{Query: "x", Source: "rutor"}); err != nil || out.Message == "" {
		t.Fatal(err, out)
	}

	origR, origT := rutorSearch, torznabSearch
	t.Cleanup(func() { rutorSearch, torznabSearch = origR, origT })
	rutorSearch = func(q string) []*models.TorrentDetails {
		hits := []*models.TorrentDetails{nil, {Title: q, Hash: strings.Repeat("a", 40), Magnet: "magnet:?xt=urn:btih:" + strings.Repeat("a", 40)}}
		for i := 0; i < 5; i++ {
			hits = append(hits, &models.TorrentDetails{Title: q})
		}
		return hits
	}
	torznabSearch = func(context.Context, string, int, string, int, int) []*models.TorrentDetails {
		return []*models.TorrentDetails{{Title: "tz"}}
	}
	set.BTsets.EnableRutorSearch = true
	set.BTsets.EnableTorznabSearch = true
	if _, out, err := searchTorrents(ctx, nil, searchTorrentsIn{Query: "show", Source: "all", Limit: 2}); err != nil || out.Count != 2 {
		t.Fatal(err, out)
	}
	if _, out, err := searchTorrents(ctx, nil, searchTorrentsIn{Query: "show", Source: "torznab"}); err != nil || out.Count != 1 {
		t.Fatal(err, out)
	}
	set.BTsets.EnableRutorSearch = false
	set.BTsets.EnableTorznabSearch = false

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = listTorrents(ctx, nil, listTorrentsIn{})
			_, _, _ = getTorrent(ctx, nil, hashIn{Hash: hash})
		}()
	}
	wg.Wait()

	if _, out, err := dropTorrent(ctx, nil, hashIn{Hash: hash}); err != nil || !out.OK {
		t.Fatal(err)
	}
	if _, _, err := removeTorrent(ctx, nil, hashIn{Hash: strings.Repeat("cd", 20)}); err == nil {
		t.Fatal("missing remove")
	}
	if _, _, err := dropTorrent(ctx, nil, hashIn{Hash: "zz"}); err == nil {
		t.Fatal("drop bad")
	}
	if _, _, err := dropTorrent(ctx, nil, hashIn{Hash: strings.Repeat("cd", 20)}); err == nil {
		t.Fatal("drop missing")
	}
}

func boolPtr(v bool) *bool { return &v }
