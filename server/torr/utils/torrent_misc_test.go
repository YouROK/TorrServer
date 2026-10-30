package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/time/rate"

	"server/settings"
)

// The qB peer id prefix is what swarms key on (AGENTS.md): the id must stay a
// 20-byte string that starts with the prefix and is random after it.
func TestPeerIDRandom(t *testing.T) {
	const prefix = "-qB4390-"
	a, b := PeerIDRandom(prefix), PeerIDRandom(prefix)
	for _, id := range []string{a, b} {
		if len(id) != 20 || !strings.HasPrefix(id, prefix) {
			t.Fatalf("peer id %q: want 20 bytes starting with %q", id, prefix)
		}
		for _, c := range id[len(prefix):] {
			if (c < 'A' || c > 'Z') && (c < '2' || c > '7') {
				t.Fatalf("peer id %q has a non-base32 byte %q", id, c)
			}
		}
	}
	if a == b {
		t.Error("two peer ids are identical")
	}
	if got := PeerIDRandom(strings.Repeat("x", 20)); len(got) != 20 {
		t.Errorf("full-length prefix = %q", got)
	}
}

func TestLimit(t *testing.T) {
	if l := Limit(0); l.Limit() != rate.Inf {
		t.Errorf("Limit(0) = %v, want unlimited", l.Limit())
	}
	if l := Limit(-5); l.Limit() != rate.Inf {
		t.Errorf("Limit(-5) = %v, want unlimited", l.Limit())
	}
	// A tiny limit still gets a 16KiB burst: a smaller one would make every
	// WaitN of a normal block fail instead of wait.
	l := Limit(1000)
	if l.Limit() != 1000 || l.Burst() != 16*1024 {
		t.Errorf("Limit(1000) = %v/%d", l.Limit(), l.Burst())
	}
	// A large limit uses itself as the burst.
	l = Limit(5 << 20)
	if l.Burst() != 5<<20 {
		t.Errorf("Limit(5MiB) burst = %d", l.Burst())
	}
}

func TestOpenTorrentFile(t *testing.T) {
	dir := t.TempDir()
	payload := filepath.Join(dir, "payload.bin")
	if err := os.WriteFile(payload, make([]byte, 40_000), 0o600); err != nil {
		t.Fatal(err)
	}
	info := metainfo.Info{PieceLength: 16 * 1024}
	if err := info.BuildFromFilePath(payload); err != nil {
		t.Fatal(err)
	}
	infoBytes, err := bencode.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	mi := metainfo.MetaInfo{InfoBytes: infoBytes, Announce: "udp://198.51.100.1:6969/announce"}
	file := filepath.Join(dir, "x.torrent")
	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := mi.Write(f); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	spec, err := OpenTorrentFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if spec.DisplayName != "payload.bin" || spec.InfoHash != mi.HashInfoBytes() || !reflect.DeepEqual(spec.InfoBytes, infoBytes) {
		t.Fatalf("spec = %+v", spec)
	}
	if len(spec.Trackers) != 1 {
		t.Errorf("trackers = %v, want one tier", spec.Trackers)
	}

	if _, err := OpenTorrentFile(filepath.Join(dir, "missing.torrent")); err == nil {
		t.Error("missing file accepted")
	}
	bad := filepath.Join(dir, "bad.torrent")
	if err := os.WriteFile(bad, []byte("not bencode"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenTorrentFile(bad); err == nil {
		t.Error("garbage accepted")
	}
}

// SetDefTrackers copies its input (the caller may reuse the slice), mirrors it
// into BTsets, and must not start a list fetch to the real mirrors.
func TestSetDefTrackersCopiesAndMirrorsToSettings(t *testing.T) {
	setupTrackersTestWithURLs(t, "", nil, "")
	oldFallback := append([]string(nil), fallbackTrackers...)
	t.Cleanup(func() {
		fallbackTrackersMu.Lock()
		fallbackTrackers = oldFallback
		fallbackTrackersMu.Unlock()
	})

	in := []string{"udp://198.51.100.1:1/announce", "udp://198.51.100.2:2/announce"}
	SetDefTrackers(in)
	in[0] = "udp://tampered:1/announce"

	fallbackTrackersMu.RLock()
	got := append([]string(nil), fallbackTrackers...)
	fallbackTrackersMu.RUnlock()
	if got[0] != "udp://198.51.100.1:1/announce" || len(got) != 2 {
		t.Fatalf("fallback = %v, caller's slice leaked in", got)
	}
	if want := "udp://198.51.100.1:1/announce\nudp://198.51.100.2:2/announce"; settings.BTsets.DefaultTrackers != want {
		t.Errorf("BTsets.DefaultTrackers = %q", settings.BTsets.DefaultTrackers)
	}

	// JNI/tests run with BTsets == nil.
	settings.BTsets = nil
	SetDefTrackers(nil)
}

func TestSanitizeTrackersNormalisesAndBounds(t *testing.T) {
	long := "udp://tracker.example:80/" + strings.Repeat("a", 2048-len("udp://tracker.example:80/"))
	if len(long) != 2048 {
		t.Fatalf("test URL is %d bytes", len(long))
	}
	cases := []struct {
		name, in string
		want     string // "" means dropped
	}{
		{"2048 bytes is the last accepted length", long, long},
		{"2049 bytes is dropped", long + "a", ""},
		{"scheme and host fold to lower case", "UDP://Tracker.EXAMPLE:80/Announce", "udp://tracker.example:80/Announce"},
		{"trailing slash and fragment are dropped", "http://t.example/announce/#frag", "http://t.example/announce"},
		{"port 0 is invalid", "udp://t.example:0/announce", ""},
		{"port 65536 is invalid", "udp://t.example:65536/announce", ""},
		{"port 65535 is valid", "udp://t.example:65535/announce", "udp://t.example:65535/announce"},
		{"IPv6 literal over udp is kept", "udp://[2001:db8::1]:6969/announce", "udp://[2001:db8::1]:6969/announce"},
		{"IPv6 literal over https has no SNI", "https://[2001:db8::1]:443/announce", ""},
		{"no host", "udp:///announce", ""},
		{"ftp", "ftp://t.example/announce", ""},
		{"magnet is not a tracker", "magnet:?xt=urn:btih:aaaa", ""},
	}
	for _, c := range cases {
		got := SanitizeTrackers([]string{c.in})
		if c.want == "" {
			if len(got) != 0 {
				t.Errorf("%s: kept %v", c.name, got)
			}
			continue
		}
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%s: got %v, want [%s]", c.name, got, c.want)
		}
	}
}

// Raw-IP trackers collapse on host:port whatever the scheme or path; named
// hosts are distinct by full URL.
func TestSanitizeTrackersDedupeKeys(t *testing.T) {
	got := SanitizeTrackers([]string{
		"udp://[2001:db8::1]:80/a",
		"http://[2001:db8:0::1]:80/b", // same address spelled differently, same port
		"udp://198.51.100.9:80/a",
		"udp://tracker.example:80/a",
		"udp://tracker.example:80/b",  // named host: a different path is a different tracker
		"UDP://TRACKER.example:80/a/", // same as the first named one
	})
	want := []string{
		"udp://[2001:db8::1]:80/a",
		"udp://198.51.100.9:80/a",
		"udp://tracker.example:80/a",
		"udp://tracker.example:80/b",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}
