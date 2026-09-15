package admin

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

//go:embed all:static
var staticFS embed.FS

// ContextKeyUser — ключ авторизованного пользователя в gin.Context
const ContextKeyUser = "user"

// Пороговые ранги доступа к разделам админки
const (
	RankAdmin = 50
	RankOwner = 100
)

// RegisterRoutes монтирует встроенную админку.
// Стили и иконки больше не отдаются отсюда: они общие и живут на /css и /img.
func RegisterRoutes(r *gin.Engine, userSvc *user.Service) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Errorf("[Admin] failed to open embedded static fs: %v", err)
		return
	}

	// Публична только страница входа
	r.GET("/admin/login", serveFile(sub, "login.html"))

	// Редирект с /admin на /admin/ для корректных относительных путей
	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/")
	})

	// Защищенная часть: оболочка SPA и её скрипт
	grp := r.Group("/admin", browserAuth(userSvc), RankMiddleware(RankAdmin))
	{
		grp.GET("/", serveFile(sub, "index.html"))
		grp.GET("/index.html", serveFile(sub, "index.html"))
		grp.GET("/app.js", serveFile(sub, "app.js"))
	}

	log.Info("[Admin] Protected admin interface registered at /admin")
}

// browserAuth извлекает токен и при неудаче редиректит на /admin/login
func browserAuth(userSvc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			if authHeader := c.GetHeader("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if token == "" {
			token, _ = c.Cookie("silo_token")
		}

		u, err := userSvc.Authenticate(token)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
			c.Abort()
			return
		}

		c.Set(ContextKeyUser, u)
		c.Next()
	}
}

// RankMiddleware пропускает только пользователей с рангом не ниже minRank
func RankMiddleware(minRank int) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(ContextKeyUser)
		if !exists {
			c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
			c.Abort()
			return
		}

		u, ok := val.(*user.User)
		if !ok || int(u.Rank) < minRank {
			// Недостаточно прав: редирект на логин с кодом ошибки
			c.Redirect(http.StatusTemporaryRedirect, "/admin/login?error=forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

// serveFile отдает один статический файл из embed-FS с правильным Content-Type
func serveFile(fsys fs.FS, name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, contentTypeByExt(path.Ext(name)), data)
	}
}

// contentTypeByExt возвращает MIME-тип по расширению файла
func contentTypeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	}
	return "application/octet-stream"
}
