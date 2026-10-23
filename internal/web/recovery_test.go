package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRecoveryMiddlewareReturns500 проверяет, что паника в обработчике
// не роняет сервер: клиент получает 500, обработка продолжается.
func TestRecoveryMiddlewareReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(RecoveryMiddleware())
	engine.GET("/boom", func(c *gin.Context) {
		panic("handler failed")
	})
	engine.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "alive")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("panicking handler status = %d, want 500", rec.Code)
	}

	// Сервер продолжает обслуживать запросы
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ok", nil)
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("next request status = %d, want 200", rec.Code)
	}
}
