package subs

import (
	"fmt"
	"testing"
)

func TestMatchSingleVideo(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Show/ep01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show/ep01.rus.srt", Size: 50_000},
		{Index: 3, Path: "Show/ep01.eng.ass", Size: 60_000},
	}
	got := Match(files[0], files)
	if len(got) != 2 {
		t.Fatalf("want 2 tracks, got %d: %+v", len(got), got)
	}
	// sorted by language: en < ru
	if got[0].FileIndex != 3 || got[0].Lang != "en" || got[0].Format != "ass" {
		t.Errorf("unexpected first track: %+v", got[0])
	}
	if got[1].FileIndex != 2 || got[1].Lang != "ru" || got[1].Format != "srt" {
		t.Errorf("unexpected second track: %+v", got[1])
	}
}

func TestMatchEpisodesInSubfolder(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Show/Show - 01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show/Show - 02.mkv", Size: 1 << 30},
		{Index: 3, Path: "Show/Subs/Show - 01.rus.ass", Size: 40_000},
		{Index: 4, Path: "Show/Subs/Show - 02.rus.ass", Size: 40_000},
	}
	got := Match(files[1], files)
	if len(got) != 1 || got[0].FileIndex != 4 || got[0].Lang != "ru" {
		t.Fatalf("want only track 4 (ru), got %+v", got)
	}
}

func TestMatchNamePrefixEndsAtWord(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Show/Show - 1.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show/Show - 1.srt", Size: 50_000},
		{Index: 3, Path: "Show/Show - 10.mkv", Size: 1 << 30},
		{Index: 4, Path: "Show/Show - 10.srt", Size: 50_000},
		{Index: 5, Path: "Show/Show - 11.mkv", Size: 1 << 30},
		{Index: 6, Path: "Show/Show - 11.rus.srt", Size: 50_000},
	}
	for i, want := range map[int]int{0: 2, 2: 4, 4: 6} {
		got := Match(files[i], files)
		if len(got) != 1 || got[0].FileIndex != want {
			t.Errorf("%s: want only subtitle %d, got %+v", files[i].Path, want, got)
		}
	}

	files = []FileInfo{
		{Index: 1, Path: "Alien.mkv", Size: 1 << 30},
		{Index: 2, Path: "Aliens.mkv", Size: 1 << 30},
		{Index: 3, Path: "Alien.eng.srt", Size: 50_000},
		{Index: 4, Path: "Aliens.eng.srt", Size: 50_000},
	}
	if got := Match(files[0], files); len(got) != 1 || got[0].FileIndex != 3 {
		t.Errorf("Alien.mkv: want only subtitle 3, got %+v", got)
	}
}

func TestMatchSkipsHugeFiles(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "movie.mkv", Size: 1 << 30},
		{Index: 2, Path: "movie.srt", Size: MaxSubSize + 1},
	}
	if got := Match(files[0], files); len(got) != 0 {
		t.Fatalf("want no tracks, got %+v", got)
	}
}

// A real-world batch: the codec name "x.265" after the episode number must not be
// taken for the episode, otherwise every video "matches" every subtitle file.
func TestMatchBatchWithCodecInName(t *testing.T) {
	const suffix = " (BD 1920x1080 x.265-10Bit Flac)"
	var files []FileInfo
	var videos []FileInfo
	for ep := 1; ep <= 12; ep++ {
		name := fmt.Sprintf("[Moozzi2] Kandagawa Jet Girls - %02d", ep)
		if ep == 12 {
			name += " END" // the video of the last episode has an extra word
		}
		v := FileInfo{Index: ep, Path: name + suffix + ".mkv", Size: 1 << 30}
		videos = append(videos, v)
		files = append(files, v)
	}
	for ep := 1; ep <= 12; ep++ {
		name := fmt.Sprintf("[Moozzi2] Kandagawa Jet Girls - %02d", ep)
		files = append(files, FileInfo{
			Index: 12 + ep,
			Path:  "Ru Subs/TiredGirl & ElSwanko/" + name + suffix + ".[TiredGirl & ElSwanko].ass",
			Size:  30_000,
		})
	}

	for ep, v := range videos {
		got := Match(v, files)
		if len(got) != 1 || got[0].FileIndex != 12+ep+1 {
			t.Errorf("episode %d: want only subtitle %d, got %+v", ep+1, 12+ep+1, got)
		}
	}
}

// If the extracted number is the same for several videos, it cannot tell them apart:
// only the file name prefix is used.
func TestMatchIgnoresAmbiguousEpisodeNumbers(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Alpha - 10.mkv", Size: 1 << 30},
		{Index: 2, Path: "Beta - 10.mkv", Size: 1 << 30},
		{Index: 3, Path: "Alpha - 10.rus.srt", Size: 10_000},
		{Index: 4, Path: "Beta - 10.rus.srt", Size: 10_000},
	}
	got := Match(files[0], files)
	if len(got) != 1 || got[0].FileIndex != 3 {
		t.Fatalf("want only subtitle 3, got %+v", got)
	}
}

// Subtitle files lying among audio files (a voice-over folder) are not subtitle tracks
// of the video, but a flat folder with the video, audio and subtitles keeps its subtitles.
func TestMatchSkipsSubtitlesInAudioOnlyFolders(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Show - 01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Sound [Dub]/Show - 01.mka", Size: 30 << 20},
		{Index: 3, Path: "Sound [Dub]/Show - 01.ass", Size: 20_000},
		{Index: 4, Path: "Subs [Group]/Show - 01.ass", Size: 40_000},
	}
	got := Match(files[0], files)
	if len(got) != 1 || got[0].FileIndex != 4 {
		t.Fatalf("want only subtitle 4, got %+v", got)
	}

	flat := []FileInfo{
		{Index: 1, Path: "Movie/Movie.mkv", Size: 1 << 30},
		{Index: 2, Path: "Movie/Movie.rus.ac3", Size: 100 << 20},
		{Index: 3, Path: "Movie/Movie.rus.srt", Size: 90_000},
	}
	if got := Match(flat[0], flat); len(got) != 1 || got[0].FileIndex != 3 {
		t.Fatalf("flat folder: want subtitle 3, got %+v", got)
	}
}

// Tracks with the same file name in different folders get the folder name as a prefix.
func TestTrackTitles(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Show 01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Subs [DJATOM & Jarly]/Show 01.ass", Size: 40_000},
		{Index: 3, Path: "Subs [Exmel & Reddog]/Show 01.ass", Size: 40_000},
		{Index: 4, Path: "Subs [Exmel & Reddog]/Show 01.rus.srt", Size: 40_000},
	}
	got := Match(files[0], files)
	titles := map[int]string{}
	for _, tr := range got {
		titles[tr.FileIndex] = tr.Title
		if tr.Name != "Show 01.ass" && tr.Name != "Show 01.rus.srt" {
			t.Errorf("Name must stay the file name, got %q", tr.Name)
		}
	}
	want := map[int]string{
		2: "DJATOM & Jarly - Show 01.ass",
		3: "Exmel & Reddog - Show 01.ass",
		4: "Show 01.rus.srt", // the name is unique: no prefix
	}
	for idx, w := range want {
		if titles[idx] != w {
			t.Errorf("title of track %d = %q, want %q", idx, titles[idx], w)
		}
	}
}

func TestDirLabel(t *testing.T) {
	cases := map[string]string{
		"Show/Subs [Exmel & Reddog]": "Exmel & Reddog",
		"Show/Ru Subs":               "Ru",
		"Show/Subs":                  "",
		"Show/Субтитры":              "",
		"Show/[Fansub] Rus":          "Fansub Rus",
		"Show/Sub_Rus":               "Rus",
		"Show/Subs - Exmel":          "Exmel",
		".":                          "",
	}
	for dir, want := range cases {
		if got := dirLabel(dir); got != want {
			t.Errorf("dirLabel(%q) = %q, want %q", dir, got, want)
		}
	}
}
