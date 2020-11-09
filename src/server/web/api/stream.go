package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"server/torr"
	"server/web/api/utils"
)

func stream(c *gin.Context) {
	link := c.Query("link")
	indexStr := c.Query("index")
	_, preload := c.GetQuery("preload")
	_, stat := c.GetQuery("stat")
	_, save := c.GetQuery("save")
	_, m3u := c.GetQuery("m3u")
	_, play := c.GetQuery("play")
	title := c.Query("title")
	poster := c.Query("poster")

	// TODO unescape args

	if link == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("link should not be empty"))
		return
	}

	spec, err := utils.ParseLink(link)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var tor *torr.Torrent

	// find torrent in bts
	for _, torrent := range bts.ListTorrents() {
		if torrent.Hash().HexString() == spec.InfoHash.HexString() {
			tor = torrent
		}
	}
	// find in db
	for _, torrent := range utils.ListTorrents() {
		if torrent.Hash().HexString() == spec.InfoHash.HexString() {
			tor = torrent
		}
	}
	// add torrent to bts
	if tor != nil {
		tor, err = torr.NewTorrent(tor.TorrentSpec, bts)
	} else {
		tor, err = torr.NewTorrent(spec, bts)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	tor.Title = title
	tor.Poster = poster

	// save to db
	if save {
		utils.AddTorrent(tor)
	}
	// wait torrent info
	if !tor.WaitInfo() {
		c.AbortWithError(http.StatusInternalServerError, errors.New("timeout torrent get info"))
		return
	}

	// find file
	index := -1
	if len(tor.Files()) == 1 {
		index = 0
	} else {
		ind, err := strconv.Atoi(indexStr)
		if err == nil {
			index = ind
		}
	}
	if index == -1 {
		c.AbortWithError(http.StatusBadRequest, errors.New("\"index\" is empty or wrong"))
		return
	}
	// preload torrent
	if preload {
		tor.Preload(index, 0)
	}
	// return stat if query
	if stat || (!m3u && !play) {
		c.JSON(200, tor.Stats())
		return
	}
	// return m3u if query
	if m3u {
		//TODO m3u
		c.JSON(200, tor.Stats())
		return
	}
	// return play if query
	if play {
		tor.Stream(index, c.Request, c.Writer)
		return
	}
}

/*
func torrentPlay(c echo.Context) error {
	link := c.QueryParam("link")
	if link == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "link should not be empty")
	}
	if settings.Get().EnableDebug {
		fmt.Println("Play:", c.QueryParams()) // mute log flood on play
	}
	qsave := c.QueryParam("save")
	qpreload := c.QueryParam("preload")
	qfile := c.QueryParam("file")
	qstat := c.QueryParam("stat")
	mm3u := c.QueryParam("m3u")

	preload := int64(0)
	stat := strings.ToLower(qstat) == "true"

	if qpreload != "" {
		preload, _ = strconv.ParseInt(qpreload, 10, 64)
		if preload > 0 {
			preload *= 1024 * 1024
		}
	}

	magnet, infoBytes, err := helpers.GetMagnet(link)
	if err != nil {
		fmt.Println("Error get magnet:", link, err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tor := bts.GetTorrent(magnet.InfoHash)
	if tor == nil {
		tor, err = bts.AddTorrent(*magnet, infoBytes, nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	if stat {
		return c.JSON(http.StatusOK, getTorPlayState(tor))
	}

	if !tor.WaitInfo() {
		return echo.NewHTTPError(http.StatusBadRequest, "torrent closed befor get info")
	}

	if strings.ToLower(qsave) == "true" {
		if t, err := settings.LoadTorrentDB(magnet.InfoHash.HexString()); t == nil && err == nil {
			torrDb := toTorrentDB(tor)
			if torrDb != nil {
				torrDb.InfoBytes = infoBytes
				settings.SaveTorrentDB(torrDb)
			}
		}
	}

	if strings.ToLower(mm3u) == "true" {
		mt := tor.Torrent.Metainfo()
		m3u := helpers.MakeM3UPlayList(tor.Stats(), mt.Magnet(tor.Name(), tor.Hash()).String(), c.Scheme()+"://"+c.Request().Host)
		c.Response().Header().Set("Content-Type", "audio/x-mpegurl")
		c.Response().Header().Set("Connection", "close")
		name := utils.CleanFName(tor.Name()) + ".m3u"
		c.Response().Header().Set("ETag", httptoo.EncodeQuotedString(fmt.Sprintf("%s/%s", tor.Hash().HexString(), name)))
		c.Response().Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
		http.ServeContent(c.Response(), c.Request(), name, time.Now(), bytes.NewReader([]byte(m3u)))
		return c.NoContent(http.StatusOK)
	}

	files := helpers.GetPlayableFiles(tor.Stats())

	if len(files) == 1 {
		file := helpers.FindFile(files[0].Id, tor)
		if file == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprint("File", files[0], "not found in torrent", tor.Name()))
		}

		return bts.Play(tor, file, preload, c)
	}

	if qfile == "" && len(files) > 1 {
		return c.JSON(http.StatusOK, getTorPlayState(tor))
	}

	fileInd, _ := strconv.Atoi(qfile)
	file := helpers.FindFile(fileInd, tor)
	if file == nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprint("File index ", fileInd, " not found in torrent ", tor.Name()))
	}
	return bts.Play(tor, file, preload, c)
}
*/
