package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	sets "server/settings"
	"server/web/auth"
)

func TestMediaRoutesRequireHTTPAuth(t *testing.T) {
	oldAuth, oldSearch := sets.HttpAuth, sets.SearchWA
	t.Cleanup(func() { sets.HttpAuth, sets.SearchWA = oldAuth, oldSearch })
	sets.HttpAuth, sets.SearchWA = true, false
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.BasicAuth(gin.Accounts{"viewer": "test-password"}))
	SetupRoute(router)

	// Empty/invalid media parameters deliberately avoid starting the BT engine.
	// Authentication must run before media lookup or validation on every route.
	for _, route := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/stream"},
		{http.MethodHead, "/stream"},
		{http.MethodGet, "/stream/file.mp4?play"},
		{http.MethodHead, "/stream/file.mp4?play"},
		{http.MethodGet, "/stream/file.m3u?m3u"},
		{http.MethodGet, "/play/invalid/1"},
		{http.MethodHead, "/play/invalid/1"},
		{http.MethodGet, "/subs/invalid/1"},
		{http.MethodGet, "/playlist"},
		{http.MethodGet, "/playlist/file.m3u"},
	} {
		for _, credentials := range []string{"missing", "wrong"} {
			t.Run(route.method+route.path+"/"+credentials, func(t *testing.T) {
				request := httptest.NewRequest(route.method, route.path, nil)
				request.Header.Set("Range", "bytes=0-9")
				if credentials == "wrong" {
					request.SetBasicAuth("viewer", "wrong-password")
				}
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				if response.Code != http.StatusUnauthorized {
					t.Fatalf("status = %d, want 401 before media handler", response.Code)
				}
				if response.Header().Get("WWW-Authenticate") == "" {
					t.Fatal("missing Basic authentication challenge")
				}
			})
		}
	}

	// A correct credential must reach the actual handler, not an unconditional deny.
	for _, path := range []string{"/stream", "/playlist"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.SetBasicAuth("viewer", "test-password")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("authorized %s: status = %d, want handler's 400", path, response.Code)
		}
	}
}

func TestMediaRoutesWithoutHTTPAuthKeepHandlerBehavior(t *testing.T) {
	oldAuth := sets.HttpAuth
	t.Cleanup(func() { sets.HttpAuth = oldAuth })
	sets.HttpAuth = false
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoute(router)
	for _, path := range []string{"/stream", "/playlist"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("auth disabled %s: status = %d, want handler's 400", path, response.Code)
		}
	}
}
