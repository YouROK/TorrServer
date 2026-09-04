package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/gin-gonic/gin"

	"server/log"
	"server/settings"
)

func SetupAuth(engine *gin.Engine) {
	if !settings.HttpAuth {
		return
	}
	accs := getAccounts()
	if accs == nil {
		return
	}
	engine.Use(BasicAuth(accs))
}

func getAccounts() gin.Accounts {
	buf, err := os.ReadFile(filepath.Join(settings.Path, "accs.db"))
	if err != nil {
		return nil
	}
	var accs gin.Accounts
	err = json.Unmarshal(buf, &accs)
	if err != nil {
		log.TLogln("Error parse accs.db", err)
	}
	return accs
}

type authPair struct {
	value string
	user  string
}
type authPairs []authPair

func (a authPairs) searchCredential(authValue string) (string, bool) {
	if authValue == "" {
		return "", false
	}
	for _, pair := range a {
		if pair.value == authValue {
			return pair.user, true
		}
	}
	return "", false
}

func BasicAuth(accounts gin.Accounts) gin.HandlerFunc {
	pairs := processAccounts(accounts)
	return func(c *gin.Context) {
		c.Set("auth_required", true)

		user, found := pairs.searchCredential(c.Request.Header.Get("Authorization"))
		if found {
			c.Set(gin.AuthUserKey, user)
			return
		}

		if initData := c.GetHeader("X-Telegram-Init-Data"); initData != "" {
			if tgUser, ok := TryTelegramAuth(initData); ok {
				c.Set(gin.AuthUserKey, tgUser)
			}
		}
	}
}

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !settings.HttpAuth {
			return
		}

		if _, ok := c.Get(gin.AuthUserKey); ok {
			return
		}

		if initData := c.GetHeader("X-Telegram-Init-Data"); initData != "" {
			if user, ok := TryTelegramAuth(initData); ok {
				c.Set(gin.AuthUserKey, user)
				return
			}
		}

		// SPA XHR/fetch probes must not trigger the browser's native Basic dialog.
		// Third-party clients that already send Authorization never hit this path.
		if !shouldSendBasicChallenge(c) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// shouldSendBasicChallenge is false for typical SPA/API clients (JSON Accept,
// X-Requested-With, or Sec-Fetch-Mode: cors) so axios can show a custom login UI.
func shouldSendBasicChallenge(c *gin.Context) bool {
	accept := strings.ToLower(c.GetHeader("Accept"))
	if strings.Contains(accept, "application/json") || strings.Contains(accept, "+json") {
		return false
	}
	if c.GetHeader("X-Requested-With") != "" {
		return false
	}
	if strings.EqualFold(c.GetHeader("Sec-Fetch-Mode"), "cors") {
		return false
	}
	return true
}

func processAccounts(accounts gin.Accounts) authPairs {
	pairs := make(authPairs, 0, len(accounts))
	for user, password := range accounts {
		value := authorizationHeader(user, password)
		pairs = append(pairs, authPair{
			value: value,
			user:  user,
		})
	}
	return pairs
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString(StringToBytes(base))
}

func StringToBytes(s string) (b []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
