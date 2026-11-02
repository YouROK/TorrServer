package api

import (
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/subs"
)

// Fonts for the ASS subtitles of a video file, in one list for the player:
//
//	GET /subs/{hash}/{id}?fonts   - JSON [{source, name, type, size, url}]
//
// source "external": a font file of the torrent, url is a /stream link of that file;
// source "embedded": a font attached to the video file (added together with the embedded subtitles,
// url is /subs/{hash}/{id}?font=N).
//
// The player does not care where a font comes from: it reads the name table of every font by Range requests to
// "url" and hands the fonts that the subtitles use to the renderer.
//
// External fonts: only the fonts that belong to the video are listed. These are the font files in the folder of
// the video or below it; if there are none, those in the folder above (with its subfolders), and so on up to the
// root of the torrent.
type fontJS struct {
	Source string `json:"source"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Size   int64  `json:"size"`
	URL    string `json:"url"`
}

// torrentFile is a file of the torrent as /stream knows it: ID is the "index" of the /stream and /play links.
type torrentFile struct {
	ID   int
	Path string
	Size int64
}

// Same formats as the client reads (WOFF and WOFF2 cannot be read by Range requests).
func isFontFileName(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".ttf", ".otf", ".ttc":
		return true
	}
	return false
}

func fontContentType(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".otf":
		return "font/otf"
	case ".ttc":
		return "font/collection"
	}
	return "font/ttf"
}

// nearestFonts returns the font files that belong to the video (see above), sorted by path.
func nearestFonts(videoPath string, files []torrentFile) []torrentFile {
	norm := func(p string) string { return strings.ReplaceAll(p, "\\", "/") }

	var fonts []torrentFile
	for _, f := range files {
		if isFontFileName(f.Path) {
			f.Path = norm(f.Path)
			fonts = append(fonts, f)
		}
	}
	sort.Slice(fonts, func(i, j int) bool { return fonts[i].Path < fonts[j].Path })

	dir := path.Dir(norm(videoPath))
	for {
		root := dir == "." || dir == "/"
		var found []torrentFile
		for _, f := range fonts {
			if root || strings.HasPrefix(f.Path, dir+"/") {
				found = append(found, f)
			}
		}
		if len(found) > 0 || root {
			return found
		}
		dir = path.Dir(dir)
	}
}

// externalFontURL is the same link that the web interface builds for the files of a torrent.
func externalFontURL(hash string, f torrentFile) string {
	return "/stream/" + url.PathEscape(path.Base(f.Path)) + "?link=" + hash + "&index=" + strconv.Itoa(f.ID) + "&play"
}

// subsFonts answers GET /subs/{hash}/{id}?fonts. files is the list of the files of the torrent that subsTracks
// already has (Index is the id of /stream and /play links, Path uses "/"); video is the requested video file.
func subsFonts(c *gin.Context, hash string, video subs.FileInfo, files []subs.FileInfo) {
	candidates := make([]torrentFile, 0, len(files))
	for _, f := range files {
		candidates = append(candidates, torrentFile{ID: f.Index, Path: f.Path, Size: f.Size})
	}

	out := make([]fontJS, 0)
	for _, f := range nearestFonts(video.Path, candidates) {
		out = append(out, fontJS{
			Source: "external",
			Name:   path.Base(f.Path),
			Type:   fontContentType(f.Path),
			Size:   f.Size,
			URL:    externalFontURL(hash, f),
		})
	}
	// Embedded subtitles add the fonts attached to a Matroska file here (source "embedded").
	c.JSON(http.StatusOK, out)
}
