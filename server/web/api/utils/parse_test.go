package utils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

func testMeta(t *testing.T) []byte {
	t.Helper()
	info := metainfo.Info{PieceLength: 16 * 1024, Name: "a.bin", Length: 4, Pieces: make([]byte, 20)}
	mi := metainfo.MetaInfo{InfoBytes: bencode.MustMarshal(info)}
	var buf bytes.Buffer
	if err := mi.Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type seekCloser struct{ *bytes.Reader }

func (seekCloser) Close() error { return nil }

func TestParseLinkAndFile(t *testing.T) {
	raw := testMeta(t)
	spec, err := ParseFromBytes(raw)
	if err != nil || spec.DisplayName != "a.bin" {
		t.Fatal(err, spec)
	}
	spec, err = ParseFile(seekCloser{bytes.NewReader(raw)})
	if err != nil || len(spec.InfoBytes) == 0 {
		t.Fatal(err)
	}

	hash := spec.InfoHash.HexString()
	spec, err = ParseLink("magnet:?xt=urn:btih:" + hash + "&dn=Show")
	if err != nil || spec.InfoHash.HexString() != hash {
		t.Fatal(err)
	}
	spec, err = ParseLink(hash)
	if err != nil || spec.InfoHash.HexString() != hash {
		t.Fatal("bare hash", err)
	}
	if _, err := ParseLink("ftp://example/a.torrent"); err == nil {
		t.Fatal("ftp")
	}
	if _, err := ParseLink("magnet:?xt=urn:btih:zz"); err == nil {
		t.Fatal("bad magnet")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "a.torrent")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseLink("file://" + path); err != nil {
		t.Fatal(err)
	}
	if _, err := fromFile(filepath.Join(dir, "missing.torrent")); err == nil {
		t.Fatal("missing file")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/no" {
			http.Error(w, "no", http.StatusNotFound)
			return
		}
		_, _ = w.Write(raw)
	}))
	defer srv.Close()
	if _, err := ParseLink(srv.URL + "/a.torrent"); err != nil {
		t.Fatal(err)
	}
	if _, err := fromHttp(srv.URL + "/no"); err == nil {
		t.Fatal("non-200")
	}

	if err := TestLink("magnet:?xt=urn:btih:"+hash, false); err != nil {
		t.Fatal(err)
	}
	if err := TestLink(hash, false); err != nil {
		t.Fatal(err)
	}
	if err := TestLink("http://example/a.torrent", false); err == nil {
		t.Fatal("http without auth")
	}
	if err := TestLink("http://example/a.torrent", true); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseFromBytes([]byte("not a torrent")); err == nil {
		t.Fatal("garbage")
	}
}
