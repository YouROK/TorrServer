package offline

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
)

func testHash() metainfo.Hash {
	return metainfo.NewHashFromHex("0123456789abcdef0123456789abcdef01234567")
}

func suffix(orig string) string {
	return "#" + strconv.FormatUint(uint64(fnv32a(orig)), 16)
}

func TestSafeComponent(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain name is kept", "Movie 2019.mkv", "Movie 2019.mkv"},
		{"separators and forbidden chars", `a<b>c:d"e/f\g|h?i*j`, "a_b_c_d_e_f_g_h_i_j" + suffix(`a<b>c:d"e/f\g|h?i*j`)},
		{"control characters", "a\x01b", "a_b" + suffix("a\x01b")},
		{"trailing dots and spaces are trimmed", "name. ", "name" + suffix("name. ")},
		{"empty", "", "_"},
		{"only dots does not collapse into _", "...", suffix("...")},
		{"dot is distinct from dotdot", ".", suffix(".")},
		{"hash suffix is deterministic", "a/b", "a_b" + suffix("a/b")},
		{"backslash is distinct from slash", `a\b`, "a_b" + suffix(`a\b`)},
		{"colon is distinct from slash", "a:b", "a_b" + suffix("a:b")},
		{"literal underscore without loss is untouched", "a_b", "a_b"},
		{"literal hash gets a suffix too", "a#b", "a#b" + suffix("a#b")},
		{"reserved name without extension", "aux", "aux" + suffix("aux")},
		{"reserved name with extension", "aux.srt", "aux" + suffix("aux.srt") + ".srt"},
		{"reserved name is case-insensitive", "CON.mkv", "CON" + suffix("CON.mkv") + ".mkv"},
		{"reserved name NUL", "nul", "nul" + suffix("nul")},
		{"reserved name COM1", "com1", "com1" + suffix("com1")},
		{"longer names are not reserved", "config", "config"},
		{"longer names with extension are not reserved", "controller.mkv", "controller.mkv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeComponent(tt.in); got != tt.want {
				t.Fatalf("safeComponent(%q) = %q, want %q", tt.in, got, tt.want)
			}
			// deterministic: the layout must be stable across restarts
			if again := safeComponent(tt.in); again != tt.want {
				t.Fatalf("safeComponent(%q) is not deterministic: %q vs %q", tt.in, again, tt.want)
			}
		})
	}
}

// TestSafeComponentNoCollisions feeds components that sanitize to the same
// string into safeComponent and checks that the results are all different:
// two torrent files must never map onto one disk file (review remark V-2).
func TestSafeComponentNoCollisions(t *testing.T) {
	groups := [][]string{
		{"a/b", `a\b`, "a:b", "a?b", "a*b", "a|b", "a<b", "a>b", `a"b`},
		{"a", "a.", "a..", "a. .", "a  "},
		{".", "..", "...", "...."},
		{"aux", "aux.srt", "AUX.txt", "con", "con.mkv", "COM1", "com1.dat"},
	}
	for i, group := range groups {
		seen := make(map[string]string)
		for _, orig := range group {
			safe := safeComponent(orig)
			if prev, ok := seen[safe]; ok {
				t.Errorf("group %d: safeComponent(%q) and safeComponent(%q) collide on %q", i, prev, orig, safe)
			}
			seen[safe] = orig
		}
	}
}

func TestFilesPaths(t *testing.T) {
	base := t.TempDir()
	ih := testHash()
	info := &metainfo.Info{
		Name:        "Show",
		PieceLength: 16384,
		Pieces:      make([]byte, 20),
		Files: []metainfo.FileInfo{
			{Path: []string{"S01", "e1.mkv"}, Length: 100},
			{Path: []string{"S01", "e2.mkv"}, Length: 200},
		},
	}
	files := Files(base, ih, info)
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	want := []File{
		{Path: "Show/S01/e1.mkv", Disk: filepath.Join(TorrentRoot(base, ih), "Show", "S01", "e1.mkv"), Offset: 0, Length: 100},
		{Path: "Show/S01/e2.mkv", Disk: filepath.Join(TorrentRoot(base, ih), "Show", "S01", "e2.mkv"), Offset: 100, Length: 200},
	}
	for i, f := range files {
		if f.Path != want[i].Path {
			t.Errorf("files[%d].Path = %q, want %q (must match torrent.File.Path())", i, f.Path, want[i].Path)
		}
		if f.Disk != want[i].Disk {
			t.Errorf("files[%d].Disk = %q, want %q", i, f.Disk, want[i].Disk)
		}
		if f.Offset != want[i].Offset || f.Length != want[i].Length {
			t.Errorf("files[%d] = %+v, want %+v", i, f, want[i])
		}
	}
}

func TestFilesIntact(t *testing.T) {
	base := t.TempDir()
	ih := testHash()
	info := &metainfo.Info{
		Name:        "Show",
		PieceLength: 16384,
		Pieces:      make([]byte, 20),
		Files: []metainfo.FileInfo{
			{Path: []string{"movie.mkv"}, Length: 400},
			{Path: []string{"empty.nfo"}, Length: 0},
		},
	}
	movie := filepath.Join(TorrentRoot(base, ih), "Show", "movie.mkv")
	if err := os.MkdirAll(filepath.Dir(movie), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(movie, make([]byte, 400), 0o644); err != nil {
		t.Fatal(err)
	}
	// a missing zero-length file is not an integrity failure
	if !FilesIntact(base, ih, info) {
		t.Fatal("a torrent whose only missing file is empty must be intact")
	}
	if err := os.WriteFile(movie, make([]byte, 399), 0o644); err != nil {
		t.Fatal(err)
	}
	if FilesIntact(base, ih, info) {
		t.Fatal("a file with an unexpected size must not be intact")
	}
	if err := os.WriteFile(movie, make([]byte, 400), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(movie); err != nil {
		t.Fatal(err)
	}
	if FilesIntact(base, ih, info) {
		t.Fatal("a missing file must not be intact")
	}
}

func TestRemoveData(t *testing.T) {
	base := t.TempDir()
	ih := testHash()
	root := TorrentRoot(base, ih)
	if err := os.MkdirAll(filepath.Join(root, "Show"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Show", "movie.mkv"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sidecarPath(base, ih), []byte("TSOFF1\n11"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RemoveData(base, ih); err != nil {
		t.Fatalf("RemoveData: %v", err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("torrent root still exists after RemoveData: %v", err)
	}
	if _, err := os.Stat(sidecarPath(base, ih)); !os.IsNotExist(err) {
		t.Fatalf("sidecar still exists after RemoveData: %v", err)
	}
	// removing again must not fail
	if err := RemoveData(base, ih); err != nil {
		t.Fatalf("second RemoveData: %v", err)
	}
}

func TestSafeComponentWindowsReservedMultiDot(t *testing.T) {
	for _, name := range []string{"con.tar.gz", "AUX.backup.srt", "nul.a.b.c"} {
		got := safeComponent(name)
		if got == name {
			t.Errorf("safeComponent(%q) was not changed", name)
			continue
		}
		stem := got
		if i := strings.IndexByte(got, '.'); i > 0 {
			stem = got[:i]
		}
		if windowsReserved(stem) {
			t.Errorf("safeComponent(%q) = %q is still a reserved Windows name", name, got)
		}
	}
}
