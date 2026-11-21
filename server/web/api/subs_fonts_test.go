package api

import (
	"reflect"
	"testing"
)

func paths(files []torrentFile) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.Path)
	}
	return out
}

func TestNearestFonts(t *testing.T) {
	files := []torrentFile{
		{ID: 1, Path: "Show/S01/E01.mkv"},
		{ID: 2, Path: "Show/S01/Fonts/B.otf"},
		{ID: 3, Path: "Show/S01/Fonts/A.TTF"},
		{ID: 4, Path: "Show/S01/E01.ass"},
		{ID: 5, Path: "Show/S02/Fonts/C.ttf"},
		{ID: 6, Path: "Show/Fonts/D.ttc"},
		{ID: 7, Path: "Show/Fonts/E.woff2"},
	}
	cases := []struct {
		video string
		want  []string
	}{
		{"Show/S01/E01.mkv", []string{"Show/S01/Fonts/A.TTF", "Show/S01/Fonts/B.otf"}}, // below the folder of the video
		// no fonts below S03: the folder above with its subfolders (the other seasons too)
		{"Show/S03/E01.mkv", []string{"Show/Fonts/D.ttc", "Show/S01/Fonts/A.TTF", "Show/S01/Fonts/B.otf", "Show/S02/Fonts/C.ttf"}},
		{"Show\\S01\\E01.mkv", []string{"Show/S01/Fonts/A.TTF", "Show/S01/Fonts/B.otf"}}, // Windows separators
		{"E01.mkv", []string{"Show/Fonts/D.ttc", "Show/S01/Fonts/A.TTF", "Show/S01/Fonts/B.otf", "Show/S02/Fonts/C.ttf"}},
	}
	for _, tc := range cases {
		if got := paths(nearestFonts(tc.video, files)); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.video, got, tc.want)
		}
	}
	if got := nearestFonts("Show/S01/E01.mkv", files[:1]); len(got) != 0 {
		t.Errorf("no fonts in the torrent: got %v", got)
	}
}

func TestExternalFontURL(t *testing.T) {
	got := externalFontURL("abc", torrentFile{ID: 3, Path: "Show/Fonts/My Font.ttf"})
	want := "/stream/My%20Font.ttf?link=abc&index=3&play"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
