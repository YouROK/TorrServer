package api

import (
	"net/http"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"
	"server/torr"
	"server/web/api/utils"
)

//Action: add, get, rem, list
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

	switch req.Action {
	case "add":
		{
			add(req, c)
		}
	case "get":
		{
			get(req, c)
		}
	case "rem":
		{
			rem(req, c)
		}
	case "list":
		{
			list(req, c)
		}
	}
}

func add(req torrReqJS, c *gin.Context) {
	torrSpec, err := utils.ParseLink(req.Link)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	torr, err := torr.NewTorrent(torrSpec, bts)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	torr.Title = req.Title
	torr.Poster = req.Poster

	if req.SaveToDB {
		utils.AddTorrent(torr)
	}

	st := torr.Stats()
	c.JSON(200, st)
}

func get(req torrReqJS, c *gin.Context) {
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

func rem(req torrReqJS, c *gin.Context) {
	hash := metainfo.NewHashFromHex(req.Hash)
	bts.RemoveTorrent(hash)
	utils.RemTorrent(hash)
	c.Status(200)
}

func list(req torrReqJS, c *gin.Context) {
	btlist := bts.ListTorrents()
	dblist := utils.ListTorrents()
	var stats []torr.TorrentStats
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

	c.JSON(200, stats)
}
