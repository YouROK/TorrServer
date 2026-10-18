package mcp

import (
	"net/http"
	"testing"

	"server/torr/state"
)

func TestPlayURL(t *testing.T) {
	u := playURL("http://127.0.0.1:8090", "abc", "Season 1/Show.S01E01.mkv", 3)
	want := "http://127.0.0.1:8090/stream/Show.S01E01.mkv?link=abc&index=3&play"
	if u != want {
		t.Fatalf("playURL=%q want %q", u, want)
	}
}

func TestShortPlayURL(t *testing.T) {
	u := shortPlayURL("http://192.168.1.10:8090", "deadbeef", 2)
	want := "http://192.168.1.10:8090/play/deadbeef/2"
	if u != want {
		t.Fatalf("shortPlayURL=%q want %q", u, want)
	}
}

func TestPlaylistURL(t *testing.T) {
	all := playlistURL("http://host:8090", "")
	if all != "http://host:8090/playlistall/all.m3u" {
		t.Fatalf("all playlist=%q", all)
	}
	one := playlistURL("http://host:8090", "abcd")
	if one != "http://host:8090/playlist?hash=abcd" {
		t.Fatalf("one playlist=%q", one)
	}
}

func TestFormatEpisodeCode(t *testing.T) {
	if got := formatEpisodeCode(1, 5); got != "S01E05" {
		t.Fatalf("got %q", got)
	}
	if got := formatEpisodeCode(0, 3); got != "E03" {
		t.Fatalf("got %q", got)
	}
	if got := formatEpisodeCode(2, 0); got != "S02" {
		t.Fatalf("got %q", got)
	}
	if got := formatEpisodeCode(0, 0); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestBaseURLFromRequest(t *testing.T) {
	r := &http.Request{
		Host:   "example.local:8090",
		Header: http.Header{"X-Forwarded-Proto": []string{"https"}},
	}
	if got := baseURLFromRequest(r); got != "https://example.local:8090" {
		t.Fatalf("got %q", got)
	}
}

func TestFilePlayURLs(t *testing.T) {
	st := &state.TorrentStatus{Hash: "hashhash"}
	f := &state.TorrentFileStat{Id: 1, Path: "a.mkv"}
	play, short := filePlayURLs("http://h", st, f)
	if play == "" || short == "" {
		t.Fatalf("empty urls play=%q short=%q", play, short)
	}
}

func TestValidInfoHash(t *testing.T) {
	if !validInfoHash("0123456789abcdef0123456789ABCDEF01234567") {
		t.Fatal("expected valid")
	}
	if validInfoHash("short") || validInfoHash("") || validInfoHash("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz") {
		t.Fatal("expected invalid")
	}
}
