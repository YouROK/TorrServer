package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"
)

func TestInternalSourceServesTorrentFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping torrent-backed source test in short mode")
	}

	s, engine, _, owner, hash := setupStreamTestEnv(t)

	lease := s.sourceRegistry.Issue(owner.ID, hash, 0)

	// Без пиров чтение синтетического файла блокируется, поэтому запрос
	// запускается в горутине и освобождается остановкой сессии.
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/internal/source/"+lease.ID, nil)
	c.Request.RemoteAddr = "127.0.0.1:55555"
	c.Params = gin.Params{{Key: "id", Value: lease.ID}}

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handleInternalSource(c)
	}()

	// Ждём, пока источник зарегистрирует читателя в движке
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if sess, ok := engine.Get(metainfo.NewHashFromHex(hash)); ok && sess.ActiveReaders() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	engine.Stop(metainfo.NewHashFromHex(hash))

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("internal source request did not complete")
	}

	if rec.Code == http.StatusNotFound {
		t.Fatalf("internal source returned 404 for a valid lease (body: %s)", rec.Body.String())
	}
}

func TestInternalSourceUnknownFileIdx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping torrent-backed source test in short mode")
	}

	s, _, _, owner, hash := setupStreamTestEnv(t)

	lease := s.sourceRegistry.Issue(owner.ID, hash, 99)

	rec := serveSourceRequest(s, lease.ID, "127.0.0.1")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for an out-of-range file", rec.Code)
	}
}

// serveSourceRequest вызывает handleInternalSource напрямую.
func serveSourceRequest(s *Server, id, remoteIP string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/internal/source/"+id, nil)
	c.Request.RemoteAddr = remoteIP + ":55555"
	c.Params = gin.Params{{Key: "id", Value: id}}

	s.handleInternalSource(c)
	return rec
}
