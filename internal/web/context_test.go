package web

import (
	"net"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// newTestContext создаёт gin-контекст с заданным адресом отправителя.
func newTestContext(remoteIP string) *gin.Context {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/api/internal/source/test", nil)
	c.Request.RemoteAddr = net.JoinHostPort(remoteIP, "55555")
	return c
}
