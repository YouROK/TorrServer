package api

import (
	"fmt"
	"net/http"
	"strings"

	"server/torrshash"

	"server/dlna"
	gstreamer "server/gstreamer/bridge"
	"server/log"
	set "server/settings"
	"server/torr"
	"server/torr/state"
	"server/torr/storage/torrstor"
	"server/web/api/utils"

	"github.com/anacrolix/torrent"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// Action: add, get, set, rem, list, drop
type torrReqJS struct {
	requestI
	Link     string `json:"link,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
	Poster   string `json:"poster,omitempty"`
	Data     string `json:"data,omitempty"`
	SaveToDB bool   `json:"save_to_db,omitempty"`
	// Index is the file id (file_stats[].id) for download and cancel_download;
	// 0 means the whole torrent
	Index int `json:"index,omitempty"`
}

// torrents godoc
//
//	@Summary		Handle torrents informations
//	@Description	Allow to list, add, remove, get, set, drop, wipe torrents on server, and download a file or a whole torrent into the long-term disk cache. The action depends of what has been asked.
//
//	@Tags			API
//
//	@Param			request	body	torrReqJS	true	"Torrent request. Available params for action: add, get, set, rem, list, drop, wipe, download, cancel_download. link required for add, hash required for get, set, rem, drop, download, cancel_download; index selects the file for download and cancel_download (0 = whole torrent)."
//
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200
//	@Router			/torrents [post]
func torrents(c *gin.Context) {
	var req torrReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch req.Action {
	case "add":
		{
			addTorrent(req, c)
		}
	case "get":
		{
			getTorrent(req, c)
		}
	case "set":
		{
			setTorrent(req, c)
		}
	case "rem":
		{
			remTorrent(req, c)
		}
	case "list":
		{
			listTorrents(c)
		}
	case "drop":
		{
			dropTorrent(req, c)
		}
	case "wipe":
		{
			wipeTorrents(c)
		}
	case "download":
		{
			downloadTorrent(req, c, false)
		}
	case "cancel_download":
		{
			downloadTorrent(req, c, true)
		}
	default:
		{
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errors.Errorf("unknown action: %q", req.Action).Error()})
		}
	}
}

func addTorrent(req torrReqJS, c *gin.Context) {
	if req.Link == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "link is empty"})
		return
	}

	log.TLogln("add torrent", req.Link)
	req.Link = strings.ReplaceAll(req.Link, "&amp;", "&")

	var torrSpec *torrent.TorrentSpec
	var torrsHash *torrshash.TorrsHash
	var err error

	if strings.HasPrefix(req.Link, "torrs://") {
		torrSpec, torrsHash, err = utils.ParseTorrsHash(req.Link)
		if err != nil {
			log.TLogln("error parse torrshash:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errors.Errorf("error parse torrshash: %v", err).Error()})
			return
		}
		if req.Title == "" {
			req.Title = torrsHash.Title()
		}
		if req.Poster == "" {
			req.Poster = torrsHash.Poster()
		}
		if req.Category == "" {
			req.Category = torrsHash.Category()
		}
	} else {
		torrSpec, err = utils.ParseLink(req.Link)
		if err != nil {
			log.TLogln("error parse link:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errors.Errorf("error parse link: %v", err).Error()})
			return
		}
	}

	tor, err := torr.AddTorrent(torrSpec, req.Title, req.Poster, req.Data, req.Category)
	if err != nil {
		log.TLogln("error add torrent:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errors.Errorf("error adding torrent: %v", err).Error()})
		return
	}

	go func() {
		if !tor.GotInfo() {
			log.TLogln("error add torrent:", "timeout connection get torrent info")
			return
		}

		if tor.Title == "" {
			tor.Title = torrSpec.DisplayName // prefer dn over name
			tor.Title = strings.ReplaceAll(tor.Title, "rutor.info", "")
			tor.Title = strings.ReplaceAll(tor.Title, "_", " ")
			tor.Title = strings.Trim(tor.Title, " ")
			if tor.Title == "" {
				tor.Title = tor.Name()
			}
		}

		if req.SaveToDB {
			torr.SaveTorrentToDB(tor)
		}
	}()

	if set.BTsets.EnableDLNA {
		dlna.Stop()
		dlna.Start()
	}
	c.JSON(200, tor.Status())
}

func getTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "hash is empty"})
		return
	}
	tor := torr.GetTorrent(req.Hash)

	if tor != nil {
		st := tor.Status()
		c.JSON(200, st)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func setTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "hash is empty"})
		return
	}
	torr.SetTorrent(req.Hash, req.Title, req.Poster, req.Category, req.Data)
	c.Status(200)
}

func remTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "hash is empty"})
		return
	}
	torr.RemTorrent(req.Hash)
	gstreamer.Remove(req.Hash)
	// TODO: remove
	if set.BTsets.EnableDLNA {
		dlna.Stop()
		dlna.Start()
	}
	c.Status(200)
}

func listTorrents(c *gin.Context) {
	list := torr.ListTorrent()
	if len(list) == 0 {
		c.JSON(200, []*state.TorrentStatus{})
		return
	}
	var stats []*state.TorrentStatus
	for _, tr := range list {
		stats = append(stats, tr.Status())
	}
	c.JSON(200, stats)
}

func dropTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "hash is empty"})
		return
	}
	torr.DropTorrent(req.Hash)
	gstreamer.Remove(req.Hash)
	c.Status(200)
}

func wipeTorrents(c *gin.Context) {
	torrents := torr.ListTorrent()
	for _, t := range torrents {
		hash := t.TorrentSpec.InfoHash.HexString()
		torr.RemTorrent(hash)
		gstreamer.Remove(hash)
	}
	// TODO: remove (copied todo from remTorrent())
	if set.BTsets.EnableDLNA {
		dlna.Stop()
		dlna.Start()
	}
	c.Status(200)
}

// downloadTorrent starts or cancels downloading a file (req.Index) or the whole
// torrent into the long-term disk cache, opening the torrent if it is only in the DB.
func downloadTorrent(req torrReqJS, c *gin.Context, cancel bool) {
	if req.Hash == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "hash is empty"})
		return
	}
	if !torrstor.IsLongTermCache() {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": torr.ErrDownloadNeedsLongTermCache.Error()})
		return
	}
	tor := torr.GetTorrent(req.Hash)
	if tor == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "torrent not found"})
		return
	}
	// GetTorrent has started loading a DB torrent in the background; AddTorrent
	// waits for that load and returns the same torrent, as in stream and play
	if tor.Stat == state.TorrentInDB {
		var err error
		tor, err = torr.AddTorrent(tor.TorrentSpec, tor.Title, tor.Poster, tor.Data, tor.Category)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error adding torrent: %v", err)})
			return
		}
	}
	if !tor.GotInfo() {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "torrent connection timeout"})
		return
	}

	var err error
	if cancel {
		err = tor.CancelDownload(req.Index)
	} else {
		err = tor.Download(req.Index)
	}
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, torr.ErrDownloadFileNotFound) {
			status = http.StatusNotFound
		}
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tor.Status())
}
