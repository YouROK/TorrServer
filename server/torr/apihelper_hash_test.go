package torr

import (
	"encoding/base32"
	"encoding/hex"
	"strings"
	"testing"
)

// API handlers pass the hash query field straight into the torr helpers.
// A 32-char base32 btih is a legal info hash (ParseLink accepts it from
// magnet links), so a client that kept the hash from the link instead of
// the add response must get the same torrent back, not a failure.
func TestHashFromHexAcceptsBase32(t *testing.T) {
	raw := []byte{0xde, 0xad, 0xbe, 0xef, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55,
		0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	hexForm := hex.EncodeToString(raw)
	b32Form := base32.StdEncoding.EncodeToString(raw)

	hHex, ok := hashFromHex(hexForm)
	if !ok {
		t.Fatalf("40-char hex info hash rejected")
	}
	if got := hHex.HexString(); got != hexForm {
		t.Fatalf("hex info hash decoded to %q, want %q", got, hexForm)
	}

	hB32, ok := hashFromHex(b32Form)
	if !ok {
		t.Fatalf("32-char base32 info hash %q rejected", b32Form)
	}
	if got := hB32.HexString(); got != hexForm {
		t.Fatalf("base32 info hash decoded to %q, want %q", got, hexForm)
	}
}

func TestHashFromHexRejectsGarbage(t *testing.T) {
	for _, s := range []string{
		"",
		"abc",
		strings.Repeat("z", 32),
		strings.Repeat("!", 40),
		strings.Repeat("g", 40),
		strings.Repeat("0", 39),
		strings.Repeat("0", 41),
	} {
		if _, ok := hashFromHex(s); ok {
			t.Errorf("hashFromHex(%q) accepted an invalid info hash", s)
		}
	}
}
