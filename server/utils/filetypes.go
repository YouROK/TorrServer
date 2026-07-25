package utils

import (
	"path/filepath"
	"strings"

	"server/torr/state"
)

var extVideo = map[string]interface{}{
	".3g2":   nil,
	".3gp":   nil,
	".aaf":   nil,
	".asf":   nil,
	".avchd": nil,
	".avi":   nil,
	".drc":   nil,
	".dv":    nil,
	".flv":   nil,
	".iso":   nil,
	".m2ts":  nil,
	".m2v":   nil,
	".m4p":   nil,
	".m4v":   nil,
	".mkv":   nil,
	".mng":   nil,
	".mov":   nil,
	".mp2":   nil,
	".mp4":   nil,
	".mpe":   nil,
	".mpeg":  nil,
	".mpg":   nil,
	".mpv":   nil,
	".mts":   nil,
	".mxf":   nil,
	".nsv":   nil,
	".ogv":   nil,
	".qt":    nil,
	".rm":    nil,
	".rmvb":  nil,
	".roq":   nil,
	".svi":   nil,
	".ts":    nil,
	".vob":   nil,
	".webm":  nil,
	".wmv":   nil,
	".yuv":   nil,
}

var extAudio = map[string]interface{}{
	".aac":  nil,
	".ac3":  nil,
	".aiff": nil,
	".ape":  nil,
	".au":   nil,
	".dff":  nil,
	".dsf":  nil,
	".flac": nil,
	".gsm":  nil,
	".it":   nil,
	".m3u":  nil,
	".m4a":  nil,
	".mid":  nil,
	".mod":  nil,
	".mp3":  nil,
	".mpa":  nil,
	".mpga": nil,
	".oga":  nil,
	".ogg":  nil,
	".opus": nil,
	".pls":  nil,
	".ra":   nil,
	".s3m":  nil,
	".sid":  nil,
	".spx":  nil,
	".wav":  nil,
	".weba": nil,
	".wma":  nil,
	".wv":   nil,
	".wvc":  nil,
	".xm":   nil,
}

func GetMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := extVideo[ext]; ok {
		return "video/*"
	}
	if _, ok := extAudio[ext]; ok {
		return "audio/*"
	}
	return "*/*"
}

func GetPlayableFiles(st state.TorrentStatus) []*state.TorrentFileStat {
	files := make([]*state.TorrentFileStat, 0)
	for _, f := range st.FileStats {
		if GetMimeType(f.Path) != "*/*" {
			files = append(files, f)
		}
	}
	return files
}
