package subs

import "testing"

func TestMatchLimitSizes(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Show/ep01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show/ep01.rus.srt", Size: 50_000},
		{Index: 3, Path: "Show/ep01.eng.ass", Size: 8 << 20},
	}

	got, large := MatchLimit(files[0], files, 2<<20)
	if len(got) != 1 || got[0].FileIndex != 2 {
		t.Fatalf("limit 2 MiB: tracks = %+v, want only file 2", got)
	}
	if len(large) != 1 || large[0].FileIndex != 3 || large[0].Size != 8<<20 {
		t.Fatalf("limit 2 MiB: too large = %+v, want only file 3 (8 MiB)", large)
	}

	got, large = MatchLimit(files[0], files, 16<<20)
	if len(got) != 2 || len(large) != 0 {
		t.Fatalf("limit 16 MiB: tracks = %+v, too large = %+v, want 2 tracks and none too large", got, large)
	}
}

func TestMatchLimitIgnoresOtherEpisodes(t *testing.T) {
	// A big subtitle of another episode is not reported for this one.
	files := []FileInfo{
		{Index: 1, Path: "Show - 01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show - 02.mkv", Size: 1 << 30},
		{Index: 3, Path: "Show - 02.eng.ass", Size: 8 << 20},
	}
	got, large := MatchLimit(files[0], files, 2<<20)
	if len(got) != 0 || len(large) != 0 {
		t.Fatalf("tracks = %+v, too large = %+v, want none for another episode", got, large)
	}
}

func TestMatchDefaultLimit(t *testing.T) {
	// Subtitles of several megabytes (for example, full ASS files of a BDRip) are listed;
	// only files over MaxSubSize are ignored.
	files := []FileInfo{
		{Index: 1, Path: "Show/ep01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show/ep01.eng.ass", Size: 8 << 20},
		{Index: 3, Path: "Show/ep01.rus.ass", Size: MaxSubSize + 1},
	}
	got := Match(files[0], files)
	if len(got) != 1 || got[0].FileIndex != 2 || got[0].Size != 8<<20 {
		t.Fatalf("tracks = %+v, want only file 2 (8 MiB)", got)
	}
}
