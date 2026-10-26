package torr

import "testing"

// A hash that is not 40 hex digits comes from a request as is: the base32 btih of a magnet link (32
// characters), an md5, garbage. None of these functions may panic on it - it is an unknown torrent.
func TestBadHashNoPanic(t *testing.T) {
	calls := map[string]func(string){
		"GetTorrent": func(h string) {
			if tor := GetTorrent(h); tor != nil {
				t.Errorf("GetTorrent(%q) = %v, want nil", h, tor)
			}
		},
		"SetTorrent": func(h string) {
			if tor := SetTorrent(h, "title", "", "", ""); tor != nil {
				t.Errorf("SetTorrent(%q) = %v, want nil", h, tor)
			}
		},
		"RemTorrent":  func(h string) { RemTorrent(h) },
		"DropTorrent": func(h string) { DropTorrent(h) },
	}
	for _, h := range []string{
		"MFRGGZDFMZTWQ2LKNNWG23TPOBYXE43U",          // base32 btih of a magnet link
		"d41d8cd98f00b204e9800998ecf8427e",          // 32 hex
		"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",  // 40 characters, not hex
		"0123456789abcdef0123456789abcdef0123456",   // 39
		" 0123456789abcdef0123456789abcdef01234567", // 41, leading space
	} {
		for name, call := range calls {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s(%q): panic: %v", name, h, r)
					}
				}()
				call(h)
			}()
		}
	}
}
