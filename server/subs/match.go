// Package subs finds external subtitle files in a torrent and converts them to WebVTT.
// It uses the standard library only. Which files of a torrent belong to a video is decided
// by package extfiles.
package subs

import (
	"path"
	"sort"
	"strings"

	"server/extfiles"
)

// FileInfo describes a torrent file; see extfiles.FileInfo.
type FileInfo = extfiles.FileInfo

// Track is an external subtitle track found for a particular video file.
type Track struct {
	FileIndex int    `json:"index"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Lang      string `json:"lang"`   // "ru", "en", ... or "und"
	Format    string `json:"format"` // "srt", "ass", "ssa", "vtt"

	dir string // cleaned name of the folder, see dirLabel
}

// Subtitle files are small (kilobytes); anything larger is ignored.
const MaxSubSize = 2 << 20

var subExt = map[string]bool{".srt": true, ".ass": true, ".ssa": true, ".vtt": true}

// dirStopWords are folder-name words that carry no information: "Subs [Group]" -> "Group".
var dirStopWords = map[string]bool{
	"subs": true, "sub": true, "subtitles": true, "subtitle": true, "субтитры": true, "субтитр": true,
}

// Match returns the subtitle tracks that belong to the given video file.
func Match(video FileInfo, files []FileInfo) []Track {
	v := extfiles.NewVideo(video, files)

	var out []Track
	for _, f := range files {
		ext := strings.ToLower(path.Ext(f.Path))
		if !subExt[ext] || f.Size > MaxSubSize {
			continue
		}
		// Subtitle files among audio files, with no video in the same folder, are
		// companions of that audio (for example "signs" for a voice-over), not subtitle
		// tracks of the video.
		if v.InAudioOnlyDir(f.Path) {
			continue
		}
		// One video in the torrent: all subtitles belong to it. Several videos: match by name or episode.
		if !v.Belongs(f.Path) {
			continue
		}
		out = append(out, Track{
			FileIndex: f.Index,
			Name:      path.Base(f.Path),
			Lang:      extfiles.LangOf(f.Path),
			Format:    strings.TrimPrefix(ext, "."),
			dir:       dirLabel(path.Dir(f.Path)),
		})
	}

	// Title is the file name; when several tracks share a file name, the cleaned
	// folder name is prepended so that the tracks can be told apart.
	names := map[string]int{}
	for _, t := range out {
		names[t.Name]++
	}
	for i := range out {
		out[i].Title = out[i].Name
		if names[out[i].Name] > 1 && out[i].dir != "" {
			out[i].Title = out[i].dir + " - " + out[i].Name
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Lang != out[j].Lang {
			return out[i].Lang < out[j].Lang
		}
		return out[i].Title < out[j].Title
	})
	return out
}

// dirLabel turns the name of a folder into a short label: square and curly brackets and
// underscores become spaces, and the words "Subs", "Sub", "Subtitles" and "Субтитры" are
// dropped: "Subs [Exmel & Reddog]" -> "Exmel & Reddog". It returns "" when nothing is left.
func dirLabel(dir string) string {
	return extfiles.DirLabel(dir, dirStopWords)
}
