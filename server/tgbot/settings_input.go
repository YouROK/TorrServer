package tgbot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"
	"server/dlna"
	"server/rutor"
	"server/settings"
	"server/torr"
	"server/torznab"
)

const pendingInputTTL = 30 * time.Minute

type pendingInput struct {
	Setting   string
	UserID    int64
	CreatedAt time.Time
}

var (
	pendingInputMu sync.Mutex
	pendingInputs  = make(map[string]pendingInput)
)

func init() {
	go pendingInputCleanup()
}

func pendingInputCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		pendingInputMu.Lock()
		now := time.Now()
		for key, p := range pendingInputs {
			if now.Sub(p.CreatedAt) > pendingInputTTL {
				delete(pendingInputs, key)
			}
		}
		pendingInputMu.Unlock()
	}
}

func sendSettingsInputPrompt(c tele.Context, uid int64, setting, hint string) error {
	msg := fmt.Sprintf("✏️ %s\n\n%s", tr(uid, "settings_input_reply"), hint)
	btnCancel := tele.InlineButton{Text: "❌ " + tr(uid, "canceled"), Unique: "fset", Data: "input_cancel"}
	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnCancel}}}
	sent, err := c.Bot().Send(c.Chat(), msg, kbd)
	if err != nil {
		return err
	}
	pendingInputMu.Lock()
	pendingInputs[chatMsgKey(sent.Chat.ID, sent.ID)] = pendingInput{Setting: setting, UserID: uid, CreatedAt: time.Now()}
	pendingInputMu.Unlock()
	return c.Respond(&tele.CallbackResponse{})
}

func handleSettingsInputReply(c tele.Context) (handled bool) {
	msg := c.Message()
	if msg.ReplyTo == nil {
		return false
	}
	key := chatMsgKey(msg.ReplyTo.Chat.ID, msg.ReplyTo.ID)
	pendingInputMu.Lock()
	pending, ok := pendingInputs[key]
	delete(pendingInputs, key)
	pendingInputMu.Unlock()
	if !ok || pending.UserID != msg.Sender.ID {
		return false
	}
	if time.Since(pending.CreatedAt) > pendingInputTTL {
		_ = c.Send(tr(msg.Sender.ID, "canceled"))
		return true
	}
	applySettingsInput(c, pending.Setting, strings.TrimSpace(msg.Text))
	return true
}

func cancelSettingsInput(c tele.Context) error {
	key := chatMsgKey(c.Callback().Message.Chat.ID, c.Callback().Message.ID)
	pendingInputMu.Lock()
	delete(pendingInputs, key)
	pendingInputMu.Unlock()
	_ = c.Bot().Delete(c.Callback().Message)
	return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "canceled")})
}

func applySettingsInput(c tele.Context, setting, value string) {
	uid := c.Sender().ID
	if !isAdmin(uid) {
		_ = c.Send(tr(uid, "admin_only"))
		return
	}
	if settings.ReadOnly {
		_ = c.Send(tr(uid, "settings_readonly"))
		return
	}
	if settings.BTsets == nil {
		_ = c.Send(tr(uid, "settings_not_loaded"))
		return
	}

	clear := strings.ToLower(value) == "clear" || strings.ToLower(value) == "очистить" || value == "-"
	if clear {
		value = ""
	}

	sets := new(settings.BTSets)
	*sets = *settings.BTsets

	switch setting {
	case "friendlyname":
		sets.FriendlyName = value
		_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_done"), "FriendlyName", valueOrClear(value)))
	case "torrentssavepath":
		if value != "" {
			abs, err := filepath.Abs(value)
			if err != nil {
				abs = value
			}
			if _, err := os.Stat(abs); err != nil && !os.IsNotExist(err) {
				_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_error"), err.Error()))
				return
			}
			sets.TorrentsSavePath = abs
			sets.UseDisk = true
		} else {
			sets.TorrentsSavePath = ""
			sets.UseDisk = false
		}
		_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_done"), "TorrentsSavePath", valueOrClear(value)))
	case "sslcert":
		sets.SslCert = value
		_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_done"), "SslCert", valueOrClear(value)))
	case "sslkey":
		sets.SslKey = value
		_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_done"), "SslKey", valueOrClear(value)))
	case "tmdbkey":
		sets.TMDBSettings.APIKey = value
		_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_done"), "TMDB API Key", valueOrClear(value)))
	case "proxyhosts":
		if value == "" {
			sets.ProxyHosts = nil
		} else {
			parts := strings.Split(value, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			sets.ProxyHosts = parts
		}
		_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_done"), "ProxyHosts", valueOrClear(value)))
	case "torznab_add":
		if value == "" {
			_ = c.Send(tr(uid, "settings_input_torznab_usage"))
			return
		}
		parts := strings.SplitN(value, "|", 3)
		cfg := settings.TorznabConfig{Host: strings.TrimSpace(parts[0])}
		if len(parts) > 1 {
			cfg.Key = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			cfg.Name = strings.TrimSpace(parts[2])
		}
		if !strings.HasPrefix(cfg.Host, "http") {
			cfg.Host = "https://" + cfg.Host
		}
		sets.TorznabUrls = append(sets.TorznabUrls, cfg)
		_ = c.Send(fmt.Sprintf(tr(uid, "settings_input_torznab_added"), cfg.Host))
	case "torznab_test":
		if value == "" {
			_ = c.Send(tr(uid, "settings_input_torznab_usage"))
			return
		}
		parts := strings.SplitN(value, "|", 3)
		host := strings.TrimSpace(parts[0])
		key := ""
		if len(parts) > 1 {
			key = strings.TrimSpace(parts[1])
		}
		if !strings.HasPrefix(host, "http") {
			host = "https://" + host
		}
		if err := torznab.Test(host, key); err != nil {
			_ = c.Send(fmt.Sprintf(tr(uid, "settings_torznab_test_fail"), err.Error()))
			return
		}
		_ = c.Send(tr(uid, "settings_torznab_test_ok"))
		return
	default:
		_ = c.Send(tr(uid, "callback_unknown"))
		return
	}

	torr.SetSettings(sets)
	dlna.Stop()
	if sets.EnableDLNA {
		dlna.Start()
	}
	rutor.Stop()
	rutor.Start()
}

func valueOrClear(v string) string {
	if v == "" {
		return "(cleared)"
	}
	if len(v) > 50 {
		return v[:47] + "..."
	}
	return v
}
