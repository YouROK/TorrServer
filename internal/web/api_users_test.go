package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// setupAuthTestEnv поднимает минимальный Server с настоящим user.Service
// и роутом регенерации токена, защищённым тестовой подстановкой актора.
func setupAuthTestEnv(t *testing.T, ownerPassword string) (*Server, *user.Service) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(filepath.Join(t.TempDir(), "web_test.db"))
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cfg := config.DefaultConfig()
	cfg.Auth.OwnerPassword = ownerPassword

	svc := user.NewService(user.NewStore(db), cfg)
	s := &Server{cfg: cfg, userSvc: svc, router: gin.New()}

	return s, svc
}

// callRegenerate вызывает handleRegenerateToken от имени actor.
func callRegenerate(t *testing.T, s *Server, actor *user.User, targetID string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/users/"+targetID+"/regenerate-token", nil)
	c.Params = gin.Params{{Key: "id", Value: targetID}}
	c.Set("user", actor)

	s.handleRegenerateToken(c)
	return rec
}

func tokenFromSetCookie(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	for _, ck := range rec.Result().Cookies() {
		if ck.Name == CookieTokenName {
			return ck.Value
		}
	}
	return ""
}

// Смена токена самому себе обязана обновить cookie, иначе пользователя
// сразу выкидывает на /login с уже недействительной сессией.
func TestRegenerateOwnTokenRefreshesCookie(t *testing.T) {
	s, svc := setupAuthTestEnv(t, "supersecret")

	owner, err := svc.GetUserByID("owner")
	if err != nil {
		t.Fatalf("Failed to get owner: %v", err)
	}

	rec := callRegenerate(t, s, owner, owner.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	cookieToken := tokenFromSetCookie(t, rec)
	if cookieToken == "" {
		t.Fatalf("Cookie %q was not refreshed on self regeneration", CookieTokenName)
	}
	if cookieToken != body.Token {
		t.Fatalf("Cookie holds stale token:\n  cookie  = %q\n  response= %q", cookieToken, body.Token)
	}

	// Cookie обязана содержать рабочий токен, а не просто совпадать с ответом.
	if _, err := svc.Authenticate(cookieToken); err != nil {
		t.Fatalf("Token stored in cookie is not authenticatable: %v", err)
	}
}

// Смена токена ДРУГОМУ пользователю не должна трогать cookie админа.
func TestRegenerateOtherUserLeavesCookieAlone(t *testing.T) {
	s, svc := setupAuthTestEnv(t, "supersecret")

	owner, err := svc.GetUserByID("owner")
	if err != nil {
		t.Fatalf("Failed to get owner: %v", err)
	}
	admin, err := svc.CreateUser(owner, "admin_y", "pass", user.RankAdmin, user.Limits{})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	rec := callRegenerate(t, s, owner, admin.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if cookieToken := tokenFromSetCookie(t, rec); cookieToken != "" {
		t.Errorf("Cookie must not change when regenerating another user's token, got %q", cookieToken)
	}
}

// Health-check /api/system/ping обязан быть публичным: он отвечает на вопрос
// «жив ли сервер», а не «авторизован ли клиент». Иначе при истёкшей куке он
// отдавал 401, и фронтенд рисовал «сервер недоступен» на работающем сервере.
// Остальные роуты при этом должны остаться защищёнными.
func TestPingIsPublicOthersProtected(t *testing.T) {
	s, _ := setupAuthTestEnv(t, "supersecret")
	s.registerRoutes()

	cases := []struct {
		path string
		want int
	}{
		{"/api/system/ping", http.StatusOK},
		{"/api/auth/me", http.StatusUnauthorized},
		{"/api/users", http.StatusUnauthorized},
		{"/api/system/logs", http.StatusUnauthorized},
	}

	for _, tc := range cases {
		rec := httptest.NewRecorder()
		s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != tc.want {
			t.Errorf("%s: got %d, want %d", tc.path, rec.Code, tc.want)
		}
	}
}
