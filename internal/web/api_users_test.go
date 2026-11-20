package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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

// Cookie с токеном должна уходить с SameSite=Lax: это защищает
// изменяющие состояние ручки от межсайтовых запросов (CSRF).
func TestTokenCookieHasSameSiteLax(t *testing.T) {
	s, _ := setupAuthTestEnv(t, "supersecret")
	s.registerRoutes()

	body := strings.NewReader(`{"username":"owner","password":"supersecret"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var found bool
	for _, ck := range rec.Result().Cookies() {
		if ck.Name != CookieTokenName {
			continue
		}
		found = true
		if ck.SameSite != http.SameSiteLaxMode {
			t.Errorf("cookie SameSite = %v, want Lax", ck.SameSite)
		}
		if !ck.HttpOnly {
			t.Errorf("cookie must stay HttpOnly")
		}
	}
	if !found {
		t.Fatalf("cookie %q not set on login", CookieTokenName)
	}
}

// Ответ на вход не должен содержать bcrypt-хэш пароля, а токен
// возвращается только самому пользователю.
func TestLoginResponseOmitsPasswordHash(t *testing.T) {
	s, svc := setupAuthTestEnv(t, "supersecret")
	s.registerRoutes()

	owner, err := svc.GetUserByID("owner")
	if err != nil {
		t.Fatalf("get owner: %v", err)
	}
	if _, err := svc.CreateUser(owner, "alice", "alicepass", user.RankUser, user.Limits{}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	body := strings.NewReader(`{"username":"alice","password":"alicepass"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "$2a$") {
		t.Fatalf("login response leaks bcrypt hash: %s", rec.Body.String())
	}

	var resp struct {
		User  user.User `json:"user"`
		Token string    `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.User.PasswordHash != "" {
		t.Errorf("password_hash = %q, want empty", resp.User.PasswordHash)
	}
	if resp.Token == "" || resp.User.APIToken != resp.Token {
		t.Errorf("api_token = %q, token = %q, want the same non-empty token", resp.User.APIToken, resp.Token)
	}
}

// publicUser не должен менять исходного пользователя: он общий для запроса.
func TestPublicUserKeepsOriginal(t *testing.T) {
	u := &user.User{ID: "u1", PasswordHash: "hash", APIToken: "ts_token"}

	got := publicUser(u, false)
	if got.PasswordHash != "" || got.APIToken != "" {
		t.Errorf("publicUser(false) = %+v, want no secrets", got)
	}
	if withToken := publicUser(u, true); withToken.APIToken != "ts_token" || withToken.PasswordHash != "" {
		t.Errorf("publicUser(true) = %+v, want token and no hash", withToken)
	}
	if u.PasswordHash != "hash" || u.APIToken != "ts_token" {
		t.Errorf("original user modified: %+v", u)
	}
}
