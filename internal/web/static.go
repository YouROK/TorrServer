package web

import (
	"embed"
	"io/fs"
	"net/http"

	"silo/internal/log"

	"github.com/gin-gonic/gin"
)

//go:embed all:static
var staticFS embed.FS

// RegisterStaticRoutes монтирует общие стили и иконки на публичные роуты
// /css/* и /img/*. Они доступны всем без авторизации: админке, плагинам и темам.
func RegisterStaticRoutes(r *gin.Engine) {
	cssFS, err := fs.Sub(staticFS, "static/css")
	if err != nil {
		log.Errorf("[Web] failed to open embedded css fs: %v", err)
		return
	}

	imgFS, err := fs.Sub(staticFS, "static/img")
	if err != nil {
		log.Errorf("[Web] failed to open embedded img fs: %v", err)
		return
	}

	// http.FS сам проставляет Content-Type и поддерживает Range/If-Modified
	r.StaticFS("/css", http.FS(cssFS))
	r.StaticFS("/img", http.FS(imgFS))

	log.Info("[Web] Shared static routes '/css' and '/img' registered")
}
