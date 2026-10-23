package torrstor

import (
	"testing"

	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

func BenchmarkTorrstorPieceWriteRead(b *testing.B) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{UseDisk: false}
	b.Cleanup(func() { settings.BTsets = prev })

	const pieceLen = 64 << 10
	const nPieces = 32
	stor := NewStorage(4 << 20)
	info := &metainfo.Info{
		Name:        "bench",
		PieceLength: pieceLen,
		Length:      pieceLen * nPieces,
		Pieces:      make([]byte, 20*nPieces),
	}
	var h metainfo.Hash
	h[0] = 0xb1
	if _, err := stor.OpenTorrent(info, h); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = stor.Close() })
	c := stor.GetCache(h)
	payload := make([]byte, pieceLen)
	readBuf := make([]byte, 4096)

	b.Run("sequential", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(pieceLen)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p := c.pieces[i%nPieces]
			if _, err := p.WriteAt(payload, 0); err != nil {
				b.Fatal(err)
			}
			_ = p.MarkComplete()
			_, _ = p.ReadAt(readBuf, int64((i*4096)%(pieceLen-4096)))
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(pieceLen)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			buf := make([]byte, 4096)
			i := 0
			for pb.Next() {
				p := c.pieces[i%nPieces]
				i++
				_, _ = p.WriteAt(payload, 0)
				_ = p.MarkComplete()
				_, _ = p.ReadAt(buf, 0)
			}
		})
	})
}
