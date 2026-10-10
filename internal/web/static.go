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

// RegisterStaticRoutes монтирует общие стили, скрипты и иконки на публичные
// роуты /css/*, /js/* и /img/*. Они доступны всем без авторизации:
// админке, плагинам и темам.
func RegisterStaticRoutes(r *gin.Engine) {
	sub := func(name string) (fs.FS, bool) {
		dir, err := fs.Sub(staticFS, "static/"+name)
		if err != nil {
			log.Errorf("[Web] failed to open embedded %s fs: %v", name, err)
			return nil, false
		}
		return dir, true
	}

	cssFS, ok := sub("css")
	if !ok {
		return
	}
	jsFS, ok := sub("js")
	if !ok {
		return
	}
	imgFS, ok := sub("img")
	if !ok {
		return
	}

	// http.FS сам проставляет Content-Type и поддерживает Range/If-Modified
	r.StaticFS("/css", http.FS(cssFS))
	r.StaticFS("/js", http.FS(jsFS))
	r.StaticFS("/img", http.FS(imgFS))

	r.StaticFileFS("/login", "static/login.html", http.FS(staticFS))

	log.Info("[Web] Shared static routes '/css', '/js' and '/img' registered")
}
