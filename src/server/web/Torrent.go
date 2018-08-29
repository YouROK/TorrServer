package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"server/settings"
	"server/torr"
	"server/utils"
	"server/web/helpers"

	"github.com/anacrolix/missinggo/httptoo"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/labstack/echo"
)

func initTorrent(e *echo.Echo) {
	e.POST("/torrent/add", torrentAdd)
	e.POST("/torrent/upload", torrentUpload)
	e.POST("/torrent/get", torrentGet)
	e.POST("/torrent/rem", torrentRem)
	e.POST("/torrent/list", torrentList)
	e.POST("/torrent/stat", torrentStat)
	e.POST("/torrent/cache", torrentCache)
	e.POST("/torrent/drop", torrentDrop)

	e.GET("/torrent/restart", torrentRestart)

	e.GET("/torrent/playlist.m3u", torrentPlayListAll)

	e.GET("/torrent/play", torrentPlay)
	e.HEAD("/torrent/play", torrentPlay)

	e.GET("/torrent/view/:hash/:file", torrentView)
	e.HEAD("/torrent/view/:hash/:file", torrentView)
	e.GET("/torrent/preload/:hash/:file", torrentPreload)
	e.GET("/torrent/preload/:size/:hash/:file", torrentPreloadSize)
}

type TorrentJsonRequest struct {
	Link     string `json:",omitempty"`
	Hash     string `json:",omitempty"`
	Title    string `json:",omitempty"`
	Info     string `json:",omitempty"`
	DontSave bool   `json:",omitempty"`
}

type TorrentJsonResponse struct {
	Name     string
	Magnet   string
	Hash     string
	AddTime  int64
	Length   int64
	Status   torr.TorrentStatus
	Playlist string
	Info     string
	Files    []TorFile `json:",omitempty"`
}

type TorFile struct {
	Name    string
	Link    string
	Preload string
	Size    int64
	Viewed  bool
}

func torrentAdd(c echo.Context) error {
	jreq, err := getJsReqTorr(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if jreq.Link == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Link must be non-empty")
	}

	magnet, err := helpers.GetMagnet(jreq.Link)
	if err != nil {
		fmt.Println("Error get magnet:", jreq.Hash, err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if jreq.Title != "" {
		magnet.DisplayName = jreq.Title
	}

	err = helpers.Add(bts, *magnet, !jreq.DontSave)
	if err != nil {
		fmt.Println("Error add torrent:", jreq.Hash, err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if jreq.Info != "" {
		go func() {
			utils.AddInfo(magnet.InfoHash.HexString(), jreq.Info)
		}()
	}

	return c.String(http.StatusOK, magnet.InfoHash.HexString())
}

func torrentUpload(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer form.RemoveAll()

	_, dontSave := form.Value["DontSave"]
	var magnets []metainfo.Magnet

	for _, file := range form.File {
		torrFile, err := file[0].Open()
		if err != nil {
			return err
		}
		defer torrFile.Close()

		mi, err := metainfo.Load(torrFile)
		if err != nil {
			fmt.Println("Error upload torrent", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		info, err := mi.UnmarshalInfo()
		if err != nil {
			fmt.Println("Error upload torrent", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		magnet := mi.Magnet(info.Name, mi.HashInfoBytes())
		magnets = append(magnets, magnet)
	}

	ret := make([]string, 0)
	for _, magnet := range magnets {
		er := helpers.Add(bts, magnet, !dontSave)
		if er != nil {
			err = er
			fmt.Println("Error add torrent:", magnet.String(), er)
		}
		ret = append(ret, magnet.InfoHash.HexString())
	}

	return c.JSON(http.StatusOK, ret)
}

func torrentGet(c echo.Context) error {
	jreq, err := getJsReqTorr(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if jreq.Hash == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Hash must be non-empty")
	}

	tor, err := settings.LoadTorrentDB(jreq.Hash)
	if err != nil {
		fmt.Println("Error get torrent:", jreq.Hash, err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	torrStatus := torr.TorrentAdded
	if tor == nil {
		hash := metainfo.NewHashFromHex(jreq.Hash)
		ts := bts.GetTorrent(hash)
		if ts != nil {
			torrStatus = ts.Status()
			tor = toTorrentDB(ts)
		}
	}

	if tor == nil {
		fmt.Println("Error get: torrent not found", jreq.Hash)
		return echo.NewHTTPError(http.StatusBadRequest, "Error get: torrent not found "+jreq.Hash)
	}

	js, err := getTorrentJS(tor)
	if err != nil {
		fmt.Println("Error get torrent:", tor.Hash, err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	js.Status = torrStatus
	return c.JSON(http.StatusOK, js)
}

func torrentRem(c echo.Context) error {
	jreq, err := getJsReqTorr(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if jreq.Hash == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Hash must be non-empty")
	}

	settings.RemoveTorrentDB(jreq.Hash)
	bts.RemoveTorrent(metainfo.NewHashFromHex(jreq.Hash))

	return c.JSON(http.StatusOK, nil)
}

func torrentList(c echo.Context) error {
	buf, _ := ioutil.ReadAll(c.Request().Body)
	jsstr := string(buf)
	decoder := json.NewDecoder(bytes.NewBufferString(jsstr))
	jsreq := struct {
		Request int
	}{}
	decoder.Decode(&jsreq)

	reqType := jsreq.Request

	js := make([]TorrentJsonResponse, 0)
	list, _ := settings.LoadTorrentsDB()

	for _, tor := range list {
		jsTor, err := getTorrentJS(tor)
		if err != nil {
			fmt.Println("Error get torrent:", err)
		} else {
			js = append(js, *jsTor)
		}
	}

	sort.Slice(js, func(i, j int) bool {
		if js[i].AddTime == js[j].AddTime {
			return js[i].Name < js[j].Name
		}
		return js[i].AddTime > js[j].AddTime
	})

	slist := bts.List()

	find := func(tjs []TorrentJsonResponse, t *torr.Torrent) bool {
		for _, j := range tjs {
			if t.Hash().HexString() == j.Hash {
				return true
			}
		}
		return false
	}

	for _, st := range slist {
		if !find(js, st) {
			tdb := toTorrentDB(st)
			jsTor, err := getTorrentJS(tdb)
			if err != nil {
				fmt.Println("Error get torrent:", err)
			} else {
				jsTor.Status = st.Status()
				js = append(js, *jsTor)
			}
		}
	}

	if reqType == 1 {
		ret := make([]TorrentJsonResponse, 0)
		for _, r := range js {
			if r.Status == torr.TorrentWorking || len(r.Files) > 0 {
				ret = append(ret, r)
			}
		}
		return c.JSON(http.StatusOK, ret)
	} else if reqType == 2 {
		ret := make([]TorrentJsonResponse, 0)
		for _, r := range js {
			if r.Status == torr.TorrentGettingInfo {
				ret = append(ret, r)
			}
		}
		return c.JSON(http.StatusOK, ret)
	}

	return c.JSON(http.StatusOK, js)
}

func torrentStat(c echo.Context) error {
	jreq, err := getJsReqTorr(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if jreq.Hash == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Hash must be non-empty")
	}

	hash := metainfo.NewHashFromHex(jreq.Hash)
	tor := bts.GetTorrent(hash)
	if tor == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	stat := tor.Stats()

	return c.JSON(http.StatusOK, stat)
}

func torrentCache(c echo.Context) error {
	jreq, err := getJsReqTorr(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if jreq.Hash == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Hash must be non-empty")
	}

	hash := metainfo.NewHashFromHex(jreq.Hash)
	stat := bts.CacheState(hash)
	if stat == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, stat)
}

func preload(hashHex, fileLink string, size int64) *echo.HTTPError {
	if size > 0 {
		hash := metainfo.NewHashFromHex(hashHex)
		tor := bts.GetTorrent(hash)
		if tor == nil {
			torrDb, err := settings.LoadTorrentDB(hashHex)
			if err != nil || torrDb == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Torrent not found: "+hashHex)
			}
			m, err := metainfo.ParseMagnetURI(torrDb.Magnet)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Error parser magnet in db: "+hashHex)
			}
			tor, err = bts.AddTorrent(m, nil)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
		}

		if !tor.WaitInfo() {
			return echo.NewHTTPError(http.StatusBadRequest, "torrent closed befor get info")
		}

		file := helpers.FindFileLink(fileLink, tor.Torrent)
		if file == nil {
			return echo.NewHTTPError(http.StatusNotFound, "file in torrent not found: "+fileLink)
		}
		tor.Preload(file, size)
	}
	return nil
}

func torrentPreload(c echo.Context) error {
	hashHex, err := url.PathUnescape(c.Param("hash"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	fileLink, err := url.PathUnescape(c.Param("file"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if hashHex == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Hash must be non-empty")
	}
	if fileLink == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "File link must be non-empty")
	}

	errHttp := preload(hashHex, fileLink, settings.Get().PreloadBufferSize)
	if err != nil {
		return errHttp
	}

	return c.NoContent(http.StatusOK)
}

func torrentPreloadSize(c echo.Context) error {
	hashHex, err := url.PathUnescape(c.Param("hash"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	fileLink, err := url.PathUnescape(c.Param("file"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	szPreload, err := url.PathUnescape(c.Param("size"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if hashHex == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Hash must be non-empty")
	}
	if fileLink == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "File link must be non-empty")
	}

	var size = settings.Get().PreloadBufferSize
	if szPreload != "" {
		sz, err := strconv.Atoi(szPreload)
		if err == nil && sz > 0 {
			size = int64(sz) * 1024 * 1024
		}
	}

	errHttp := preload(hashHex, fileLink, size)
	if err != nil {
		return errHttp
	}
	//redirectUrl := c.Scheme() + "://" + c.Request().Host + filepath.Join("/torrent/view/", hashHex, fileLink)
	//return c.Redirect(http.StatusFound, redirectUrl)
	return c.NoContent(http.StatusOK)
}

func torrentDrop(c echo.Context) error {
	jreq, err := getJsReqTorr(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if jreq.Hash == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Hash must be non-empty")
	}

	bts.RemoveTorrent(metainfo.NewHashFromHex(jreq.Hash))
	return c.NoContent(http.StatusOK)
}

func torrentRestart(c echo.Context) error {
	fmt.Println("Restart torrent engine")
	err := bts.Reconnect()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Ok")
}

func torrentPlayListAll(c echo.Context) error {
	list, err := settings.LoadTorrentsDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	m3u := helpers.MakeM3ULists(list, c.Scheme()+"://"+c.Request().Host)

	c.Response().Header().Set("Content-Type", "audio/x-mpegurl")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="playlist.m3u"`)
	http.ServeContent(c.Response(), c.Request(), "playlist.m3u", time.Now(), bytes.NewReader([]byte(m3u)))
	return c.NoContent(http.StatusOK)
}

func torrentPlay(c echo.Context) error {
	link := c.QueryParam("link")
	if link == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "link should not be empty")
	}
	fmt.Println("Play:", c.QueryParams())

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

	magnet, err := helpers.GetMagnet(link)
	if err != nil {
		fmt.Println("Error get magnet:", link, err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tor := bts.GetTorrent(magnet.InfoHash)
	if tor == nil {
		tor, err = bts.AddTorrent(*magnet, nil)
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
		http.ServeContent(c.Response(), c.Request(), name, time.Time{}, bytes.NewReader([]byte(m3u)))
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
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprint("File", files[fileInd], "not found in torrent", tor.Name()))
	}
	return bts.Play(tor, file, preload, c)
}

func torrentView(c echo.Context) error {
	hashHex, err := url.PathUnescape(c.Param("hash"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	fileLink, err := url.PathUnescape(c.Param("file"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	hash := metainfo.NewHashFromHex(hashHex)
	tor := bts.GetTorrent(hash)
	if tor == nil {
		torrDb, err := settings.LoadTorrentDB(hashHex)
		if err != nil || torrDb == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Torrent not found: "+hashHex)
		}

		m, err := metainfo.ParseMagnetURI(torrDb.Magnet)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Error parser magnet in db: "+hashHex)
		}

		tor, err = bts.AddTorrent(m, nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	if !tor.WaitInfo() {
		return echo.NewHTTPError(http.StatusBadRequest, "torrent closed befor get info")
	}

	file := helpers.FindFileLink(fileLink, tor.Torrent)
	if file == nil {
		return echo.NewHTTPError(http.StatusNotFound, "File in torrent not found: "+fileLink)
	}
	return bts.View(tor, file, c)
}

func toTorrentDB(t *torr.Torrent) *settings.Torrent {
	if t == nil {
		return nil
	}
	tor := new(settings.Torrent)
	tor.Name = t.Name()
	tor.Hash = t.Hash().HexString()
	tor.Timestamp = settings.StartTime.Unix()
	mi := t.Torrent.Metainfo()
	tor.Magnet = mi.Magnet(t.Name(), t.Torrent.InfoHash()).String()
	tor.Size = t.Length()
	files := t.Files()
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path() < files[j].Path()
	})
	for _, f := range files {
		tf := settings.File{
			Name:   f.Path(),
			Size:   f.Length(),
			Viewed: false,
		}
		tor.Files = append(tor.Files, tf)
	}
	return tor
}

func getTorrentJS(tor *settings.Torrent) (*TorrentJsonResponse, error) {
	js := new(TorrentJsonResponse)
	mag, err := metainfo.ParseMagnetURI(tor.Magnet)
	js.Name = tor.Name
	if err == nil && len(tor.Name) < len(mag.DisplayName) {
		js.Name = mag.DisplayName
	}
	js.Magnet = tor.Magnet
	js.Hash = tor.Hash
	js.AddTime = tor.Timestamp
	js.Length = tor.Size
	//fname is fake param for file name
	js.Playlist = "/torrent/play?link=" + url.QueryEscape(tor.Magnet) + "&m3u=true&fname=" + utils.CleanFName(tor.Name+".m3u")
	var size int64 = 0
	for _, f := range tor.Files {
		size += f.Size
		tf := TorFile{
			Name:    f.Name,
			Link:    "/torrent/view/" + js.Hash + "/" + utils.CleanFName(f.Name),
			Preload: "/torrent/preload/" + js.Hash + "/" + utils.CleanFName(f.Name),
			Size:    f.Size,
			Viewed:  f.Viewed,
		}
		js.Files = append(js.Files, tf)
	}
	if tor.Size == 0 {
		js.Length = size
	}

	js.Info = settings.GetInfo(tor.Hash)

	return js, nil
}

func getJsReqTorr(c echo.Context) (*TorrentJsonRequest, error) {
	buf, _ := ioutil.ReadAll(c.Request().Body)
	jsstr := string(buf)
	decoder := json.NewDecoder(bytes.NewBufferString(jsstr))
	js := new(TorrentJsonRequest)
	err := decoder.Decode(js)
	if err != nil {
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, offset=%v", ute.Type, ute.Value, ute.Offset))
		} else if se, ok := err.(*json.SyntaxError); ok {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error()))
		} else {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
	return js, nil
}
