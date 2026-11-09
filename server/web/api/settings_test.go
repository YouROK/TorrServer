package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// "set" without "sets" used to pass nil to torr.SetSettings and panic.
func TestSettingsSetWithoutSetsIsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/settings", settings)

	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(`{"action":"set"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), `"error"`) {
		t.Errorf("no error in body: %q", w.Body.String())
	}
}
