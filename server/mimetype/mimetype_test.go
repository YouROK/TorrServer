package mimetype

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMimeTypePredicates(t *testing.T) {
	cases := []struct {
		mt                mimeType
		video, audio, img bool
		sub, media        bool
		kind              string
	}{
		{"video/mp4", true, false, false, false, true, "video"},
		{"application/vnd.rn-realmedia-vbr", true, false, false, false, true, "application"},
		{"audio/mpeg", false, true, false, false, true, "audio"},
		{"image/png", false, false, true, false, true, "image"},
		{"text/srt", false, false, false, true, false, "text"},
		{"text/smi; charset=utf-8", false, false, false, true, false, "text"},
		{"text/plain", false, false, false, false, false, "text"},
		{"", false, false, false, false, false, ""},
	}
	for _, tc := range cases {
		if tc.mt.IsVideo() != tc.video || tc.mt.IsAudio() != tc.audio || tc.mt.IsImage() != tc.img || tc.mt.IsSub() != tc.sub || tc.mt.IsMedia() != tc.media {
			t.Errorf("%q predicates video=%v audio=%v image=%v sub=%v media=%v", tc.mt, tc.mt.IsVideo(), tc.mt.IsAudio(), tc.mt.IsImage(), tc.mt.IsSub(), tc.mt.IsMedia())
		}
		if tc.mt.Type() != tc.kind || tc.mt.String() != string(tc.mt) {
			t.Errorf("%q type %q", tc.mt, tc.mt.Type())
		}
	}
}

func TestMimeTypeByPath(t *testing.T) {
	dir := t.TempDir()
	png := filepath.Join(dir, "pic")
	if err := os.WriteFile(png, []byte("\x89PNG\r\n\x1a\nrest"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := MimeTypeByPath(png)
	if err != nil || !strings.HasPrefix(string(got), "image/png") {
		t.Fatalf("sniff = %q %v", got, err)
	}

	mkv, err := MimeTypeByPath(filepath.Join(dir, "movie.mkv.part"))
	if err != nil || mkv != "video/x-matroska" {
		t.Fatalf("mkv = %q %v", mkv, err)
	}
	avi, err := MimeTypeByPath("clip.avi")
	if err != nil || avi != "video/avi" {
		t.Fatalf("avi = %q %v", avi, err)
	}
	ts, err := MimeTypeByPath("stream.ts")
	if err != nil || ts != "video/mpeg" {
		t.Fatalf("ts = %q %v", ts, err)
	}

	missing, err := MimeTypeByPath(filepath.Join(dir, "no-ext"))
	if err == nil || missing != "application/octet-stream" {
		t.Fatalf("missing = %q %v", missing, err)
	}
}
