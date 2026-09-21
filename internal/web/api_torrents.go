package web

import (
	"errors"
	"io"
	"net/http"
	torr "silo/internal/torrent"
	"sort"
	"strconv"
	"strings"

	"silo/internal/log"
	"silo/internal/torrshash"
	"silo/internal/user"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"
)

func (s *Server) handleListTorrents(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	list, err := s.torrentMgr.ListTorrents(currentUser)
	if err != nil {
		log.Errorf("[Web] Failed to list torrents for user %s: %v", currentUser.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch torrents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"torrents": list})
}

// isHexHash проверяет, что строка является 40-символьным hex info-хэшем
func isHexHash(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// handleAddTorrent принимает magnet-ссылку, голый info-hash или torrs:// ссылку
func (s *Server) handleAddTorrent(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	var req struct {
		Link     string `json:"link" binding:"required"`
		Title    string `json:"title"`
		Poster   string `json:"poster"`
		Category string `json:"category"`
		SaveToDB bool   `json:"save_to_db"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	link := strings.TrimSpace(req.Link)
	title, poster, category := req.Title, req.Poster, req.Category
	var trackers []string
	var spec *torrent.TorrentSpec

	switch {
	// torrs:// — наш упакованный формат со всеми полями
	case strings.HasPrefix(link, "torrs://"):
		th, err := torrshash.Unpack(strings.TrimPrefix(link, "torrs://"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid torrs link"})
			return
		}
		if title == "" {
			title = th.Title()
		}
		if poster == "" {
			poster = th.Poster()
		}
		if category == "" {
			category = th.Category()
		}
		trackers = th.Trackers()
		spec = &torrent.TorrentSpec{
			InfoHash:    metainfo.NewHashFromHex(th.Hash),
			DisplayName: title,
		}

	// magnet: — стандартная ссылка
	case strings.HasPrefix(link, "magnet:"):
		mi, err := metainfo.ParseMagnetURI(link)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid magnet link"})
			return
		}
		trackers = mi.Trackers
		spec = &torrent.TorrentSpec{
			InfoHash:    mi.InfoHash,
			DisplayName: mi.DisplayName,
		}

	// Голый info-hash
	case isHexHash(link):
		spec = &torrent.TorrentSpec{
			InfoHash:    metainfo.NewHashFromHex(link),
			DisplayName: title,
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported link format: use magnet, hash or torrs://"})
		return
	}

	// Плоский список трекеров оборачиваем в тиры
	var tiers [][]string
	for _, tr := range trackers {
		tiers = append(tiers, []string{tr})
	}
	spec.Trackers = tiers

	status, err := s.torrentMgr.AddTorrent(currentUser, spec, title, poster, category, req.SaveToDB)
	if err != nil {
		log.Errorf("[Web] Failed to add torrent: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, status)
}

// handleExportLibrary отдает библиотеку пользователя файлом .torrs
func (s *Server) handleExportLibrary(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	lines, err := s.torrentMgr.ExportLibrary(currentUser)
	if err != nil {
		log.Errorf("[Web] Failed to export library: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export library"})
		return
	}

	body := strings.Join(lines, "\n") + "\n"
	c.Header("Content-Disposition", `attachment; filename="silo-library.torrs"`)
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
}

// handleImportLibrary принимает текст или файл .torrs и добавляет раздачи в библиотеку
func (s *Server) handleImportLibrary(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	var text string
	if file, _, err := c.Request.FormFile("file"); err == nil {
		data, _ := io.ReadAll(io.LimitReader(file, 5*1024*1024))
		file.Close()
		text = string(data)
	} else {
		data, err := io.ReadAll(io.LimitReader(c.Request.Body, 5*1024*1024))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}
		text = string(data)
	}

	imported, err := s.torrentMgr.ImportLibrary(currentUser, strings.Split(text, "\n"))
	if err != nil {
		log.Errorf("[Web] Failed to import library: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to import library"})
		return
	}

	log.Infof("[Web] User %s imported %d torrent(s)", currentUser.Username, imported)
	c.JSON(http.StatusOK, gin.H{"status": "imported", "count": imported})
}

// handleUpdateTorrentMeta правит личную карточку раздачи (Edit в контекстном меню)
func (s *Server) handleUpdateTorrentMeta(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)
	hashHex := c.Param("hash")

	var req struct {
		Title    string `json:"title"`
		Poster   string `json:"poster"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := s.userSvc.UpdateTorrentMeta(currentUser, hashHex, req.Title, req.Poster, req.Category); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) handleGetTorrent(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)
	hashHex := c.Param("hash")

	status, err := s.torrentMgr.GetTorrentStatus(currentUser, hashHex)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (s *Server) handleDropTorrent(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)
	hashHex := c.Param("hash")

	if err := s.torrentMgr.RemoveTorrent(currentUser, hashHex); err != nil {
		log.Errorf("[Web] Failed to drop torrent %s: %v", hashHex, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (s *Server) handleSetFileViewed(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)
	hashHex := c.Param("hash")
	fileIdxStr := c.Param("idx")

	fileIdx, err := strconv.Atoi(fileIdxStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file index"})
		return
	}

	var req struct {
		Viewed bool `json:"viewed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := s.torrentMgr.SetFileViewed(currentUser, hashHex, fileIdx, req.Viewed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// handleWakeTorrent явно будит торрент в движок (только владелец карточки)
func (s *Server) handleWakeTorrent(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)
	hashHex := c.Param("hash")

	err := s.torrentMgr.WakeTorrent(currentUser, hashHex)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "wakeup"})
}

// handlePreloadTorrent запускает предзагрузку файла раздачи
func (s *Server) handlePreloadTorrent(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)
	hashHex := c.Param("hash")

	idx, err := strconv.Atoi(c.Param("idx"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file index"})
		return
	}

	err = s.torrentMgr.PreloadTorrent(currentUser, hashHex, idx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "preloaded"})
}

// cacheResponse — sparse-ответ для вкладки Cache.
// Передаются только заполненные куски: ids + параллельные base64-массивы.
type cacheResponse struct {
	Hash        string        `json:"hash"`
	Capacity    int64         `json:"capacity"`
	Filled      int64         `json:"filled"`
	PieceLength int64         `json:"piece_length"`
	PieceCount  int           `json:"piece_count"`
	IDs         []int         `json:"ids"`
	Sizes       []int         `json:"sizes"`
	Priorities  []int         `json:"priorities"`
	Readers     []cacheReader `json:"readers"`
}

type cacheReader struct {
	Start  int `json:"start"`
	End    int `json:"end"`
	Reader int `json:"reader"`
}

func (s *Server) handleGetCache(c *gin.Context) {
	val, _ := c.Get("user")
	currentUser := val.(*user.User)
	hashHex := c.Param("hash")

	state, err := s.torrentMgr.GetCacheState(currentUser, hashHex)
	if err != nil {
		if errors.Is(err, torr.ErrSessionNotInRAM) {
			c.JSON(http.StatusNoContent, nil)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Собираем индексы заполненных кусков, отбрасывая выходящие за границы.
	ids := make([]int, 0, len(state.Pieces))
	for id := range state.Pieces {
		if id >= 0 && id < state.PiecesCount {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)

	sizes := make([]int, len(ids))
	priorities := make([]int, len(ids))
	for i, id := range ids {
		item := state.Pieces[id]
		if state.PiecesLength > 0 {
			p := int(item.Size * 255 / state.PiecesLength)
			if p > 255 {
				p = 255
			}
			if p < 0 {
				p = 0
			}
			sizes[i] = p
		}
		priorities[i] = item.Priority
	}

	readers := make([]cacheReader, 0, len(state.Readers))
	for _, r := range state.Readers {
		readers = append(readers, cacheReader{
			Start:  r.Start,
			End:    r.End,
			Reader: r.Reader,
		})
	}

	c.JSON(http.StatusOK, cacheResponse{
		Hash:        state.Hash,
		Capacity:    state.Capacity,
		Filled:      state.Filled,
		PieceLength: state.PiecesLength,
		PieceCount:  state.PiecesCount,
		IDs:         ids,
		Sizes:       sizes,
		Priorities:  priorities,
		Readers:     readers,
	})
}
