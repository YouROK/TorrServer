package api

import (
	"net/http"
	"strings"
	"os"
	"path/filepath"
	"net/url"
	"fmt"
	"time"

	"server/dlna"
	"server/log"
	set "server/settings"
	"server/torr"
	"server/torr/state"
	"server/web/api/utils"
	utils2 "server/utils"

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
}

// torrents godoc
//
//	@Summary		Handle torrents informations
//	@Description	Allow to list, add, remove, get, set, drop, wipe torrents on server. The action depends of what has been asked.
//
//	@Tags			API
//
//	@Param			request	body	torrReqJS	true	"Torrent request. Available params for action: add, get, set, rem, list, drop, wipe. link required for add, hash required for get, set, rem, drop."
//
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Router			/torrents [post]
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
	case "addJlfn":
		{
			addJlfn(req, c)
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
	}
}

func addTorrent(req torrReqJS, c *gin.Context) {
	if req.Link == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("link is empty"))
		return
	}

	log.TLogln("add torrent", req.Link)
	req.Link = strings.ReplaceAll(req.Link, "&amp;", "&")
	torrSpec, err := utils.ParseLink(req.Link)
	if err != nil {
		log.TLogln("error parse link:", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	

	tor, err := torr.AddTorrent(torrSpec, req.Title, req.Poster, req.Data, req.Category)
	// if tor.Data != "" && set.BTsets.EnableDebug {
	// 	log.TLogln("torrent data:", tor.Data)
	// }
	// if tor.Category != "" && set.BTsets.EnableDebug {
	// 	log.TLogln("torrent category:", tor.Category)
	// }
	if err != nil {
		log.TLogln("error add torrent:", err)
		c.AbortWithError(http.StatusInternalServerError, err)
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
	// TODO: remove
	if set.BTsets.EnableDLNA {
		dlna.Stop()
		dlna.Start()
	}
	c.JSON(200, tor.Status())
}

func getTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
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

func addJlfn(req torrReqJS, c *gin.Context) {

	addTorrent(req , c)
	
	go func() {
		time.Sleep(15 * time.Second)
		req.Link = strings.ReplaceAll(req.Link, "&amp;", "&")
		torrSpec, err := utils.ParseLink(req.Link)
		if err != nil {
			log.TLogln("error parse link:", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	
		hash := torrSpec.InfoHash.String()
		log.TLogln("new add torrent", hash)
	
		tor := torr.GetTorrent(hash)

		if tor == nil {
			return
			log.TLogln("tor", "null")
		}
	
		basePath := set.JlfnAddr
	
		if basePath == "" {
			return
			log.TLogln("basePath", "null")
		}

		if tor.Stat == state.TorrentInDB {
			tor = torr.LoadTorrent(tor)
			if tor == nil {
				return
				log.TLogln("LoadTorrent", "null")
			}
		}
		host := utils2.GetScheme(c) + "://" + c.Request.Host
		torrents := tor.Status()

		from := 0

		// Создаем базовый путь для сохранения
		CatPath := ""
		if len(torrents.FileStats) > 1 {
			CatPath = "torrSerials"
		} else {
			CatPath = "torrFilms"
		}
		
	
		torname := tor.Name()
		basePath = filepath.Join(basePath, CatPath)
		basePath = filepath.Join(basePath, torname)
	
		for i, f := range torrents.FileStats {
			if i >= from {
				if utils2.GetMimeType(f.Path) != "*/*" {
					list := ""
								
					name := filepath.Base(f.Path)
					list += host + "/stream/" + url.PathEscape(name) + "?link=" + torrents.Hash + "&index=" + fmt.Sprint(f.Id) + "&play"
				
					// Создаем имя файла .strm на основе имени файла
					strmName := filepath.Base(f.Path)
					strmName = strings.ReplaceAll(strmName, `/`, "") // strip starting / from param
				
					// Добавляем расширение .strm если его нет
					if !strings.HasSuffix(strings.ToLower(strmName), ".strm") {
						strmName += ".strm"
					}
				
					// Полный путь к файлу = базовый путь + имя файла
				
					fullPath := filepath.Join(basePath, strmName)
				
					// Создаем директорию, если не существует
					if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
						return
					}
				
					// Создаем и записываем файл
					if err := os.WriteFile(fullPath, []byte(list), 0644); err != nil {
						return
					}
				}
			}
		}
		go func() {
			time.Sleep(15 * time.Second)
			torr.DropTorrent(hash)
		}()
	}()
	
}



func setTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	torr.SetTorrent(req.Hash, req.Title, req.Poster, req.Category, req.Data)
	c.Status(200)
}

func remTorrent(req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	torr.RemTorrent(req.Hash)
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
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	torr.DropTorrent(req.Hash)
	c.Status(200)
}

func wipeTorrents(c *gin.Context) {
	torrents := torr.ListTorrent()
	for _, t := range torrents {
		torr.RemTorrent(t.TorrentSpec.InfoHash.HexString())
	}
	// TODO: remove (copied todo from remTorrent())
	if set.BTsets.EnableDLNA {
		dlna.Stop()
		dlna.Start()
	}
	c.Status(200)
}
