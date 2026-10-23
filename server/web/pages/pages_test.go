package pages

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMagnetsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/magnets", nil)
	getTorrents(c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "<div>") || strings.Contains(w.Body.String(), "magnet:") {
		t.Fatalf("%d %q", w.Code, w.Body.String())
	}
}
