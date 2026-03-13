package tgbot

import (
	"strings"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"
	"server/dlna"
	"server/rutor"
	"server/settings"
	"server/torr"
)

type pendingPreset struct {
	Sets      *settings.BTSets
	Preset    string // name for display
	UserID    int64
	IsDef     bool
	CreatedAt time.Time
}

var (
	pendingPresetMu sync.Mutex
	pendingPresets  = make(map[string]pendingPreset)
)

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			pendingPresetMu.Lock()
			now := time.Now()
			for key, p := range pendingPresets {
				if now.Sub(p.CreatedAt) > 30*time.Minute {
					delete(pendingPresets, key)
				}
			}
			pendingPresetMu.Unlock()
		}
	}()
}

func cmdPreset(c tele.Context) error {
	uid := c.Sender().ID
	if !isAdmin(uid) {
		return c.Send(tr(uid, "admin_only"))
	}
	if settings.BTsets == nil {
		return c.Send(tr(uid, "settings_not_loaded"))
	}
	if settings.ReadOnly {
		return c.Send(tr(uid, "settings_readonly"))
	}

	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return c.Send(tr(uid, "preset_usage"))
	}

	sets := new(settings.BTSets)
	*sets = *settings.BTsets

	first := strings.ToLower(args[1])
	presetName := first

	if len(args) == 2 {
		if ok, _ := applyNamedPreset(sets, first, uid); ok {
			return sendPresetConfirm(c, uid, sets, presetName, false)
		}
		if first == "default" || first == "def" || first == "сброс" {
			return sendPresetConfirm(c, uid, nil, "default", true)
		}
	}

	// Parse key-value pairs: cache 256 preload 50 conn 100
	applied, errMsg := applyPresetKV(sets, args[1:], uid)
	if !applied {
		return c.Send(errMsg)
	}
	presetName = strings.Join(args[1:], " ")
	return sendPresetConfirm(c, uid, sets, presetName, false)
}

func sendPresetConfirm(c tele.Context, uid int64, sets *settings.BTSets, presetName string, isDef bool) error {
	btnYes := tele.InlineButton{Text: tr(uid, "btn_yes"), Unique: "fpreset", Data: "1"}
	btnNo := tele.InlineButton{Text: tr(uid, "btn_no"), Unique: "fpreset", Data: "0"}
	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnYes, btnNo}}}
	msg := tr(uid, "preset_confirm") + "\n\n<code>" + presetName + "</code>"
	sent, err := c.Bot().Send(c.Chat(), msg, kbd, tele.ModeHTML)
	if err != nil {
		return err
	}
	pendingPresetMu.Lock()
	pendingPresets[chatMsgKey(sent.Chat.ID, sent.ID)] = pendingPreset{
		Sets: sets, Preset: presetName, UserID: uid, IsDef: isDef, CreatedAt: time.Now(),
	}
	pendingPresetMu.Unlock()
	return nil
}

func presetConfirm(c tele.Context, confirm string) error {
	uid := c.Sender().ID
	if !isAdmin(uid) {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "admin_only")})
	}
	key := chatMsgKey(c.Callback().Message.Chat.ID, c.Callback().Message.ID)
	pendingPresetMu.Lock()
	p, ok := pendingPresets[key]
	delete(pendingPresets, key)
	pendingPresetMu.Unlock()
	if !ok || p.UserID != uid {
		_ = c.Bot().Delete(c.Callback().Message)
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "canceled")})
	}
	if confirm != "1" {
		_ = c.Bot().Delete(c.Callback().Message)
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "canceled")})
	}
	_ = c.Bot().Delete(c.Callback().Message)
	if settings.ReadOnly {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_readonly")})
	}
	if p.IsDef {
		torr.SetDefSettings()
		dlna.Stop()
		rutor.Stop()
		rutor.Start()
		return c.Send(tr(uid, "settings_reset_done"))
	}
	if p.Sets == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "callback_unknown")})
	}
	torr.SetSettings(p.Sets)
	dlna.Stop()
	if p.Sets.EnableDLNA {
		dlna.Start()
	}
	rutor.Stop()
	rutor.Start()
	return c.Send(tr(uid, "preset_applied") + p.Preset)
}

func applyNamedPreset(s *settings.BTSets, name string, uid int64) (bool, string) {
	switch name {
	case "performance", "perf", "производительность":
		s.CacheSize = 512 * 1024 * 1024
		s.PreloadCache = 95
		s.ReaderReadAHead = 100
		s.ConnectionsLimit = 100
		s.TorrentDisconnectTimeout = 60
		s.PeersListenPort = 0
		s.DownloadRateLimit = 0
		s.UploadRateLimit = 0
		s.RetrackersMode = 1
		s.ResponsiveMode = true
		return true, tr(uid, "preset_applied") + " performance"
	case "storage", "store", "хранение":
		s.CacheSize = 64 * 1024 * 1024
		s.PreloadCache = 25
		s.ReaderReadAHead = 50
		s.RemoveCacheOnDrop = true
		return true, tr(uid, "preset_applied") + " storage"
	case "streaming", "stream", "стриминг":
		s.CacheSize = 256 * 1024 * 1024
		s.PreloadCache = 75
		s.ReaderReadAHead = 95
		s.ConnectionsLimit = 50
		s.ResponsiveMode = true
		return true, tr(uid, "preset_applied") + " streaming"
	case "low", "minimal", "минимум":
		s.CacheSize = 64 * 1024 * 1024
		s.PreloadCache = 25
		s.ReaderReadAHead = 50
		s.ConnectionsLimit = 25
		s.TorrentDisconnectTimeout = 30
		return true, tr(uid, "preset_applied") + " low"
	case "default", "def", "сброс":
		return false, "" // handled in cmdPreset
	}
	return false, ""
}

func applyPresetKV(s *settings.BTSets, args []string, uid int64) (bool, string) {
	if len(args) < 2 {
		return false, tr(uid, "preset_usage")
	}
	applied := false
	for i := 0; i < len(args)-1; i += 2 {
		key := strings.ToLower(args[i])
		val := strings.ToLower(strings.TrimSpace(args[i+1]))
		ok := false
		switch key {
		case "cache":
			if v := parseInt(val); v > 0 {
				s.CacheSize = int64(v) * 1024 * 1024
				ok = true
			}
		case "preload":
			if v := parseInt(val); v >= 0 && v <= 100 {
				s.PreloadCache = v
				ok = true
			}
		case "readahead":
			if v := parseInt(val); v >= 5 && v <= 100 {
				s.ReaderReadAHead = v
				ok = true
			}
		case "conn", "connections":
			if v := parseInt(val); v > 0 {
				s.ConnectionsLimit = v
				ok = true
			}
		case "timeout":
			if v := parseInt(val); v > 0 {
				s.TorrentDisconnectTimeout = v
				ok = true
			}
		case "port":
			v := parseInt(val)
			if val == "auto" || val == "0" {
				v = 0
			}
			if v >= 0 && (v == 0 || (v >= 1024 && v <= 65535)) {
				s.PeersListenPort = v
				ok = true
			}
		case "down", "download":
			v := 0
			if val != "inf" && val != "∞" && val != "0" {
				v = parseInt(val)
			}
			s.DownloadRateLimit = v
			ok = true
		case "up", "upload":
			v := 0
			if val != "inf" && val != "∞" && val != "0" {
				v = parseInt(val)
			}
			s.UploadRateLimit = v
			ok = true
		case "retr", "retrackers":
			var v int
			switch val {
			case "off":
				v = 0
			case "add":
				v = 1
			case "rem", "remove":
				v = 2
			case "repl", "replace":
				v = 3
			default:
				v = parseInt(val)
			}
			if v >= 0 && v <= 3 {
				s.RetrackersMode = v
				ok = true
			}
		case "responsive":
			s.ResponsiveMode = val == "1" || val == "on" || val == "true" || val == "да" || val == "yes"
			ok = true
		case "cachedrop":
			s.RemoveCacheOnDrop = val == "1" || val == "on" || val == "true" || val == "да" || val == "yes"
			ok = true
		}
		if ok {
			applied = true
		}
	}
	if !applied {
		return false, tr(uid, "preset_usage")
	}
	return true, ""
}
