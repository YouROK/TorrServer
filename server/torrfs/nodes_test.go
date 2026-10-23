package torrfs

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr"
	"server/torr/state"
)

func TestVFSNodesWithoutMount(t *testing.T) {
	dir := t.TempDir()
	oldPath, oldSets, oldRO, oldWA := settings.Path, settings.BTsets, settings.ReadOnly, settings.SearchWA
	settings.Path = dir
	t.Cleanup(func() {
		settings.CloseDB()
		settings.Path, settings.BTsets, settings.ReadOnly, settings.SearchWA = oldPath, oldSets, oldRO, oldWA
	})
	if err := settings.InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	settings.BTsets.ShowFSActiveTorr = false
	settings.BTsets.TorrentDisconnectTimeout = 0
	torr.InitApiHelper(torr.NewBTS())

	hash := strings.Repeat("ab", 20)
	spec := &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex(hash), DisplayName: "Show"}
	settings.AddTorrent(&settings.TorrentDB{TorrentSpec: spec, Title: "Show", Category: "tv"})

	root := NewRootDir()
	ents, err := root.ReadDir(-1)
	if err != nil || len(ents) != 1 || ents[0].Name() != "tv" {
		t.Fatalf("root %v %v", ents, err)
	}
	if n, err := root.ReadDir(0); err != nil || len(n) != 1 {
		t.Fatal(n, err)
	}
	if _, err := root.Read(nil); err != fs.ErrInvalid || root.Close() != nil || root.Parent() != nil || !root.IsDir() {
		t.Fatal("root file ops")
	}
	if _, err := root.Stat(); err != nil || root.Type() == 0 {
		t.Fatal("root stat")
	}
	opened, err := root.Open(".")
	if err != nil || opened != root {
		t.Fatal(err)
	}
	if _, err := root.Open("../x"); err == nil {
		t.Fatal("escape")
	}

	cat := NewCategoryDir("tv")
	ents, err = cat.ReadDir(-1)
	if err != nil || len(ents) != 1 {
		t.Fatalf("cat %v %v", ents, err)
	}
	if _, err := cat.Open("nope"); err == nil {
		t.Fatal("missing child")
	}
	if f, err := cat.Open("tv"); err != nil || f != cat {
		t.Fatalf("self %v %v", f, err)
	}
	if _, err := cat.Read(nil); err != fs.ErrInvalid || cat.Torrent() != nil {
		t.Fatal("cat file")
	}
	cat.SetTorrent(nil)

	live := &torr.Torrent{TorrentSpec: spec, Title: "Show", Timestamp: 10, Stat: state.TorrentPreload}
	td := NewTorrDir(cat, "Show", live)
	ents, err = td.ReadDir(-1)
	if err != nil || len(ents) != 0 {
		t.Fatalf("empty files %v %v", ents, err)
	}
	live.Stat = state.TorrentClosed
	if _, err := td.ReadDir(1); err == nil {
		t.Fatal("closed torrent should not list files")
	}
	if td.Parent() != cat || td.Torrent() != live {
		t.Fatal("parent")
	}
	if _, err := td.Stat(); err != nil {
		t.Fatal(err)
	}
	if _, err := td.Read(nil); err != fs.ErrInvalid || td.Close() != nil {
		t.Fatal("read")
	}
	if f, err := td.Open("Show"); err != nil || f != td {
		t.Fatal(err)
	}
	if _, err := Open(td, "Show/missing"); err == nil {
		t.Fatal("node open")
	}
	if _, err := (&TorrDir{info: info{name: "empty"}, torr: nil}).ReadDir(-1); err == nil {
		t.Fatal("nil torrent")
	}
	tf := &TorrFile{parent: td, torr: live, info: info{name: "a.mkv", size: 4, mode: 0o444}}
	if tf.Name() != "a.mkv" || tf.IsDir() || tf.Parent() != td || tf.Torrent() != live {
		t.Fatal("file node")
	}
	tf.SetTorrent(live)
	if _, err := tf.Stat(); err != nil || tf.Type() == 0 {
		t.Fatal("file stat")
	}
	if _, err := tf.Read(nil); err != fs.ErrInvalid || tf.Close() != nil {
		t.Fatal("file read")
	}
	if _, err := tf.ReadDir(0); err != fs.ErrInvalid {
		t.Fatal("file readdir")
	}
	if _, err := tf.Info(); err != nil {
		t.Fatal(err)
	}
}
