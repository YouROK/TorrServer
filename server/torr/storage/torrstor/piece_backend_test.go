package torrstor

import (
	"bytes"
	"testing"

	"server/settings"
)

// Saving settings with a different UseDisk value must not break the pieces that exist at
// that moment: the engine keeps writing and reading them until the torrent is dropped.
func TestPieceKeepsItsBackendWhenUseDiskChanges(t *testing.T) {
	initTestSettings(t)
	oldUseDisk := settings.BTsets.UseDisk
	t.Cleanup(func() { settings.BTsets.UseDisk = oldUseDisk })

	c := NewCache(4<<20, nil)
	c.pieceLength = 1 << 20
	c.pieceCount = 1

	settings.BTsets.UseDisk = false
	p := NewPiece(0, c)
	c.pieces[0] = p

	// Keep the asynchronous cleanup that WriteAt/ReadAt trigger out of the way, as in
	// TestPartialPieceRAMAccounting: it needs a real torrent to release pieces.
	c.muRemove.Lock()
	defer c.muRemove.Unlock()
	defer p.mPiece.Release()

	settings.BTsets.UseDisk = true

	data := bytes.Repeat([]byte{0x42}, 16<<10)
	if _, err := p.WriteAt(data, 0); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(data))
	if _, err := p.ReadAt(got, 0); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("ReadAt returned different data")
	}
}
