package torrfs

import (
	"fmt"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr"
)

func BenchmarkVFSReadDir(b *testing.B) {
	dir := b.TempDir()
	oldPath, oldSets := settings.Path, settings.BTsets
	settings.Path = dir
	b.Cleanup(func() {
		settings.CloseDB()
		settings.Path, settings.BTsets = oldPath, oldSets
	})
	if err := settings.InitSets(false, false); err != nil {
		b.Fatal(err)
	}
	settings.BTsets.TorrentDisconnectTimeout = 0
	settings.BTsets.ShowFSActiveTorr = false
	torr.InitApiHelper(torr.NewBTS())

	for i := 0; i < 32; i++ {
		hash := fmt.Sprintf("%040x", i+1)
		spec := &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex(hash), DisplayName: "Show"}
		cat := []string{"tv", "movie", "music", "other"}[i%4]
		settings.AddTorrent(&settings.TorrentDB{TorrentSpec: spec, Title: "Show", Category: cat})
	}
	root := NewRootDir()
	tv := NewCategoryDir("tv")

	b.Run("root", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ents, err := root.ReadDir(-1)
			if err != nil || len(ents) != 4 {
				b.Fatalf("%v %d", err, len(ents))
			}
		}
	})
	b.Run("category", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ents, err := tv.ReadDir(-1)
			if err != nil || len(ents) == 0 {
				b.Fatalf("%v %d", err, len(ents))
			}
		}
	})
}
