package auth

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"server/settings"
)

func TestInternalListenerAuthBoundary(t *testing.T) {
	oldAuth, oldPort := settings.HttpAuth, settings.InternalPort
	t.Cleanup(func() { settings.HttpAuth, settings.InternalPort = oldAuth, oldPort })
	settings.HttpAuth, settings.InternalPort = true, "54321"
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(BasicAuth(gin.Accounts{"viewer": "test-password"}))
	router.GET("/media", CheckAuth(), func(c *gin.Context) { c.String(http.StatusOK, "media") })

	for _, tc := range []struct {
		name     string
		remote   string
		local    net.Addr
		port     string
		wantCode int
	}{
		{"internal loopback", "127.0.0.1:12345", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}, "54321", 200},
		{"internal IPv6 peer", "[::1]:12345", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}, "54321", 200},
		{"local proxy on public listener", "127.0.0.1:12345", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8090}, "54321", 401},
		{"remote peer", "192.0.2.1:12345", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}, "54321", 401},
		{"non-loopback listener", "127.0.0.1:12345", &net.TCPAddr{IP: net.ParseIP("192.0.2.1"), Port: 54321}, "54321", 401},
		{"missing server address", "127.0.0.1:12345", nil, "54321", 401},
		{"inactive internal listener", "127.0.0.1:12345", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}, "", 401},
		{"malformed peer", "not-an-address", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}, "54321", 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings.InternalPort = tc.port
			request := httptest.NewRequest(http.MethodGet, "/media", nil)
			request.RemoteAddr = tc.remote
			// Client-controlled forwarding headers must never grant the internal exemption.
			request.Header.Set("X-Forwarded-For", "127.0.0.1")
			request.Header.Set("X-Forwarded-Host", "127.0.0.1:54321")
			if tc.local != nil {
				request = request.WithContext(context.WithValue(request.Context(), http.LocalAddrContextKey, tc.local))
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.wantCode {
				t.Fatalf("status=%d, want%d", response.Code, tc.wantCode)
			}
		})
	}

	request := httptest.NewRequest(http.MethodGet, "/media", nil)
	request.SetBasicAuth("viewer", "test-password")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "media" {
		t.Fatalf("valid public credentials: %d %q", response.Code, response.Body.String())
	}
}
