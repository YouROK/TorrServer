package web

import (
	"net/http"
	"strings"

	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

const CookieTokenName = "silo_token"

func AuthMiddleware(userSvc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Токен из параметра URL: ?token=ts_xxx (для плееров)
		token := c.Query("token")

		// Токен из заголовка Authorization: Bearer ts_xxx (для API)
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Токен из Cookie: silo_token (для браузера)
		if token == "" {
			token, _ = c.Cookie(CookieTokenName)
		}

		// Проверяем пользователя
		u, err := userSvc.Authenticate(token)
		if err != nil {
			log.Debugf("[Web] Authentication failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Сохраняем в контекст для последующих хэндлеров
		c.Set("user", u)
		c.Next()
	}
}
