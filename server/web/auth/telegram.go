package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"server/settings"
	"server/tgbot/config"
)

const telegramInitDataMaxAge = 24 * time.Hour

// ValidateTelegramInitData checks Telegram Mini App initData HMAC and freshness.
// See https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func ValidateTelegramInitData(initData, botToken string) (userID int64, ok bool) {
	if initData == "" || botToken == "" {
		return 0, false
	}
	values, err := url.ParseQuery(initData)
	if err != nil {
		return 0, false
	}
	hash := values.Get("hash")
	if hash == "" {
		return 0, false
	}
	values.Del("hash")

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+values.Get(k))
	}
	dataCheck := strings.Join(parts, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(dataCheck))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(hash)) {
		return 0, false
	}

	if authDateStr := values.Get("auth_date"); authDateStr != "" {
		authDate, err := strconv.ParseInt(authDateStr, 10, 64)
		if err != nil {
			return 0, false
		}
		if time.Since(time.Unix(authDate, 0)) > telegramInitDataMaxAge {
			return 0, false
		}
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return 0, false
	}
	// Minimal parse: "id":123456 without full JSON dependency for robustness
	id := extractJSONInt64Field(userJSON, "id")
	if id == 0 {
		return 0, false
	}
	return id, true
}

func extractJSONInt64Field(jsonStr, field string) int64 {
	key := "\"" + field + "\":"
	idx := strings.Index(jsonStr, key)
	if idx < 0 {
		return 0
	}
	rest := jsonStr[idx+len(key):]
	rest = strings.TrimLeft(rest, " \t")
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0
	}
	n, err := strconv.ParseInt(rest[:end], 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// IsTelegramUserAllowed mirrors bot access: BlackIds block; empty WhiteIds allow all.
func IsTelegramUserAllowed(userID int64) bool {
	if config.Cfg == nil {
		config.LoadConfig()
	}
	if config.Cfg == nil {
		return true
	}
	for _, id := range config.Cfg.BlackIds {
		if id == userID {
			return false
		}
	}
	if len(config.Cfg.WhiteIds) == 0 {
		return true
	}
	for _, id := range config.Cfg.WhiteIds {
		if id == userID {
			return true
		}
	}
	return false
}

func telegramBotToken() string {
	if settings.Args != nil {
		return settings.Args.TGToken
	}
	return ""
}

// TryTelegramAuth validates X-Telegram-Init-Data and returns username key if allowed.
func TryTelegramAuth(initData string) (userKey string, ok bool) {
	token := telegramBotToken()
	if token == "" {
		return "", false
	}
	uid, valid := ValidateTelegramInitData(initData, token)
	if !valid || !IsTelegramUserAllowed(uid) {
		return "", false
	}
	return "tg:" + strconv.FormatInt(uid, 10), true
}
