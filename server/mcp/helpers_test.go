package mcp

import (
	"strings"
	"testing"

	"server/torrshash"
)

func TestValidInfoHashRejectsNonHex(t *testing.T) {
	good := strings.Repeat("aB09", 10)
	if !validInfoHash(good) {
		t.Errorf("%q rejected", good)
	}
	for _, bad := range []string{
		"", good[:39], good + "a", strings.Repeat("g", 40),
		strings.Repeat("a", 39) + " ", "../" + strings.Repeat("a", 37),
		strings.Repeat("я", 20), // 40 bytes, but not hex
	} {
		if validInfoHash(bad) {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestParseTorrentLinkMagnetAndHash(t *testing.T) {
	const hash = "e3e497e1efd8db8df93822cad554e13740f51997"

	for _, link := range []string{
		"magnet:?xt=urn:btih:" + hash + "&dn=Show",
		"  magnet:?xt=urn:btih:" + hash + "&amp;dn=Show  ", // HTML-escaped by a chat client
		hash, // bare info hash
	} {
		spec, title, poster, cat, err := parseTorrentLink(link)
		if err != nil {
			t.Fatalf("%q: %v", link, err)
		}
		if spec.InfoHash.HexString() != hash {
			t.Errorf("%q: hash = %s", link, spec.InfoHash.HexString())
		}
		if title != "" || poster != "" || cat != "" {
			t.Errorf("%q: plain links carry no metadata, got %q %q %q", link, title, poster, cat)
		}
	}
	// &amp; must be decoded, otherwise the display name is "Show" glued to "amp;".
	spec, _, _, _, err := parseTorrentLink("magnet:?xt=urn:btih:" + hash + "&amp;dn=Show")
	if err != nil || spec.DisplayName != "Show" {
		t.Errorf("&amp; magnet: %+v %v", spec, err)
	}
}

func TestParseTorrentLinkTorrsCarriesMetadata(t *testing.T) {
	th := torrshash.New(strings.Repeat("a", 40))
	th.AddField(torrshash.TagTitle, "Show A")
	token, err := torrshash.Pack(th)
	if err != nil {
		t.Fatal(err)
	}
	spec, title, _, _, err := parseTorrentLink("TORRS://" + token) // scheme is case-insensitive
	if err != nil {
		t.Fatal(err)
	}
	if title != "Show A" || spec.InfoHash.HexString() != strings.Repeat("a", 40) {
		t.Errorf("title=%q hash=%s", title, spec.InfoHash.HexString())
	}
}

func TestParseTorrentLinkRejects(t *testing.T) {
	for _, bad := range []string{"", "   ", "ftp://example.com/x.torrent", "torrs://", "magnet:?dn=nohash"} {
		if _, _, _, _, err := parseTorrentLink(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
