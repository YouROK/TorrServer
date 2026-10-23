package torr

import (
	"io"
	"testing"

	"github.com/anacrolix/torrent"

	"server/torr/offlinetest"
)

func BenchmarkTorrReaderSeekAndRead(b *testing.B) {
	offlinetest.Prepare(b)
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(bt.Disconnect)

	show := offlinetest.WriteShow(b)
	seed := offlinetest.Seed(b, show)
	tor, err := AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil || !tor.GotInfo() {
		b.Fatal(err)
	}
	tor.AddPeers([]torrent.Peer{seed.Peer()})

	var file *torrent.File
	for _, f := range tor.Files() {
		if f.Length() > 4096 {
			file = f
			break
		}
	}
	if file == nil {
		b.Fatal("no file")
	}
	r := tor.NewReader(file)
	if r == nil {
		b.Fatal("reader")
	}
	b.Cleanup(func() { tor.CloseReader(r) })
	r.SetResponsive()
	buf := make([]byte, 4096)
	if _, err := io.ReadFull(r, buf); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(buf)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Seek(int64((i*4096)%(32*1024)), io.SeekStart); err != nil {
			b.Fatal(err)
		}
		if _, err := io.ReadFull(r, buf); err != nil {
			b.Fatal(err)
		}
	}
}
