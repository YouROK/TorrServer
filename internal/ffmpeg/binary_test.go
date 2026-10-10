package ffmpeg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in      string
		want    Version
		wantErr bool
	}{
		{in: "6.0", want: Version{6, 0, 0}},
		{in: "8.1.3", want: Version{8, 1, 3}},
		{in: "n6.1.1", want: Version{6, 1, 1}},
		{in: "4.4", want: Version{4, 4, 0}},
		{in: "7.0.1", want: Version{7, 0, 1}},
		{in: "", wantErr: true},
		{in: "n-109000-g1234567", wantErr: true},
		{in: "abc", wantErr: true},
	}

	for _, c := range cases {
		got, err := ParseVersion(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseVersion(%q): expected error, got %v", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseVersion(%q): unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseVersion(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestVersionAtLeast(t *testing.T) {
	cases := []struct {
		version Version
		min     string
		want    bool
	}{
		{Version{6, 0, 0}, "6.0", true},
		{Version{8, 1, 3}, "6.0", true},
		{Version{5, 1, 9}, "6.0", false},
		{Version{4, 4, 2}, "6.0", false},
		{Version{6, 1, 0}, "6.0", true},
		{Version{6, 0, 0}, "abc", true},
	}

	for _, c := range cases {
		if got := c.version.AtLeast(c.min); got != c.want {
			t.Errorf("%v.AtLeast(%q) = %v, want %v", c.version, c.min, got, c.want)
		}
	}
}

func TestParseVersionOutput(t *testing.T) {
	cases := []struct {
		name    string
		out     string
		want    Version
		wantStr string
		wantErr bool
	}{
		{
			name:    "jellyfin build",
			out:     "ffmpeg version 8.1.3-Jellyfin Copyright (c) 2000-2026\nbuilt with gcc",
			want:    Version{8, 1, 3},
			wantStr: "8.1.3",
		},
		{
			name:    "ubuntu build",
			out:     "ffmpeg version 8.0.1-3ubuntu2 Copyright (c) 2000-2025\nconfiguration: ...",
			want:    Version{8, 0, 1},
			wantStr: "8.0.1",
		},
		{
			name:    "release prefix",
			out:     "ffmpeg version n6.1.1 Copyright (c) 2000-2023\n",
			want:    Version{6, 1, 1},
			wantStr: "6.1.1",
		},
		{
			name:    "avconv",
			out:     "ffmpeg version 4.4.2\n",
			want:    Version{4, 4, 2},
			wantStr: "4.4.2",
		},
		{
			name:    "git build",
			out:     "ffmpeg version N-109000-g1234567890 Copyright (c) 2000-2022\n",
			wantErr: true,
		},
		{
			name:    "empty",
			out:     "",
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, gotStr, err := parseVersionOutput(c.out)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("version = %v, want %v", got, c.want)
			}
			if gotStr != c.wantStr {
				t.Errorf("version string = %q, want %q", gotStr, c.wantStr)
			}
		})
	}
}

func TestParseCodecList(t *testing.T) {
	out := `Encoders:
 V..... = Video
 ------
 V....D libx264              H.264 (codec h264)
 A....D aac                  AAC (codec aac)
 A....D libfdk_aac           FDK AAC (codec aac)
`

	got := parseCodecList(out)
	for _, name := range []string{"libx264", "aac", "libfdk_aac"} {
		if _, ok := got[name]; !ok {
			t.Errorf("encoder %q not parsed", name)
		}
	}
	if _, ok := got["Video"]; ok {
		t.Error("header line was parsed as an encoder")
	}
	if len(got) != 3 {
		t.Errorf("expected 3 encoders, got %d: %v", len(got), got)
	}
}

func TestParseFilterList(t *testing.T) {
	out := `Filters:
  T.. = Timeline support
  ... = Slice threading
  ------
 ... scale              V->V       Scale the input video size.
 ... aresample          A->A       Resample audio data.
`

	got := parseFilterList(out)
	if _, ok := got["scale"]; !ok {
		t.Error("filter scale not parsed")
	}
	if _, ok := got["aresample"]; !ok {
		t.Error("filter aresample not parsed")
	}
}

func TestParseHwaccelList(t *testing.T) {
	out := "Hardware acceleration methods:\ncuda\nvaapi\nqsv\n"

	got := parseHwaccelList(out)
	for _, name := range []string{"cuda", "vaapi", "qsv"} {
		if _, ok := got[name]; !ok {
			t.Errorf("hwaccel %q not parsed", name)
		}
	}
	if _, ok := got["Hardware"]; ok {
		t.Error("header line was parsed as a hwaccel")
	}
}

func TestFindCandidatesOrder(t *testing.T) {
	dir := t.TempDir()
	name := binaryName("ffmpeg")

	nearPath := filepath.Join(dir, name)
	if err := os.WriteFile(nearPath, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(dir, "custom-"+name)
	if err := os.WriteFile(cfgPath, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir)

	got := findCandidates(cfgPath)
	if len(got) == 0 {
		t.Fatal("no candidates found")
	}
	if got[0].source != sourceConfig {
		t.Errorf("first candidate source = %v, want %v", got[0].source, sourceConfig)
	}
}

func TestFindCandidatesSkipsMissing(t *testing.T) {
	got := findCandidates(filepath.Join(t.TempDir(), "missing-ffmpeg"))
	for _, c := range got {
		if c.source == sourceConfig {
			t.Errorf("missing config binary was offered as candidate: %s", c.path)
		}
	}
}

func TestExecutableDir(t *testing.T) {
	dir := executableDir()
	if dir == "" {
		t.Skip("executable dir is not available on this platform")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("executable dir %q is not absolute", dir)
	}
	if runtime.GOOS == "windows" && filepath.Ext(dir) != "" {
		t.Errorf("unexpected path %q", dir)
	}
}
