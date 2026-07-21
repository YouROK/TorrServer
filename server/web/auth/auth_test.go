package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"server/settings"
)

func TestCheckAuthSkipsChallengeForSPAClients(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settings.HttpAuth = true
	t.Cleanup(func() { settings.HttpAuth = false })

	r := gin.New()
	r.GET("/probe", CheckAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	cases := []struct {
		name      string
		headers   map[string]string
		wantChall bool
	}{
		{
			name:      "json accept",
			headers:   map[string]string{"Accept": "application/json"},
			wantChall: false,
		},
		{
			name:      "x-requested-with",
			headers:   map[string]string{"X-Requested-With": "XMLHttpRequest"},
			wantChall: false,
		},
		{
			name:      "cors fetch",
			headers:   map[string]string{"Sec-Fetch-Mode": "cors"},
			wantChall: false,
		},
		{
			name:      "plain navigation",
			headers:   map[string]string{"Accept": "text/html"},
			wantChall: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status %d, want 401", w.Code)
			}
			chall := w.Header().Get("WWW-Authenticate")
			if tc.wantChall && chall == "" {
				t.Fatal("expected WWW-Authenticate challenge")
			}
			if !tc.wantChall && chall != "" {
				t.Fatalf("unexpected challenge %q", chall)
			}
		})
	}
}

func TestCheckAuthAllowsAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settings.HttpAuth = true
	t.Cleanup(func() { settings.HttpAuth = false })

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(gin.AuthUserKey, "alice")
		c.Next()
	})
	r.GET("/probe", CheckAuth(), func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
}
