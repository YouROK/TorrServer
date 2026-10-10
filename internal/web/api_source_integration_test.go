package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/ffmpeg/source"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// newLeaseTestServer создаёт сервер без торрент-движка: проверкам доступа
// и разбора адресов он не нужен.
func newLeaseTestServer(t *testing.T) (*Server, *user.User) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(t.TempDir() + "/lease_test.db")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cfg := config.DefaultConfig()
	svc := user.NewService(user.NewStore(db), cfg)

	owner, err := svc.Authenticate("")
	if err != nil {
		t.Fatalf("authenticate owner: %v", err)
	}

	s := &Server{
		cfg:            cfg,
		userSvc:        svc,
		router:         gin.New(),
		streamTracker:  NewStreamTracker(),
		sourceRegistry: source.NewRegistry(0),
	}
	s.registerRoutes()

	return s, owner
}

// callSource вызывает handleInternalSource с заданным адресом отправителя.
func callSource(s *Server, id, remoteIP string, headers map[string]string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/internal/source/"+id, nil)
	c.Request.RemoteAddr = remoteIP + ":55555"
	for k, v := range headers {
		c.Request.Header.Set(k, v)
	}
	c.Params = gin.Params{{Key: "id", Value: id}}

	s.handleInternalSource(c)
	return rec
}

func TestInternalSourceRejectsUnknownLease(t *testing.T) {
	s, _ := newLeaseTestServer(t)

	rec := callSource(s, "deadbeefdeadbeefdeadbeefdeadbeef", "127.0.0.1", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestInternalSourceRejectsExternalAddress(t *testing.T) {
	s, owner := newLeaseTestServer(t)

	lease := s.sourceRegistry.Issue(owner.ID, "aabbccddeeff00112233445566778899aabbccdd", 0)

	for _, ip := range []string{"203.0.113.9", "192.168.1.50", "10.0.0.7"} {
		t.Run(ip, func(t *testing.T) {
			rec := callSource(s, lease.ID, ip, nil)
			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want 404 for %s", rec.Code, ip)
			}
		})
	}
}

func TestInternalSourceIgnoresForwardedHeaders(t *testing.T) {
	s, owner := newLeaseTestServer(t)

	lease := s.sourceRegistry.Issue(owner.ID, "aabbccddeeff00112233445566778899aabbccdd", 0)

	// Подделка X-Forwarded-For не должна давать доступ с внешнего адреса
	rec := callSource(s, lease.ID, "203.0.113.9", map[string]string{
		"X-Forwarded-For": "127.0.0.1",
		"X-Real-IP":       "127.0.0.1",
	})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404: forwarded headers must not grant access", rec.Code)
	}
}

func TestInternalSourceUnknownUser(t *testing.T) {
	s, _ := newLeaseTestServer(t)

	lease := s.sourceRegistry.Issue("missing-user", "aabbccddeeff00112233445566778899aabbccdd", 0)

	rec := callSource(s, lease.ID, "127.0.0.1", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for an unknown user", rec.Code)
	}
}

func TestInternalSourceRouteRegistered(t *testing.T) {
	s, _ := newLeaseTestServer(t)

	var found bool
	for _, r := range s.router.Routes() {
		if r.Path == "/api/internal/source/:id" {
			found = true
			if r.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", r.Method)
			}
		}
	}
	if !found {
		t.Error("internal source route is not registered")
	}
}

func TestInternalSourceNotInAuthGroup(t *testing.T) {
	// Роут должен отвечать на запрос без токена, иначе ffmpeg не сможет его прочитать.
	s, _ := newLeaseTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/internal/source/unknown", nil)
	req.RemoteAddr = "127.0.0.1:55555"

	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Error("internal source must not require a token")
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for an unknown lease", rec.Code)
	}
}

func TestSourceURLShape(t *testing.T) {
	s, _ := newLeaseTestServer(t)

	url := s.SourceURL("abc123")
	if !strings.HasPrefix(url, "http://") {
		t.Errorf("url %q must be absolute", url)
	}
	if !strings.HasSuffix(url, "/api/internal/source/abc123") {
		t.Errorf("url %q has an unexpected shape", url)
	}
}
