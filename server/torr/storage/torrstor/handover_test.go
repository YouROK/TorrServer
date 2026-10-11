package torrstor

import (
	"fmt"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
)

// ViMu drops its connection once its buffer is full and comes back when it needs more — after a
// pause, many minutes later. By then the torrent has been closed for want of readers, and the
// cache that kept the record went with it. Seen on 2026-10-10: the connection after a twelve-minute
// pause found nothing to inherit, measured 137MB of a 300MB buffer, and kept the position twenty
// seconds in front of the picture for the rest of the film.
func TestHandoverOutlivesTheTorrent(t *testing.T) {
	resetHandovers()
	hash := metainfo.NewHashFromHex("42b4115a3b4ffe0ecbfb628c5943066a3720cebf")
	before := &Cache{hash: hash}
	before.noteRead("film.mkv", 1, 300<<20, 300<<20, 11760<<20, 1475)

	after := &Cache{hash: hash} // the same torrent, opened again
	held, size, ok := after.takeOver("film.mkv", 12025<<20)
	if !ok || held != 265<<20 || size != 300<<20 {
		t.Fatalf("после переоткрытия торрента перенесено %dМБ из %dМБ (%v), ждали 265 из 300", held>>20, size>>20, ok)
	}
}

func TestHandoverBelongsToItsTorrent(t *testing.T) {
	resetHandovers()
	a := &Cache{hash: metainfo.NewHashFromHex("42b4115a3b4ffe0ecbfb628c5943066a3720cebf")}
	b := &Cache{hash: metainfo.NewHashFromHex("d797565ccceec75ef1500cc4520f7935712f7ed2")}
	a.noteRead("S01E01.mkv", 1, 300<<20, 300<<20, 600<<20, 60)
	if _, _, ok := b.takeOver("S01E01.mkv", 900<<20); ok {
		t.Fatal("файл с тем же именем из другого торрента унаследовал чужой буфер")
	}
}

// Only the files paused on the devices in the house matter; the rest can go.
func TestHandoversKeepOnlyTheLatest(t *testing.T) {
	resetHandovers()
	c := &Cache{hash: metainfo.NewHashFromHex("42b4115a3b4ffe0ecbfb628c5943066a3720cebf")}
	for i := 0; i <= maxHandovers; i++ {
		c.noteRead(fmt.Sprintf("E%02d.mkv", i), int64(i+1), 300<<20, 300<<20, 600<<20, 60)
	}
	if _, _, ok := c.takeOver("E00.mkv", 900<<20); ok {
		t.Fatal("самая давняя запись не вытеснена")
	}
	if _, _, ok := c.takeOver(fmt.Sprintf("E%02d.mkv", maxHandovers), 900<<20); !ok {
		t.Fatal("последняя запись потерялась")
	}
}
