package web

import (
	"net/http"
	"strings"

	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

const CookieTokenName = "silo_token"

// MaxRequestBody - предел входящего тела запроса. Выше самого большого
// из существующих лимитов (архив плагина 50 МБ), чтобы их не срезать.
// На отдачу данных не влияет: стриминг и транскодирование - это ответ.
const MaxRequestBody = 64 << 20

// BodyLimitMiddleware ограничивает размер тела запроса. Чтение сверх лимита
// завершается ошибкой, обработчик отвечает 400: без предела запрос
// с большим телом вычитывается в память целиком.
func BodyLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestBody)
		}
		c.Next()
	}
}

// RecoveryMiddleware перехватывает панику в обработчике: клиент получает 500,
// а в лог со стеком попадает место падения, в том числе из плагинов.
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, r any) {
		log.Panic("[Web] Handler panicked: "+c.Request.Method+" "+c.Request.URL.Path, r)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

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
