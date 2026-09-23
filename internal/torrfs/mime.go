package torrfs

import (
	"path/filepath"
	"strings"
)

var mimeByExt = map[string]string{
	// video
	".mp4":  "video/mp4",
	".mkv":  "video/x-matroska",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".wmv":  "video/x-ms-wmv",
	".flv":  "video/x-flv",
	".mpg":  "video/mpeg",
	".mpeg": "video/mpeg",
	".m4v":  "video/x-m4v",
	".ts":   "video/mp2t",
	".m2ts": "video/mp2t",
	".3gp":  "video/3gpp",
	".ogv":  "video/ogg",
	".vob":  "video/dvd",

	// audio
	".mp3":  "audio/mpeg",
	".flac": "audio/flac",
	".aac":  "audio/aac",
	".ogg":  "audio/ogg",
	".oga":  "audio/ogg",
	".opus": "audio/opus",
	".wav":  "audio/wav",
	".m4a":  "audio/mp4",
	".wma":  "audio/x-ms-wma",
	".ape":  "audio/x-ape",

	// image
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".bmp":  "image/bmp",
	".tiff": "image/tiff",
	".ico":  "image/x-icon",

	// subtitles & text
	".srt":  "application/x-subrip",
	".vtt":  "text/vtt",
	".ass":  "text/x-ssa",
	".ssa":  "text/x-ssa",
	".sub":  "text/plain; charset=utf-8",
	".txt":  "text/plain; charset=utf-8",
	".nfo":  "text/plain; charset=utf-8",
	".md":   "text/markdown; charset=utf-8",
	".json": "application/json",
	".xml":  "application/xml",
	".html": "text/html; charset=utf-8",
	".htm":  "text/html; charset=utf-8",
	".css":  "text/css",
	".js":   "application/javascript",

	// archives
	".zip": "application/zip",
	".rar": "application/vnd.rar",
	".7z":  "application/x-7z-compressed",
	".tar": "application/x-tar",
	".gz":  "application/gzip",

	// docs
	".pdf": "application/pdf",
}

func mimeByExtension(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if m, ok := mimeByExt[ext]; ok {
		return m
	}
	return "application/octet-stream"
}
