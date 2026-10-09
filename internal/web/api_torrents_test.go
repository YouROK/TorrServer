package web

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
	torr "silo/internal/torrent"
	"silo/internal/torrent/storage/torrstor"
	"silo/internal/user"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"
)

// setupAddTestEnv поднимает Server с движком и менеджером для проверки добавления
func setupAddTestEnv(t *testing.T) (*Server, *user.User) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(filepath.Join(t.TempDir(), "add_test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cfg := config.DefaultConfig()
	userSvc := user.NewService(user.NewStore(db), cfg)

	owner, err := userSvc.Authenticate("")
	if err != nil {
		t.Fatalf("authenticate owner: %v", err)
	}

	engineCfg := torr.DefaultConfig()
	engineCfg.ListenPort = 0
	engineCfg.DisableDHT = true
	engineCfg.DisablePEX = true
	engineCfg.Storage = &torrstor.Config{Capacity: 10 << 20, UseDisk: false}

	engine, err := torr.NewEngine(engineCfg)
	if err != nil {
		t.Fatalf("start engine: %v", err)
	}

	mgr := torr.NewManager(engine, torr.NewStore(db), userSvc)
	t.Cleanup(func() { mgr.Close() })

	return &Server{cfg: cfg, userSvc: userSvc, torrentMgr: mgr, router: gin.New()}, owner
}

// torrentFileBytes собирает настоящий .torrent файл с одним файлом внутри
func torrentFileBytes(t *testing.T, name string) []byte {
	t.Helper()

	info := metainfo.Info{
		PieceLength: 32 * 1024,
		Name:        name,
		Length:      64 * 1024,
	}
	info.Pieces = make([]byte, 2*20)

	infoBytes, err := bencode.Marshal(info)
	if err != nil {
		t.Fatalf("bencode info: %v", err)
	}

	mi := metainfo.MetaInfo{InfoBytes: infoBytes}
	var buf bytes.Buffer
	if err := mi.Write(&buf); err != nil {
		t.Fatalf("write torrent: %v", err)
	}
	return buf.Bytes()
}

// TestHandleAddTorrentFromFile проверяет добавление раздачи загруженным .torrent файлом
func TestHandleAddTorrentFromFile(t *testing.T) {
	s, owner := setupAddTestEnv(t)

	raw := torrentFileBytes(t, "Uploaded_Movie_2024")

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "movie.torrent")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(raw); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = writer.WriteField("title", "")
	_ = writer.WriteField("poster", "http://poster.jpg")
	_ = writer.WriteField("category", "Movies")
	writer.Close()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/torrents", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user", owner)

	s.handleAddTorrent(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var status struct {
		Hash  string `json:"hash"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Имя берется из info-словаря .torrent файла, а не из формы
	if status.Title != "Uploaded_Movie_2024" {
		t.Errorf("expected name from torrent file, got: %q", status.Title)
	}

	infoHash := metainfo.Hash(sha1.Sum(mustUnmarshalInfo(t, raw)))
	if status.Hash != infoHash.HexString() {
		t.Errorf("hash mismatch: got %s, want %s", status.Hash, infoHash.HexString())
	}

	// Раздача сохранена в библиотеке пользователя
	if _, err := s.userSvc.GetUserTorrent(owner.ID, status.Hash); err != nil {
		t.Errorf("torrent missing from user library: %v", err)
	}
}

// TestHandleAddTorrentFileInvalid проверяет отказ на мусорный файл
func TestHandleAddTorrentFileInvalid(t *testing.T) {
	s, owner := setupAddTestEnv(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "broken.torrent")
	_, _ = part.Write([]byte("this is not a torrent"))
	writer.Close()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/torrents", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user", owner)

	s.handleAddTorrent(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestHandleAddTorrentJsonStillWorks проверяет, что JSON-запросы продолжают работать
func TestHandleAddTorrentJsonStillWorks(t *testing.T) {
	s, owner := setupAddTestEnv(t)

	raw := torrentFileBytes(t, "Json_Movie")
	infoHash := metainfo.Hash(sha1.Sum(mustUnmarshalInfo(t, raw)))

	payload, _ := json.Marshal(map[string]any{
		"link":       infoHash.HexString(),
		"title":      "Custom Title",
		"save_to_db": true,
	})

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/torrents", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", owner)

	s.handleAddTorrent(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var status struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if status.Title != "Custom Title" {
		t.Errorf("expected user title, got: %q", status.Title)
	}
}

// mustUnmarshalInfo разбирает info-словарь из собранного .torrent файла
func mustUnmarshalInfo(t *testing.T, raw []byte) []byte {
	t.Helper()

	mi, err := metainfo.Load(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("load metainfo: %v", err)
	}
	return mi.InfoBytes
}
