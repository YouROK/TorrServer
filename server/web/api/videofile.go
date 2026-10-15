package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/extfiles"
	"server/torr"
	"server/torr/state"
	"server/web/api/utils"
)

// videoContext is what the /subs and /audio handlers need to know about the requested
// video file and the torrent it belongs to.
type videoContext struct {
	tor      *torr.Torrent
	infoHash string
	videoID  int
	video    extfiles.FileInfo
	files    []extfiles.FileInfo
}

// loadVideoContext resolves the :hash and :id parameters of the request into a video
// file of a torrent that is ready to be read. On failure it aborts the request with the
// proper status and returns false.
func loadVideoContext(c *gin.Context) (*videoContext, bool) {
	hash := c.Param("hash")
	indexStr := c.Param("id")
	notAuth := c.GetBool("auth_required") && c.GetString(gin.AuthUserKey) == ""

	if hash == "" || indexStr == "" {
		c.AbortWithError(http.StatusNotFound, errors.New("no infohash or file index in link"))
		return nil, false
	}

	spec, err := utils.ParseLink(hash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return nil, false
	}

	tor := torr.GetTorrent(spec.InfoHash.HexString())
	if tor == nil && notAuth {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil, false
	}
	if tor == nil {
		c.AbortWithError(http.StatusNotFound, errors.New("torrent not found"))
		return nil, false
	}

	if tor.Stat == state.TorrentInDB {
		tor, err = torr.AddTorrent(spec, tor.Title, tor.Poster, tor.Data, tor.Category)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return nil, false
		}
	}

	if !tor.GotInfo() {
		c.AbortWithError(http.StatusInternalServerError, errors.New("torrent connection timeout"))
		return nil, false
	}

	// Video index: same rules as in play.go.
	videoID := -1
	if len(tor.Files()) == 1 {
		videoID = 1
	} else if id, err := strconv.Atoi(indexStr); err == nil {
		videoID = id
	}
	if videoID == -1 {
		c.AbortWithError(http.StatusBadRequest, errors.New("file \"index\" is wrong"))
		return nil, false
	}

	lengths := make(map[string]int64)
	for _, f := range tor.Files() {
		lengths[f.Path()] = f.Length()
	}

	var files []extfiles.FileInfo
	for _, fs := range tor.Status().FileStats {
		files = append(files, extfiles.FileInfo{
			Index: fs.Id,
			Path:  strings.ReplaceAll(fs.Path, "\\", "/"),
			Size:  lengths[fs.Path],
		})
	}

	var video *extfiles.FileInfo
	for i := range files {
		if files[i].Index == videoID {
			video = &files[i]
			break
		}
	}
	if video == nil {
		c.AbortWithError(http.StatusNotFound, errors.New("video file not found"))
		return nil, false
	}

	if !extfiles.IsVideo(video.Path) {
		c.AbortWithError(http.StatusBadRequest, errors.New("file is not a video"))
		return nil, false
	}

	return &videoContext{
		tor:      tor,
		infoHash: spec.InfoHash.HexString(),
		videoID:  videoID,
		video:    *video,
		files:    files,
	}, true
}
