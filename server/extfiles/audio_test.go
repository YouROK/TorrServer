package extfiles

import (
	"fmt"
	"testing"
)

// yamadaFiles is the layout of a real torrent: the episodes lie in the root, the voice-overs
// (.mka with companion signs .ass) in two "Sound [...]" folders, full subtitles in two
// "Subs [...]" folders, creditless OP/ED in Bonus.
func yamadaFiles() []FileInfo {
	var files []FileInfo
	add := func(path string, size int64) {
		files = append(files, FileInfo{Index: len(files) + 1, Path: path, Size: size})
	}
	name := func(n int) string {
		return fmt.Sprintf("[Beatrice-Raws] Yamada-kun to 7-nin no Majo %02d [BDRip 1920x1080 x264 FLAC]", n)
	}
	for n := 1; n <= 12; n++ {
		add(name(n)+".mkv", 1<<30)
	}
	add("Bonus/[Beatrice-Raws] Yamada-kun to 7-nin no Majo (Creditless ED) [BDRip 1920x1080 x264 FLAC].mkv", 121_350_132)
	add("Bonus/[Beatrice-Raws] Yamada-kun to 7-nin no Majo (Creditless OP) [BDRip 1920x1080 x264 FLAC].mkv", 78_030_761)
	add("Bonus/Scans/01.png", 9_000_000)
	for _, dir := range []string{"Sound [JAM & Nika Lenina]", "Sound [Shuuma & Gomer & Aska & Ruri & NesTea]"} {
		for n := 1; n <= 12; n++ {
			add(dir+"/"+name(n)+".ass", 5_000)
			add(dir+"/"+name(n)+".mka", 35_000_000)
		}
	}
	for _, dir := range []string{"Subs [DJATOM & Jarly & kievbear]", "Subs [Exmel & Reddog]"} {
		for n := 1; n <= 5; n++ {
			add(dir+"/"+name(n)+".ass", 45_000)
		}
	}
	return files
}

func TestMatchAudioVoiceOverFolders(t *testing.T) {
	files := yamadaFiles()
	for _, episode := range []int{1, 7, 12} { // 7 is also the number in "to 7-nin"
		got := MatchAudio(files[episode-1], files)
		if len(got) != 2 {
			t.Fatalf("episode %d: want 2 audio tracks, got %d: %+v", episode, len(got), got)
		}
		for _, track := range got {
			if track.Format != "mka" || track.Size != 35_000_000 {
				t.Errorf("episode %d: unexpected track %+v", episode, track)
			}
		}
		prefix := fmt.Sprintf("[Beatrice-Raws] Yamada-kun to 7-nin no Majo %02d [BDRip 1920x1080 x264 FLAC].mka", episode)
		if got[0].Title != "JAM & Nika Lenina - "+prefix {
			t.Errorf("episode %d: first title = %q", episode, got[0].Title)
		}
		if got[1].Title != "Shuuma & Gomer & Aska & Ruri & NesTea - "+prefix {
			t.Errorf("episode %d: second title = %q", episode, got[1].Title)
		}
	}
}

func TestMatchAudioCreditlessVideosGetNothing(t *testing.T) {
	files := yamadaFiles()
	if got := MatchAudio(files[12], files); len(got) != 0 { // Creditless ED
		t.Fatalf("want no tracks for a bonus video, got %+v", got)
	}
}

func TestMatchAudioSingleVideo(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Movie/Movie.mkv", Size: 4 << 30},
		{Index: 2, Path: "Movie/OST/01 Theme.flac", Size: 30_000_000},
		{Index: 3, Path: "Movie/Audio/Rus/dub.ac3", Size: 200_000_000},
		{Index: 4, Path: "Movie/Movie.Eng.ac3", Size: 200_000_000},
		{Index: 5, Path: "Movie/Commentary.mka", Size: 100_000_000},
		{Index: 6, Path: "Movie/Music/cover.mp3", Size: 5_000_000},
	}
	got := MatchAudio(files[0], files)
	var indexes []int
	for _, track := range got {
		indexes = append(indexes, track.FileIndex)
	}
	// sorted by language: en < ru < und; the soundtrack and the music folder are not dubbing
	want := []int{4, 3, 5}
	if fmt.Sprint(indexes) != fmt.Sprint(want) {
		t.Fatalf("got tracks %v, want %v: %+v", indexes, want, got)
	}
	if got[0].Lang != "en" || got[1].Lang != "ru" || got[2].Lang != "und" {
		t.Errorf("unexpected languages: %+v", got)
	}
}

func TestMatchAudioSeveralVideosByEpisode(t *testing.T) {
	files := []FileInfo{
		{Index: 1, Path: "Show/S01E01.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show/S01E02.mkv", Size: 1 << 30},
		{Index: 3, Path: "Show/Audio/Show.S01E01.rus.mka", Size: 30_000_000},
		{Index: 4, Path: "Show/Audio/Show.S01E02.rus.mka", Size: 30_000_000},
		{Index: 5, Path: "Show/Audio/commentary.mka", Size: 30_000_000},
	}
	got := MatchAudio(files[1], files)
	if len(got) != 1 || got[0].FileIndex != 4 || got[0].Lang != "ru" {
		t.Fatalf("want only track 4 (ru), got %+v", got)
	}
}
