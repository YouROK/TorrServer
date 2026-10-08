package web

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"testing"
	"time"

	"silo/internal/config"
	"silo/internal/database"
	torr "silo/internal/torrent"
	"silo/internal/torrent/storage/torrstor"
	"silo/internal/user"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"
)

// synthSpec создаёт тестовую раздачу целиком в памяти (без сети).
func synthSpec(t *testing.T) *torrent.TorrentSpec {
	t.Helper()

	info := metainfo.Info{
		PieceLength: 32 * 1024,
		Name:        "Test_Series",
		Files: []metainfo.FileInfo{
			{Length: 64 * 1024, Path: []string{"S01E01.mkv"}},
			{Length: 64 * 1024, Path: []string{"S01E02.mkv"}},
		},
	}

	total := int64(128 * 1024)
	numPieces := (total + info.PieceLength - 1) / info.PieceLength
	info.Pieces = make([]byte, numPieces*20)

	infoBytes, err := bencode.Marshal(info)
	if err != nil {
		t.Fatalf("bencode failed: %v", err)
	}
	h := sha1.Sum(infoBytes)

	return &torrent.TorrentSpec{
		InfoBytes:   infoBytes,
		InfoHash:    metainfo.Hash(h),
		DisplayName: "Test Series",
	}
}

// setupStreamTestEnv поднимает движок, менеджер и Server для проверки стриминга.
func setupStreamTestEnv(t *testing.T) (*Server, *torr.Engine, *user.Service, *user.User, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(filepath.Join(t.TempDir(), "stream_test.db"))
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
	// DHT выключен: в тестах сеть не нужна, а фоновый DHT-клиент библиотеки
	// torrent конфликтует с Close() (предсуществующая гонка в самой библиотеке,
	// воспроизводится и в пакете internal/torrent без нашего кода).
	engineCfg.DisableDHT = true
	engineCfg.DisablePEX = true
	engineCfg.Storage = &torrstor.Config{Capacity: 10 << 20, UseDisk: false}

	engine, err := torr.NewEngine(engineCfg)
	if err != nil {
		t.Fatalf("start engine: %v", err)
	}

	// Manager.Close() сам закрывает движок, поэтому отдельный cleanup для engine
	// не регистрируем: повторное закрытие даёт гонку внутри библиотеки torrent.
	mgr := torr.NewManager(engine, torr.NewStore(db), userSvc)
	t.Cleanup(func() { mgr.Close() })

	s := &Server{
		cfg:           cfg,
		userSvc:       userSvc,
		torrentMgr:    mgr,
		router:        gin.New(),
		streamTracker: NewStreamTracker(),
	}

	spec := synthSpec(t)
	hashHex := spec.InfoHash.HexString()
	if _, err := mgr.AddTorrent(owner, spec, "Test Series", "", "Series", true); err != nil {
		t.Fatalf("add torrent: %v", err)
	}

	return s, engine, userSvc, owner, hashHex
}

// adminTorrentsResponse повторяет форму ответа handleAdminListTorrents.
type adminTorrentsResponse struct {
	Torrents []struct {
		Hash        string           `json:"torrent_hash"`
		StreamCount int              `json:"stream_count"`
		Streams     []StreamSnapshot `json:"streams"`
	} `json:"torrents"`
}

// loadAdminTorrents вызывает handleAdminListTorrents и разбирает ответ.
func loadAdminTorrents(t *testing.T, s *Server, actor *user.User, targetID string) adminTorrentsResponse {
	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/users/"+targetID+"/torrents", nil)
	c.Params = gin.Params{{Key: "id", Value: targetID}}
	c.Set("user", actor)

	s.handleAdminListTorrents(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("handleAdminListTorrents status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var parsed adminTorrentsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("decode admin torrents: %v", err)
	}
	return parsed
}

// Пока запрос стрима «висит» (плеер читает), поток обязан быть виден в админском API
// с именем файла, раздачей и IP клиента, а после завершения - исчезнуть.
func TestStreamVisibleInAdminWhileOpen(t *testing.T) {
	s, engine, _, owner, hash := setupStreamTestEnv(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)

		req := httptest.NewRequest(http.MethodGet, "/api/stream/"+hash+"/0", nil).WithContext(ctx)
		req.RemoteAddr = "203.0.113.9:5555"
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		req.Header.Set("User-Agent", "TestPlayer/1.0")

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = req
		c.Params = gin.Params{{Key: "hash", Value: hash}, {Key: "fileIdx", Value: "0"}}
		c.Set("user", owner)

		s.handleStream(c)
	}()

	// Ждём, пока поток зарегистрируется.
	deadline := time.Now().Add(5 * time.Second)
	for s.streamTracker.Count() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if got := s.streamTracker.Count(); got == 0 {
		cancel()
		t.Fatalf("stream was not registered in tracker")
	}

	// Проверяем то, что увидит админка.
	admin := loadAdminTorrents(t, s, owner, owner.ID)
	if len(admin.Torrents) == 0 {
		t.Fatalf("admin torrents list is empty")
	}

	var found bool
	for _, tr := range admin.Torrents {
		if tr.Hash != hash {
			continue
		}
		found = true

		if tr.StreamCount != 1 {
			t.Errorf("StreamCount = %d, want 1", tr.StreamCount)
		}
		if len(tr.Streams) != 1 {
			t.Fatalf("Streams = %+v, want 1 entry", tr.Streams)
		}

		st := tr.Streams[0]
		if st.FileName != "S01E01.mkv" {
			t.Errorf("FileName = %q, want S01E01.mkv", st.FileName)
		}
		// Реальный адрес соединения и адрес из заголовка показываются раздельно.
		if st.ClientIP != "203.0.113.9" {
			t.Errorf("ClientIP = %q, want 203.0.113.9", st.ClientIP)
		}
		if st.ForwardedIP != "198.51.100.7" {
			t.Errorf("ForwardedIP = %q, want 198.51.100.7", st.ForwardedIP)
		}
		if st.UserAgent != "TestPlayer/1.0" {
			t.Errorf("UserAgent = %q", st.UserAgent)
		}
		if st.Hash != hash {
			t.Errorf("Hash = %q, want %q", st.Hash, hash)
		}
		if st.UserID != owner.ID {
			t.Errorf("UserID = %q, want %q", st.UserID, owner.ID)
		}
	}

	if !found {
		t.Fatalf("torrent %s not present in admin list", hash)
	}

	// Снимаем поток так же, как это делает defer в handleStream, и убеждаемся,
	// что админка сразу перестаёт его показывать.
	s.streamTracker.mu.Lock()
	for id := range s.streamTracker.streams {
		delete(s.streamTracker.streams, id)
	}
	s.streamTracker.mu.Unlock()

	if got := s.streamTracker.Count(); got != 0 {
		t.Errorf("tracker still holds %d stream(s) after close", got)
	}

	after := loadAdminTorrents(t, s, owner, owner.ID)
	for _, tr := range after.Torrents {
		if tr.Hash == hash && tr.StreamCount != 0 {
			t.Errorf("StreamCount = %d after close, want 0", tr.StreamCount)
		}
	}

	// Без пиров ServeContent блокируется на чтении, поэтому горутина handleStream
	// сама не завершится. Останавливаем сессию, чтобы разбудить читателя, и ждём
	// горутину: иначе cleanup ниже закроет движок под работающим запросом (гонка).
	cancel()
	engine.Stop(metainfo.NewHashFromHex(hash))

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Errorf("stream request did not finish after session stop")
	}
}

// Список раздач обязан отдавать живые поля (статус, размер, факт наличия в RAM),
// даже когда потоков нет - админка рисует по ним таблицу.
func TestAdminTorrentsHasLiveFields(t *testing.T) {
	s, _, _, owner, hash := setupStreamTestEnv(t)

	admin := loadAdminTorrents(t, s, owner, owner.ID)
	if len(admin.Torrents) == 0 {
		t.Fatalf("admin torrents list is empty")
	}

	var found bool
	for _, tr := range admin.Torrents {
		if tr.Hash != hash {
			continue
		}
		found = true
		if tr.StreamCount != 0 {
			t.Errorf("StreamCount = %d, want 0 when nobody watches", tr.StreamCount)
		}
		if len(tr.Streams) != 0 {
			t.Errorf("Streams = %v, want empty", tr.Streams)
		}
	}
	if !found {
		t.Fatalf("torrent %s not present in admin list", hash)
	}
}

// forwardedIP вытаскивает первый адрес из X-Forwarded-For и X-Real-IP.
func TestForwardedIPParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		header string
		value  string
		want   string
	}{
		{"xff single", "X-Forwarded-For", "203.0.113.5", "203.0.113.5"},
		{"xff chain", "X-Forwarded-For", "203.0.113.5, 10.0.0.1, 10.0.0.2", "203.0.113.5"},
		{"real ip", "X-Real-IP", "198.51.100.9", "198.51.100.9"},
		{"absent", "", "", ""},
		{"spaces", "X-Forwarded-For", "  203.0.113.7  ", "203.0.113.7"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
			if tc.header != "" {
				c.Request.Header.Set(tc.header, tc.value)
			}

			if got := forwardedIP(c); got != tc.want {
				t.Errorf("forwardedIP() = %q, want %q", got, tc.want)
			}
		})
	}
}
