package mcp

import (
	"testing"

	"server/torr/state"
)

func TestParseEpisode(t *testing.T) {
	cases := []struct {
		path       string
		season, ep int
		ok         bool
	}{
		{"Show.Name.S01E05.1080p.mkv", 1, 5, true},
		{"show.s1e5.mkv", 1, 5, true},
		{"Show Name - 1x05.mkv", 1, 5, true},
		{"Show/Season 1/Episode 05.mkv", 1, 5, true},
		{"Show/Season 02/Show - Episode 03.mp4", 2, 3, true},
		{"Show/сезон 1/серия 08.mkv", 1, 8, true},
		{"Show.S01.E05.mkv", 1, 5, true},
		{"Movie.Name.2020.1080p.BluRay.mkv", 0, 0, false},
		{"Concert.Live.2160p.webm", 0, 0, false},
		{"Show.Name.s02e10.mkv", 2, 10, true},
		{`Show\Season 3\E12.mkv`, 3, 12, true},
		{"", 0, 0, false},
	}
	for _, tc := range cases {
		s, e, ok := ParseEpisode(tc.path)
		if ok != tc.ok || s != tc.season || e != tc.ep {
			t.Errorf("ParseEpisode(%q)=%d,%d,%v want %d,%d,%v", tc.path, s, e, ok, tc.season, tc.ep, tc.ok)
		}
	}
}

func TestParseEpisodeDoesNotTreatResolutionAsEpisode(t *testing.T) {
	s, e, ok := ParseEpisode("Movie.1920x1080.mkv")
	if ok {
		t.Fatalf("ParseEpisode treated resolution as episode: season=%d episode=%d", s, e)
	}
}

func files(paths ...string) []*state.TorrentFileStat {
	out := make([]*state.TorrentFileStat, len(paths))
	for i, p := range paths {
		out[i] = &state.TorrentFileStat{Id: i + 1, Path: p, Length: 1000}
	}
	return out
}

func TestSelectNextUnwatchedMidSeason(t *testing.T) {
	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	snaps := []TorrentSnapshot{{
		Title:    "Some Show",
		Category: "tv",
		Hash:     hash,
		Files: files(
			"Some.Show.S01E01.mkv",
			"Some.Show.S01E02.mkv",
			"Some.Show.S01E03.mkv",
			"Some.Show.S01E01.en.srt",
		),
	}}
	viewed := ViewedMap{hash: {1: 0, 2: 0}}
	res := SelectNextUnwatched(snaps, viewed, "", "tv", "")
	if res.CaughtUp {
		t.Fatal("expected next episode, got caught up")
	}
	if res.Season != 1 || res.Episode != 3 || res.FileIndex != 3 {
		t.Fatalf("got S%02dE%02d file %d, want S01E03 file 3", res.Season, res.Episode, res.FileIndex)
	}
	if res.RemainingUnwatched != 1 {
		t.Fatalf("remaining=%d want 1", res.RemainingUnwatched)
	}
	if res.Code != "S01E03" {
		t.Fatalf("code=%q want S01E03", res.Code)
	}
}

func TestSelectNextUnwatchedCaughtUp(t *testing.T) {
	hash := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	snaps := []TorrentSnapshot{{
		Title:    "Some Show",
		Category: "tv",
		Hash:     hash,
		Files:    files("Some.Show.S01E01.mkv", "Some.Show.S01E02.mkv"),
	}}
	viewed := ViewedMap{hash: {1: 0, 2: 0}}
	res := SelectNextUnwatched(snaps, viewed, "", "tv", "")
	if !res.CaughtUp {
		t.Fatalf("expected caught up, got %+v", res)
	}
	if res.LastWatched == nil || res.LastWatched.Episode != 2 {
		t.Fatalf("last watched = %+v, want episode 2", res.LastWatched)
	}
	if res.RemainingUnwatched != 0 {
		t.Fatalf("remaining=%d want 0", res.RemainingUnwatched)
	}
}

func TestSelectNextUnwatchedMultiTorrentShow(t *testing.T) {
	h1 := "cccccccccccccccccccccccccccccccccccccccc"
	h2 := "dddddddddddddddddddddddddddddddddddddddd"
	snaps := []TorrentSnapshot{
		{
			Title:    "Show Season 1",
			Category: "tv",
			Hash:     h1,
			Files:    files("Show.S01E01.mkv", "Show.S01E02.mkv"),
		},
		{
			Title:    "Show Season 2",
			Category: "tv",
			Hash:     h2,
			Files:    files("Show.S02E01.mkv", "Show.S02E02.mkv"),
		},
	}
	viewed := ViewedMap{h1: {1: 0, 2: 0}}
	res := SelectNextUnwatched(snaps, viewed, "show", "tv", "")
	if res.CaughtUp {
		t.Fatal("expected next episode from season 2")
	}
	if res.Hash != h2 || res.Season != 2 || res.Episode != 1 {
		t.Fatalf("got hash=%s S%02dE%02d, want season 2 torrent E01", res.Hash, res.Season, res.Episode)
	}
}

func TestSelectNextUnwatchedSkipsSample(t *testing.T) {
	hash := "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	snaps := []TorrentSnapshot{{
		Title:    "Show",
		Category: "tv",
		Hash:     hash,
		Files: files(
			"Show.S01E01.sample.mkv",
			"Show.S01E01.mkv",
		),
	}}
	res := SelectNextUnwatched(snaps, ViewedMap{}, "", "tv", "")
	if res.CaughtUp || res.FileIndex != 2 {
		t.Fatalf("expected real episode file 2, got %+v", res)
	}
}

func TestSelectNextUnwatchedNoMatch(t *testing.T) {
	res := SelectNextUnwatched(nil, nil, "missing", "tv", "")
	if !res.CaughtUp || res.Message == "" {
		t.Fatalf("expected no-match result, got %+v", res)
	}
}

func TestSelectNextUnwatchedQueryFiltersTitle(t *testing.T) {
	snaps := []TorrentSnapshot{
		{Title: "Alpha Show", Category: "tv", Hash: "1111111111111111111111111111111111111111", Files: files("Alpha.S01E01.mkv")},
		{Title: "Beta Show", Category: "tv", Hash: "2222222222222222222222222222222222222222", Files: files("Beta.S01E01.mkv")},
	}
	res := SelectNextUnwatched(snaps, ViewedMap{}, "beta", "tv", "")
	if res.ShowTitle != "Beta Show" {
		t.Fatalf("got title %q want Beta Show", res.ShowTitle)
	}
}

func TestIsSampleFile(t *testing.T) {
	if !isSampleFile("Show.S01E01.sample.mkv") {
		t.Fatal("expected sample")
	}
	if isSampleFile("Show.S01E01.mkv") {
		t.Fatal("did not expect sample")
	}
}
