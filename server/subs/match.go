// Package subs finds external subtitle files in a torrent and converts them to WebVTT.
// It uses the standard library only.
package subs

import (
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// FileInfo describes a torrent file. Index must match the file index
// used by /play/:hash/:id.
type FileInfo struct {
	Index int
	Path  string // path inside the torrent, e.g. "Show/Subs/Rus/ep01.ass"
	Size  int64
}

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

var (
	subExt   = map[string]bool{".srt": true, ".ass": true, ".ssa": true, ".vtt": true}
	videoExt = map[string]bool{".mkv": true, ".mp4": true, ".avi": true, ".m4v": true, ".ts": true, ".mov": true, ".webm": true}
	audioExt = map[string]bool{".mka": true, ".aac": true, ".ac3": true, ".eac3": true, ".dts": true, ".flac": true, ".mp3": true, ".ogg": true, ".opus": true, ".wav": true, ".m4a": true, ".thd": true}

	reSE  = regexp.MustCompile(`(?i)s(\d{1,2})[ ._-]?e(\d{1,3})`)
	reX   = regexp.MustCompile(`(?i)(?:^|[^0-9a-z])(\d{1,2})x(\d{1,3})(?:[^0-9a-z]|$)`)
	reNum = regexp.MustCompile(`(?:^|[\s._\[(-])(\d{1,3})(?:[\s._\])-]|$)`)

	// Numbers that are not episode numbers: codec names ("x.265", "H.264") and decimals ("5.1", "2.0").
	reCodecBefore   = regexp.MustCompile(`(?i)[xh]\.$`)
	reDecimalBefore = regexp.MustCompile(`\d\.$`)
	reDecimalAfter  = regexp.MustCompile(`^\.\d`)
)

// dirStopWords are folder-name words that carry no information: "Subs [Group]" -> "Group".
var dirStopWords = map[string]bool{
	"subs": true, "sub": true, "subtitles": true, "subtitle": true, "субтитры": true, "субтитр": true,
}

var langTokens = map[string]string{
	"ru": "ru", "rus": "ru", "russian": "ru", "рус": "ru", "русский": "ru", "русские": "ru",
	"en": "en", "eng": "en", "english": "en", "англ": "en", "английский": "en", "английские": "en",
	"ja": "ja", "jpn": "ja", "jap": "ja", "japanese": "ja", "японский": "ja", "японские": "ja",
}

// IsVideo reports whether the file looks like a video (by extension).
func IsVideo(p string) bool {
	return videoExt[strings.ToLower(path.Ext(p))]
}

// Match returns the subtitle tracks that belong to the given video file.
func Match(video FileInfo, files []FileInfo) []Track {
	videoDir := path.Dir(video.Path)
	videoBase := stripExt(path.Base(video.Path))

	var videoBases []string
	audioDirs := map[string]bool{}
	videoDirs := map[string]bool{}
	for _, f := range files {
		dir := path.Dir(f.Path)
		switch ext := strings.ToLower(path.Ext(f.Path)); {
		case videoExt[ext]:
			videoBases = append(videoBases, stripExt(path.Base(f.Path)))
			videoDirs[dir] = true
		case audioExt[ext]:
			audioDirs[dir] = true
		}
	}
	several := len(videoBases) > 1
	// An episode number is a usable key only if it tells this video apart from the others.
	byNumber := several && uniqueEpisode(videoBase, videoBases)

	var out []Track
	for _, f := range files {
		ext := strings.ToLower(path.Ext(f.Path))
		if !subExt[ext] || f.Size > MaxSubSize {
			continue
		}
		dir := path.Dir(f.Path)
		// Subtitle files among audio files, with no video in the same folder, are
		// companions of that audio (for example "signs" for a voice-over), not subtitle
		// tracks of the video.
		if audioDirs[dir] && !videoDirs[dir] {
			continue
		}
		if !related(videoDir, dir) {
			continue
		}
		// One video in the torrent: all subtitles belong to it. Several videos: match by name or episode.
		if several && !sameEpisode(videoBase, stripExt(path.Base(f.Path)), byNumber) {
			continue
		}
		out = append(out, Track{
			FileIndex: f.Index,
			Name:      path.Base(f.Path),
			Lang:      langOf(f.Path),
			Format:    strings.TrimPrefix(ext, "."),
			dir:       dirLabel(dir),
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
	name := path.Base(dir)
	if name == "." || name == "/" {
		return ""
	}
	name = strings.NewReplacer("[", " ", "]", " ", "{", " ", "}", " ", "_", " ").Replace(name)
	var kept []string
	for _, w := range strings.Fields(name) {
		if dirStopWords[strings.ToLower(strings.Trim(w, ".,:;-"))] {
			continue
		}
		kept = append(kept, w)
	}
	return strings.Trim(strings.Join(kept, " "), " .,:;-")
}

// related reports whether one directory contains the other (subtitles next to
// the video, in a Subs/ subfolder, or in a common parent directory).
func related(a, b string) bool {
	if a == "." || b == "." || a == b {
		return true
	}
	return strings.HasPrefix(b, a+"/") || strings.HasPrefix(a, b+"/")
}

// sameEpisode reports whether the subtitle file belongs to the video: its name starts
// with the video name, or (when byNumber is set) both names carry the same episode number.
func sameEpisode(videoBase, subBase string, byNumber bool) bool {
	if hasNamePrefix(subBase, videoBase) {
		return true
	}
	if !byNumber {
		return false
	}
	vs, ve, ok1 := episodeOf(videoBase)
	ss, se, ok2 := episodeOf(subBase)
	return ok1 && ok2 && sameEpisodeNumber(vs, ve, ss, se)
}

// hasNamePrefix reports whether name starts with prefix (case-insensitive) and the
// prefix ends at a word boundary: "ep01.rus" starts with "ep01", but "Show - 10"
// does not start with "Show - 1" and "Aliens" does not start with "Alien".
func hasNamePrefix(name, prefix string) bool {
	name, prefix = strings.ToLower(name), strings.ToLower(prefix)
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	last, _ := utf8.DecodeLastRuneInString(prefix)
	next, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return len(name) == len(prefix) || !isWordRune(last) || !isWordRune(next)
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// sameEpisodeNumber compares (season, episode) pairs; an unknown season (-1) matches any.
func sameEpisodeNumber(s1, e1, s2, e2 int) bool {
	return e1 == e2 && (s1 < 0 || s2 < 0 || s1 == s2)
}

// uniqueEpisode reports whether the episode number extracted from video differs from
// the numbers of all the other videos. If it does not (for example, the extracted
// number is a resolution or a codec name in every file name), it cannot be used to
// pick subtitles.
func uniqueEpisode(video string, videos []string) bool {
	s, e, ok := episodeOf(video)
	if !ok {
		return false
	}
	n := 0
	for _, other := range videos {
		if s2, e2, ok := episodeOf(other); ok && sameEpisodeNumber(s, e, s2, e2) {
			n++
		}
	}
	return n == 1
}

// episodeOf extracts (season, episode). Season is -1 when not specified.
func episodeOf(name string) (season, ep int, ok bool) {
	if m := reSE.FindStringSubmatch(name); m != nil {
		s, _ := strconv.Atoi(m[1])
		e, _ := strconv.Atoi(m[2])
		return s, e, true
	}
	if m := reX.FindStringSubmatch(name); m != nil {
		s, _ := strconv.Atoi(m[1])
		e, _ := strconv.Atoi(m[2])
		return s, e, true
	}

	// Standalone numbers. The last one that is not part of a codec name ("x.265",
	// "H.264") or a decimal ("5.1", "2.0") is the episode: "Mob Psycho 100 - 02" -> 02.
	all := reNum.FindAllStringSubmatchIndex(name, -1)
	for i := len(all) - 1; i >= 0; i-- {
		start, end := all[i][2], all[i][3]
		before, after := name[:start], name[end:]
		if reCodecBefore.MatchString(before) || reDecimalBefore.MatchString(before) || reDecimalAfter.MatchString(after) {
			continue
		}
		e, _ := strconv.Atoi(name[start:end])
		return -1, e, true
	}
	return 0, 0, false
}

// langOf detects the language of a subtitle file.
//
//  1. An explicit language marker in the file name or in a parent folder name
//     wins (nearest first): "ep01.eng.srt", "Subs/en/ep01.srt".
//  2. Otherwise, if the path (the file name or any folder name) contains Cyrillic
//     letters, the language is assumed to be Russian.
//  3. Otherwise the language is unknown ("und").
func langOf(p string) string {
	segs := strings.Split(p, "/")
	for i := len(segs) - 1; i >= 0; i-- {
		for _, w := range words(segs[i]) {
			if l, ok := langTokens[w]; ok {
				return l
			}
		}
	}
	for _, r := range p {
		if unicode.Is(unicode.Cyrillic, r) {
			return "ru"
		}
	}
	return "und"
}

// words splits a name into lower-case words on any non-letter character.
func words(name string) []string {
	return strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return !unicode.IsLetter(r)
	})
}

func stripExt(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}
