// Package offline implements the on-disk storage used by TorrServer's
// "offline mode": a torrent is downloaded completely into ordinary files
// (<base>/<infohash>/<torrent name>/...) and can later be served from them
// without touching the BitTorrent network.
package offline

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

// File describes one file of a torrent.
type File struct {
	// Path is the slash separated path exactly as torrent.File.Path() reports it
	// (torrent name + the path from the metainfo), so ids and paths in the status
	// of an offline torrent match the ones of a live torrent.
	Path string
	// Disk is the absolute location of the file on disk.
	Disk string
	// Offset of the file inside the torrent data stream.
	Offset int64
	Length int64
}

// TorrentRoot returns the directory holding all data of one torrent.
func TorrentRoot(base string, hash metainfo.Hash) string {
	return filepath.Join(base, hash.HexString())
}

// windowsReservedNames are DOS device names that cannot be created on Windows,
// not even with an extension ("aux.srt" is rejected just like "aux").
var windowsReservedNames = map[string]struct{}{
	"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
	"COM1": {}, "COM2": {}, "COM3": {}, "COM4": {}, "COM5": {},
	"COM6": {}, "COM7": {}, "COM8": {}, "COM9": {},
	"LPT1": {}, "LPT2": {}, "LPT3": {}, "LPT4": {}, "LPT5": {},
	"LPT6": {}, "LPT7": {}, "LPT8": {}, "LPT9": {},
}

// safeComponent makes one path element safe to use on disk (no separators,
// no characters that Windows rejects, no "." / "..").
//
// Two extra cases are handled explicitly (review remark V-2):
//
//   - different components can sanitize to the same string ("a/b" and "a\\b"
//     both become "a_b"): when sanitization loses information, a short hash of
//     the original component is appended ("a_b#1b2e5c7d"), so two different
//     torrent files never map onto the same disk file. The hash is
//     deterministic, the layout stays stable across restarts.
//
//   - Windows-reserved names are rejected even with an extension; a suffix is
//     inserted before the extension ("aux.srt" -> "aux#1b2e5c7d.srt").
//
// Only the on-disk path (Disk) is affected: the torrent-relative Path keeps
// the original component, so file ids and names in the web UI do not change.
func safeComponent(s string) string {
	orig := s
	mapped := strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
			return '_'
		}
		if r < 32 {
			return '_'
		}
		return r
	}, s)
	trimmed := strings.TrimRight(mapped, ". ")
	// Information was lost while sanitizing: a hash of the original component
	// disambiguates the result, so "a/b", "a\\b" and "a:b" (all mapped to
	// "a_b"), "a." and "a" (trailing dot trimmed) stay different files. The
	// '#' check keeps a literal "#..." component from colliding with a
	// suffixed one.
	lossy := mapped != orig || trimmed != mapped || strings.Contains(orig, "#")
	if lossy {
		trimmed += "#" + strconv.FormatUint(uint64(fnv32a(orig)), 16)
	}
	s = trimmed
	// Windows takes everything before the FIRST dot for the device name:
	// "con.tar.gz" is as reserved as "con.gz"
	stem, rest := s, ""
	if i := strings.IndexByte(s, '.'); i > 0 {
		stem, rest = s[:i], s[i:]
	}
	if windowsReserved(stem) {
		s = stem + "#" + strconv.FormatUint(uint64(fnv32a(orig)), 16) + rest
	}
	if s == "" {
		return "_"
	}
	return s
}

// windowsReserved reports whether base is a Windows-reserved device name
// (case-insensitively: "Aux", "AUX" and "aux" are all rejected by Windows).
func windowsReserved(base string) bool {
	_, ok := windowsReservedNames[strings.ToUpper(base)]
	return ok
}

// fnv32a is a small deterministic hash (FNV-1a, 32 bit) used to disambiguate
// sanitized path components.
func fnv32a(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// Files lists the files of the torrent in metainfo order.
func Files(base string, hash metainfo.Hash, info *metainfo.Info) []File {
	root := TorrentRoot(base, hash)
	name := info.BestName()
	var (
		off int64
		out []File
	)
	for _, fi := range info.UpvertedFiles() {
		comps := append([]string{name}, fi.BestPath()...)
		parts := make([]string, 0, len(comps)+1)
		parts = append(parts, root)
		for _, c := range comps {
			parts = append(parts, safeComponent(c))
		}
		out = append(out, File{
			Path:   strings.Join(comps, "/"),
			Disk:   filepath.Join(parts...),
			Offset: off,
			Length: fi.Length,
		})
		off += fi.Length
	}
	return out
}

// LoadInfo parses the info dictionary stored in TorrentSpec.InfoBytes.
func LoadInfo(infoBytes []byte) (*metainfo.Info, error) {
	mi := metainfo.MetaInfo{InfoBytes: infoBytes}
	info, err := mi.UnmarshalInfo()
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// FilesIntact reports whether every non-empty file exists with the expected size.
func FilesIntact(base string, hash metainfo.Hash, info *metainfo.Info) bool {
	for _, f := range Files(base, hash, info) {
		if f.Length == 0 {
			continue
		}
		st, err := os.Stat(f.Disk)
		if err != nil || st.IsDir() || st.Size() != f.Length {
			return false
		}
	}
	return true
}

// RemoveData deletes everything stored for the torrent.
func RemoveData(base string, hash metainfo.Hash) error {
	if base == "" || hash.HexString() == "" {
		return nil
	}
	_ = os.Remove(sidecarPath(base, hash))
	root := TorrentRoot(base, hash)
	if _, err := os.Stat(root); err != nil {
		return nil
	}
	return os.RemoveAll(root)
}
