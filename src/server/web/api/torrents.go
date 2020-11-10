package api

import (
	"net/http"

	"server/log"
	"server/torr"
	"server/torr/state"
	"server/web/api/utils"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//Action: add, get, rem, list, drop
type torrReqJS struct {
	requestI
	Link     string `json:"link,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Title    string `json:"title,omitempty"`
	Poster   string `json:"poster,omitempty"`
	SaveToDB bool   `json:"save_to_db,omitempty"`
}

func torrents(c *gin.Context) {
	var req torrReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusBadRequest)
	switch req.Action {
	case "add":
		{
			addTorrent(req, c)
		}
	case "get":
		{
			getTorrent(req, c)
		}
	case "rem":
		{
			remTorrent(req, c)
		}
	case "list":
		{
			listTorrent(req, c)
		}
	case "drop":
		{
			dropTorrent(req, c)
		}
	}
}

func addTorrent(req torrReqJS, c *gin.Context) {
	log.TLogln("add torrent", req.Link)
	torrSpec, err := utils.ParseLink(req.Link)
	if err != nil {
		log.TLogln("error add torrent:", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	torr, err := torr.NewTorrent(torrSpec, bts)
	if err != nil {
		log.TLogln("error add torrent:", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !torr.GotInfo() {
		log.TLogln("error add torrent:", "timeout connection torrent")
		c.AbortWithError(http.StatusNotFound, errors.New("timeout connection torrent"))
		return
	}

	torr.Title = req.Title
	torr.Poster = req.Poster

	if torr.Title == "" {
		torr.Title = torr.Name()
	}

	if req.SaveToDB {
		log.TLogln("save to db:", torr.Torrent.InfoHash().HexString())
		utils.AddTorrent(torr)
	}

	st := torr.Stats()
	c.JSON(200, st)
}

func getTorrent(req torrReqJS, c *gin.Context) {
	hash := metainfo.NewHashFromHex(req.Hash)
	tor := bts.GetTorrent(hash)
	if tor == nil {
		tor = utils.GetTorrent(hash)
	}

	if tor != nil {
		st := tor.Stats()
		c.JSON(200, st)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func remTorrent(req torrReqJS, c *gin.Context) {
	hash := metainfo.NewHashFromHex(req.Hash)
	bts.RemoveTorrent(hash)
	utils.RemTorrent(hash)
	c.Status(200)
}

func listTorrent(req torrReqJS, c *gin.Context) {
	stats := listTorrents()
	c.JSON(200, stats)
}

func listTorrents() []*state.TorrentStats {
	btlist := bts.ListTorrents()
	dblist := utils.ListTorrents()
	var stats []*state.TorrentStats
	for _, tr := range btlist {
		stats = append(stats, tr.Stats())
	}

mainloop:
	for _, db := range dblist {
		for _, tr := range btlist {
			if tr.Hash() == db.Hash() {
				continue mainloop
			}
		}
		stats = append(stats, db.Stats())
	}
	return stats
}

func dropTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	hash := metainfo.NewHashFromHex(req.Hash)

	bts.RemoveTorrent(hash)
	c.Status(200)
}
