package extfiles

import "testing"

func TestLangDetection(t *testing.T) {
	cases := []struct{ path, want string }{
		// Explicit markers.
		{"Show/Russian subs/ep01.srt", "ru"},
		{"Show/ep01.eng.srt", "en"},
		{"Show/Subs/English/ep01.srt", "en"},

		// No marker and no Cyrillic: unknown.
		{"Show/ep01.srt", "und"},

		// Cyrillic in a folder or file name, no explicit marker: Russian.
		{"Субтитры/ep01.srt", "ru"},
		{"Show/Субтитры/Show - 01.srt", "ru"},
		{"Сериал/Subs/ep01.srt", "ru"},
		{"Show/ep01.субтитры.srt", "ru"},

		// An explicit marker wins over Cyrillic, wherever it is.
		{"Субтитры/en/ep01.srt", "en"},
		{"Субтитры/ep01.eng.srt", "en"},
		{"Субтитры/Английские/ep01.srt", "en"},
		{"English/Субтитры/ep01.srt", "en"},
		{"Сериал.ENG/Subs/ep01.srt", "en"},
	}
	for _, c := range cases {
		if got := LangOf(c.path); got != c.want {
			t.Errorf("LangOf(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestIsVideo(t *testing.T) {
	cases := map[string]bool{
		"Show/ep01.mkv":     true,
		"Show/EP01.MP4":     true,
		"Show/ep01.rus.srt": false,
		"Show/readme.txt":   false,
	}
	for p, want := range cases {
		if got := IsVideo(p); got != want {
			t.Errorf("IsVideo(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestIsAudio(t *testing.T) {
	for p, want := range map[string]bool{
		"a/ep01.mka": true, "ep01.FLAC": true, "x.ac3": true,
		"ep01.mkv": false, "ep01.ass": false, "noext": false,
	} {
		if got := IsAudio(p); got != want {
			t.Errorf("IsAudio(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestEpisodeOfIgnoresCodecsAndDecimals(t *testing.T) {
	cases := map[string]int{
		"[Moozzi2] Kandagawa Jet Girls - 03 (BD 1920x1080 x.265-10Bit Flac)": 3,
		"Show - 04 [BDRip 1920x1080 x264 FLAC 2.0]":                           4,
		"Show - 05 (H.264 5.1)":                                              5,
		"Mob Psycho 100 - 02":                                                2,
	}
	for name, want := range cases {
		if _, got, ok := episodeOf(name); !ok || got != want {
			t.Errorf("episodeOf(%q) = %d (ok=%v), want %d", name, got, ok, want)
		}
	}
}

// DirLabel with the real set of audio-folder words, so that a change of audioDirWords
// that spoils the labels of the tracks is noticed. (The subtitle words are checked by
// TestDirLabel in package subs.)
func TestDirLabelAudioFolders(t *testing.T) {
	cases := map[string]string{
		"Show/Sound [Group & Co]": "Group & Co",
		"Show/Sound - Group":      "Group",
		"Show/Озвучка [Группа]":   "Группа",
		"Show/Dub Rus":            "Rus",
		"Show/[Fansub] Rus":       "Fansub Rus",
		"Show/Audio":              "",
		"Show/Sound_Dub":          "",
		".":                       "",
	}
	for dir, want := range cases {
		if got := DirLabel(dir, audioDirWords); got != want {
			t.Errorf("DirLabel(%q, audioDirWords) = %q, want %q", dir, got, want)
		}
	}
}
