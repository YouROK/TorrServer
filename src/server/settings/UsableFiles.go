package settings

import (
	"path/filepath"
	"sort"
	"strings"
)

var (
	uFiles = map[string]interface{}{
		".3g2":   nil,
		".3gp":   nil,
		".aaf":   nil,
		".asf":   nil,
		".avchd": nil,
		".avi":   nil,
		".drc":   nil,
		".flv":   nil,
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
		".mxf":   nil,
		".nsv":   nil,
		".ogg":   nil,
		".ogv":   nil,
		".ts":    nil,
		".m2ts":  nil,
		".mts":   nil,
		".qt":    nil,
		".rm":    nil,
		".rmvb":  nil,
		".roq":   nil,
		".svi":   nil,
		".vob":   nil,
		".webm":  nil,
		".wmv":   nil,
		".yuv":   nil,

		".aac":  nil,
		".aiff": nil,
		".ape":  nil,
		".au":   nil,
		".flac": nil,
		".gsm":  nil,
		".it":   nil,
		".m3u":  nil,
		".m4a":  nil,
		".mid":  nil,
		".mod":  nil,
		".mp3":  nil,
		".mpa":  nil,
		".pls":  nil,
		".ra":   nil,
		".s3m":  nil,
		".sid":  nil,
		".wav":  nil,
		".wma":  nil,
		".xm":   nil,
	}
)

func SortFiles(files []File) {
	sort.Slice(files, func(i, j int) bool {
		if haveUsable(files[i].Name) && !haveUsable(files[j].Name) {
			return true
		}
		if !haveUsable(files[i].Name) && haveUsable(files[j].Name) {
			return false
		}

		return files[i].Name < files[j].Name
	})
}

func haveUsable(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := uFiles[ext]
	return ok
}
