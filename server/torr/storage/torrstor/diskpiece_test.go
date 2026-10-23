package torrstor

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

func TestDiskPieceReadWriteRelease(t *testing.T) {
	testBTSets(t)
	dir := t.TempDir()
	settings.BTsets.UseDisk = true
	settings.BTsets.TorrentsSavePath = dir

	c := &Cache{
		pieces:       map[int]*Piece{},
		activePieces: map[int]struct{}{},
		hash:         metainfo.Hash{9},
		pieceLength:  8,
		pieceCount:   2,
		totalLength:  8,
		capacity:     1 << 20,
	}
	p := &Piece{Id: 0, cache: c}
	c.pieces[0] = p
	if err := os.MkdirAll(filepath.Join(dir, c.hash.HexString()), 0o777); err != nil {
		t.Fatal(err)
	}
	d := NewDiskPiece(p)
	if _, err := d.ReadAt(make([]byte, 2), 0); !os.IsNotExist(err) {
		t.Fatalf("missing read %v", err)
	}
	n, err := d.WriteAt([]byte("abcdefghij"), 0)
	if err != nil || n != 10 {
		t.Fatalf("write %d %v", n, err)
	}
	if p.Size != 8 || len(c.activePieces) != 1 {
		t.Fatalf("size %d active %d", p.Size, len(c.activePieces))
	}
	again := NewDiskPiece(p)
	buf := make([]byte, 4)
	n, err = again.ReadAt(buf, 0)
	if err != nil || string(buf) != "abcd" {
		t.Fatalf("read %q %d %v", buf, n, err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = again.WriteAt([]byte("zz"), 2)
			b := make([]byte, 2)
			_, _ = again.ReadAt(b, 0)
		}()
	}
	wg.Wait()
	again.Release()
	if _, err := os.Stat(again.name); !os.IsNotExist(err) {
		t.Fatalf("release left file: %v", err)
	}
	c.isClosed.Store(true)
	if _, err := again.WriteAt([]byte("x"), 0); !os.IsNotExist(err) {
		t.Fatalf("closed write %v", err)
	}
	if _, err := again.ReadAt(buf, 0); !os.IsNotExist(err) {
		t.Fatalf("closed read %v", err)
	}
}

func TestMemPieceLifecycle(t *testing.T) {
	testBTSets(t)
	settings.BTsets.UseDisk = false
	c := &Cache{
		pieces:       map[int]*Piece{},
		activePieces: map[int]struct{}{},
		readers:      map[*Reader]struct{}{},
		pieceLength:  16,
		pieceCount:   1,
		totalLength:  16,
		capacity:     1 << 30,
	}
	p := NewPiece(0, c)
	c.pieces[0] = p
	if n, err := p.WriteAt([]byte("hello"), 0); err != nil || n != 5 {
		t.Fatalf("write %d %v", n, err)
	}
	buf := make([]byte, 5)
	if n, err := p.ReadAt(buf, 0); err != nil || string(buf[:n]) != "hello" {
		t.Fatalf("read %q %v", buf[:n], err)
	}
	if err := p.MarkComplete(); err != nil || !p.Completion().Complete {
		t.Fatal(err)
	}
	if _, err := p.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if err := p.MarkNotComplete(); err != nil || p.Completion().Complete {
		t.Fatal("not complete")
	}
	if _, err := p.ReadAt(buf, 0); !os.IsNotExist(err) {
		t.Fatalf("released read %v", err)
	}
	empty := &Piece{cache: c}
	if err := empty.MarkComplete(); err != nil || !empty.Completion().Complete {
		t.Fatal("bare piece")
	}
	if err := empty.MarkNotComplete(); err != nil || empty.Size != 0 {
		t.Fatal(err)
	}
}
