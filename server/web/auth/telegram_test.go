package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"server/settings"
	"server/tgbot/config"
)

func signInitData(token string, vals url.Values) string {
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+vals.Get(k))
	}
	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secretMAC.Write([]byte(token))
	mac := hmac.New(sha256.New, secretMAC.Sum(nil))
	_, _ = mac.Write([]byte(strings.Join(parts, "\n")))
	vals.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return vals.Encode()
}

func TestValidateTelegramInitData(t *testing.T) {
	token := "123:ABC"
	now := time.Now().Unix()
	vals := url.Values{}
	vals.Set("auth_date", strconv.FormatInt(now, 10))
	vals.Set("user", `{"id":42,"first_name":"A"}`)
	data := signInitData(token, vals)
	id, ok := ValidateTelegramInitData(data, token)
	if !ok || id != 42 {
		t.Fatalf("valid %d %v", id, ok)
	}
	if _, ok := ValidateTelegramInitData(data, "other"); ok {
		t.Fatal("wrong token")
	}
	if _, ok := ValidateTelegramInitData("", token); ok {
		t.Fatal("empty")
	}
	if _, ok := ValidateTelegramInitData("hash=abc", token); ok {
		t.Fatal("bad hash")
	}
	old := url.Values{}
	old.Set("auth_date", "1000")
	old.Set("user", `{"id":42}`)
	if _, ok := ValidateTelegramInitData(signInitData(token, old), token); ok {
		t.Fatal("expired")
	}
	noUser := url.Values{}
	noUser.Set("auth_date", strconv.FormatInt(now, 10))
	if _, ok := ValidateTelegramInitData(signInitData(token, noUser), token); ok {
		t.Fatal("no user")
	}
	zero := url.Values{}
	zero.Set("auth_date", strconv.FormatInt(now, 10))
	zero.Set("user", `{"id":0}`)
	if _, ok := ValidateTelegramInitData(signInitData(token, zero), token); ok {
		t.Fatal("id 0")
	}
	badDate := url.Values{}
	badDate.Set("auth_date", "nope")
	badDate.Set("user", `{"id":7}`)
	if _, ok := ValidateTelegramInitData(signInitData(token, badDate), token); ok {
		t.Fatal("bad date")
	}

	oldCfg := config.Cfg
	config.Cfg = &config.Config{WhiteIds: []int64{42}, BlackIds: []int64{7}}
	t.Cleanup(func() { config.Cfg = oldCfg })
	if !IsTelegramUserAllowed(42) || IsTelegramUserAllowed(7) || IsTelegramUserAllowed(9) {
		t.Fatal("whitelist")
	}
	config.Cfg.WhiteIds = nil
	if !IsTelegramUserAllowed(9) {
		t.Fatal("open")
	}

	oldArgs := settings.Args
	settings.Args = &settings.ExecArgs{TGToken: token}
	t.Cleanup(func() { settings.Args = oldArgs })
	key, ok := TryTelegramAuth(data)
	if !ok || key != "tg:42" {
		t.Fatalf("try %q %v", key, ok)
	}
	settings.Args.TGToken = ""
	if _, ok := TryTelegramAuth(data); ok {
		t.Fatal("no token")
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = ValidateTelegramInitData(data, token)
		}()
	}
	wg.Wait()
}

func TestBasicAuthCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BasicAuth(gin.Accounts{"alice": "secret"}))
	r.GET("/p", func(c *gin.Context) {
		u, _ := c.Get(gin.AuthUserKey)
		c.String(http.StatusOK, "%v", u)
	})
	raw := base64.StdEncoding.EncodeToString([]byte("alice:secret"))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Basic "+raw)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "alice") {
		t.Fatalf("auth %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("alice:nope")))
	r.ServeHTTP(w, req)
	if strings.Contains(w.Body.String(), "alice") {
		t.Fatalf("rejected user leaked: %s", w.Body.String())
	}
}
