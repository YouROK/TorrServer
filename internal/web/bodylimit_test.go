package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestBodyLimitMiddleware проверяет, что большое тело запроса не вычитывается
// целиком, а нормальный запрос проходит без изменений.
func TestBodyLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(BodyLimitMiddleware())
	engine.POST("/echo", func(c *gin.Context) {
		var req struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "too large or invalid"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"len": len(req.Text)})
	})

	// Обычный запрос проходит
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"text":"ok"}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("normal request status = %d, want 200", rec.Code)
	}

	// Тело сверх лимита не принимается
	big := `{"text":"` + strings.Repeat("a", MaxRequestBody+1024) + `"}`
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(big))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Errorf("oversized request status = %d, want an error", rec.Code)
	}
}
