package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleI18nJS(c *gin.Context) {
	if s.pluginMgr == nil {
		c.Header("Cache-Control", "no-cache")
		stub := "var i18n={t:function(k,d){return typeof d==='string'?d:k;},get:function(){return 'en';},set:function(){},langs:function(){return[];},apply:function(){}};window.i18n=i18n;"
		c.Data(http.StatusOK, "application/javascript; charset=utf-8", []byte(stub))
		return
	}

	data, ver := s.pluginMgr.I18n().BuildJS()
	etag := fmt.Sprintf(`W/"i18n-%d"`, ver)

	c.Header("ETag", etag)
	c.Header("Cache-Control", "no-cache, must-revalidate")

	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Data(http.StatusOK, "application/javascript; charset=utf-8", data)
}
