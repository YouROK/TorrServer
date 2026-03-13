package tgbot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"
	"server/settings"
)

const (
	LangRU            = "ru"
	LangEN            = "en"
	saveUserLangsWait = 2 * time.Second
)

var (
	userLang           = make(map[int64]string)
	userLangMu         sync.RWMutex
	saveUserLangsMu    sync.Mutex
	saveUserLangsTimer *time.Timer
)

func getUserLang(userID int64) string {
	userLangMu.RLock()
	defer userLangMu.RUnlock()
	if lang, ok := userLang[userID]; ok {
		return lang
	}
	return LangRU
}

func setUserLang(userID int64, lang string) {
	if lang != LangRU && lang != LangEN {
		return
	}
	userLangMu.Lock()
	userLang[userID] = lang
	userLangMu.Unlock()
	scheduleSaveUserLangs()
}

func scheduleSaveUserLangs() {
	saveUserLangsMu.Lock()
	defer saveUserLangsMu.Unlock()
	if saveUserLangsTimer != nil {
		saveUserLangsTimer.Stop()
	}
	saveUserLangsTimer = time.AfterFunc(saveUserLangsWait, func() {
		saveUserLangsMu.Lock()
		saveUserLangsTimer = nil
		saveUserLangsMu.Unlock()
		saveUserLangs()
	})
}

func loadUserLangs() {
	fn := filepath.Join(settings.Path, "tg_langs.json")
	buf, err := os.ReadFile(fn)
	if err != nil {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	userLangMu.Lock()
	for k, v := range m {
		if v == LangRU || v == LangEN {
			if id, parseErr := strconv.ParseInt(k, 10, 64); parseErr == nil {
				userLang[id] = v
			}
		}
	}
	userLangMu.Unlock()
}

func saveUserLangs() {
	userLangMu.RLock()
	m := make(map[string]string)
	for k, v := range userLang {
		m[strconv.FormatInt(k, 10)] = v
	}
	userLangMu.RUnlock()
	buf, err := json.Marshal(m)
	if err != nil {
		return
	}
	fn := filepath.Join(settings.Path, "tg_langs.json")
	_ = os.WriteFile(fn, buf, 0o600)
}

func cmdLang(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) == 0 {
		lang := getUserLang(uid)
		if lang == LangEN {
			return c.Send(tr(uid, "lang_current_en") + "\n/lang RU — " + tr(uid, "lang_switch_ru"))
		}
		return c.Send(tr(uid, "lang_current_ru") + "\n/lang EN — " + tr(uid, "lang_switch_en"))
	}
	lang := strings.ToUpper(strings.TrimSpace(args[0]))
	if lang == "EN" {
		setUserLang(uid, LangEN)
		return c.Send(tr(uid, "lang_set_en"))
	}
	if lang == "RU" || lang == "РУ" {
		setUserLang(uid, LangRU)
		return c.Send(tr(uid, "lang_set"))
	}
	return c.Send(tr(uid, "lang_usage"))
}
