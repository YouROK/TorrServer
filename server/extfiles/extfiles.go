// Package extfiles finds the files that lie next to a video in a torrent and belong to
// it: external subtitles (see package subs) and external audio tracks (MatchAudio).
// It uses the standard library only.
package extfiles

import (
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// FileInfo describes a torrent file. Index must match the file index
// used by /play/:hash/:id.
type FileInfo struct {
	Index int
	Path  string // path inside the torrent, e.g. "Show/Subs/Rus/ep01.ass"
	Size  int64
}

var (
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

var langTokens = map[string]string{
	"ru": "ru", "rus": "ru", "russian": "ru", "рус": "ru", "русский": "ru", "русские": "ru",
	"en": "en", "eng": "en", "english": "en", "англ": "en", "английский": "en", "английские": "en",
	"ja": "ja", "jpn": "ja", "jap": "ja", "japanese": "ja", "японский": "ja", "японские": "ja",
}

// IsVideo reports whether the file looks like a video (by extension).
func IsVideo(p string) bool {
	return videoExt[strings.ToLower(path.Ext(p))]
}

// IsAudio reports whether the file looks like an audio file (by extension).
func IsAudio(p string) bool {
	return audioExt[strings.ToLower(path.Ext(p))]
}

// Video is a video file together with what is needed to tell which other files of the
// torrent belong to it.
type Video struct {
	dir       string
	base      string // name of the video without the extension
	audioDirs map[string]bool
	videoDirs map[string]bool
	several   bool // the torrent has more than one video file
	byNumber  bool // the episode number tells this video apart from the others
}

// NewVideo prepares the matching of the files of a torrent against one of its videos.
func NewVideo(video FileInfo, files []FileInfo) Video {
	v := Video{
		dir:       path.Dir(video.Path),
		base:      stripExt(path.Base(video.Path)),
		audioDirs: map[string]bool{},
		videoDirs: map[string]bool{},
	}
	var videoBases []string
	for _, f := range files {
		dir := path.Dir(f.Path)
		switch ext := strings.ToLower(path.Ext(f.Path)); {
		case videoExt[ext]:
			videoBases = append(videoBases, stripExt(path.Base(f.Path)))
			v.videoDirs[dir] = true
		case audioExt[ext]:
			v.audioDirs[dir] = true
		}
	}
	v.several = len(videoBases) > 1
	// An episode number is a usable key only if it tells this video apart from the others.
	v.byNumber = v.several && uniqueEpisode(v.base, videoBases)
	return v
}

// Belongs reports whether the file may belong to the video: the folders are related, and
// with several videos in the torrent the file name or the episode number matches too.
// With one video only, any file in a related folder qualifies.
func (v Video) Belongs(filePath string) bool {
	if !related(v.dir, path.Dir(filePath)) {
		return false
	}
	return !v.several || sameEpisode(v.base, stripExt(path.Base(filePath)), v.byNumber)
}

// InAudioOnlyDir reports whether the file lies in a folder that has audio files but no
// video: such a folder is a collection of audio tracks, and other files in it (for example
// "signs" subtitles of a voice-over) are companions of that audio.
func (v Video) InAudioOnlyDir(filePath string) bool {
	dir := path.Dir(filePath)
	return v.audioDirs[dir] && !v.videoDirs[dir]
}

// DirLabel turns the name of a folder into a short label: square and curly brackets and
// underscores become spaces, and the words of stopWords (lower case) are dropped:
// with the stop word "subs", "Subs [Exmel & Reddog]" -> "Exmel & Reddog". It returns ""
// when nothing is left.
func DirLabel(dir string, stopWords map[string]bool) string {
	name := path.Base(dir)
	if name == "." || name == "/" {
		return ""
	}
	name = strings.NewReplacer("[", " ", "]", " ", "{", " ", "}", " ", "_", " ").Replace(name)
	var kept []string
	for _, w := range strings.Fields(name) {
		if stopWords[strings.ToLower(strings.Trim(w, ".,:;-"))] {
			continue
		}
		kept = append(kept, w)
	}
	return strings.Trim(strings.Join(kept, " "), " .,:;-")
}

// LangOf detects the language of a subtitle or audio file.
//
//  1. An explicit language marker in the file name or in a parent folder name
//     wins (nearest first): "ep01.eng.srt", "Subs/en/ep01.srt".
//  2. Otherwise, if the path (the file name or any folder name) contains Cyrillic
//     letters, the language is assumed to be Russian.
//  3. Otherwise the language is unknown ("und").
func LangOf(p string) string {
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

// related reports whether one directory contains the other (a file next to the video,
// in a subfolder, or in a common parent directory).
func related(a, b string) bool {
	if a == "." || b == "." || a == b {
		return true
	}
	return strings.HasPrefix(b, a+"/") || strings.HasPrefix(a, b+"/")
}

// sameEpisode reports whether the file belongs to the video: its name starts with the
// video name, or (when byNumber is set) both names carry the same episode number.
func sameEpisode(videoBase, fileBase string, byNumber bool) bool {
	if strings.HasPrefix(strings.ToLower(fileBase), strings.ToLower(videoBase)) {
		return true
	}
	if !byNumber {
		return false
	}
	vs, ve, ok1 := episodeOf(videoBase)
	fs, fe, ok2 := episodeOf(fileBase)
	return ok1 && ok2 && sameEpisodeNumber(vs, ve, fs, fe)
}

// sameEpisodeNumber compares (season, episode) pairs; an unknown season (-1) matches any.
func sameEpisodeNumber(s1, e1, s2, e2 int) bool {
	return e1 == e2 && (s1 < 0 || s2 < 0 || s1 == s2)
}

// uniqueEpisode reports whether the episode number extracted from video differs from
// the numbers of all the other videos. If it does not (for example, the extracted
// number is a resolution or a codec name in every file name), it cannot be used to
// pick the files of the video.
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

// words splits a name into lower-case words on any non-letter character.
func words(name string) []string {
	return strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return !unicode.IsLetter(r)
	})
}

func stripExt(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}
