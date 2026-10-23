package extfiles

import (
	"path"
	"sort"
	"strings"
)

// AudioTrack is an external audio track (a separate audio file) found for a particular
// video file, for example a voice-over in "Sound [Group]/ep01.mka".
type AudioTrack struct {
	FileIndex int    `json:"index"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Lang      string `json:"lang"`   // "ru", "en", ... or "und"
	Format    string `json:"format"` // file extension without the dot: "mka", "ac3", "flac", ...
	Size      int64  `json:"size"`

	dir string // cleaned name of the folder, see DirLabel
}

// audioDirWords are folder-name words that mark a folder with audio tracks. They are
// dropped from the folder label ("Sound [Group]" -> "Group") and, when the torrent has
// one video only, they tell a voice-over folder from a folder with music.
var audioDirWords = map[string]bool{
	"sound": true, "audio": true, "dub": true, "dubs": true, "dubbing": true,
	"voice": true, "voices": true, "voiceover": true,
	"озвучка": true, "озвучки": true, "звук": true, "аудио": true, "дубляж": true,
	"дорожка": true, "дорожки": true,
}

// MatchAudio returns the external audio tracks that belong to the given video file.
//
// A file belongs to the video by the rules of Video.Belongs. With one video only they
// are stricter, because a folder with the soundtrack of a movie must not turn into
// dubbing tracks: an audio file is accepted when its name starts with the video name,
// when it lies in the folder of the video, or when it lies below that folder in a folder
// named like an audio folder (Sound, Audio, Dub, Озвучка, ...).
func MatchAudio(video FileInfo, files []FileInfo) []AudioTrack {
	v := NewVideo(video, files)

	var out []AudioTrack
	for _, f := range files {
		ext := strings.ToLower(path.Ext(f.Path))
		if !audioExt[ext] || !v.Belongs(f.Path) {
			continue
		}
		if !v.several && !v.singleVideoAudio(f.Path) {
			continue
		}
		out = append(out, AudioTrack{
			FileIndex: f.Index,
			Name:      path.Base(f.Path),
			Lang:      LangOf(f.Path),
			Format:    strings.TrimPrefix(ext, "."),
			Size:      f.Size,
			dir:       DirLabel(path.Dir(f.Path), audioDirWords),
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

// singleVideoAudio decides whether an audio file belongs to the only video of the torrent.
func (v Video) singleVideoAudio(filePath string) bool {
	dir := path.Dir(filePath)
	if strings.HasPrefix(strings.ToLower(stripExt(path.Base(filePath))), strings.ToLower(v.base)) {
		return true
	}
	if dir == v.dir {
		return true
	}
	// A folder below the folder of the video, with an audio word in its name.
	rest := ""
	switch {
	case v.dir == ".":
		rest = dir
	case strings.HasPrefix(dir, v.dir+"/"):
		rest = dir[len(v.dir)+1:]
	}
	for _, segment := range strings.Split(rest, "/") {
		for _, w := range words(segment) {
			if audioDirWords[w] {
				return true
			}
		}
	}
	return false
}
