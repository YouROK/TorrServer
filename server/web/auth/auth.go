package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
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
