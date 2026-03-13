package tgbot

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/log"
	"server/torr"
)

// humanizeSpeedBits formats bytes/s as bits/s (bps, kbps, Mbps, Gbps, Tbps) — same as web mode.
func humanizeSpeedBits(uid int64, bytesPerSec float64) string {
	if bytesPerSec <= 0 {
		return "0 " + tr(uid, "speed_bps")
	}
	bits := bytesPerSec * 8
	i := int(math.Floor(math.Log(bits) / math.Log(1000)))
	if i < 0 {
		i = 0
	}
	units := []string{"speed_bps", "speed_kbps", "speed_Mbps", "speed_Gbps", "speed_Tbps"}
	if i >= len(units) {
		i = len(units) - 1
	}
	val := bits / math.Pow(1000, float64(i))
	return fmt.Sprintf("%.0f %s", val, tr(uid, units[i]))
}

var (
	statusStopChans   = make(map[int]chan struct{})
	statusStopChansMu sync.Mutex
)

func cmdStatus(c tele.Context) error {
	arg := ""
	if args := c.Args(); len(args) > 0 {
		arg = args[0]
	}
	hash := resolveHash(c, arg)

	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}

	if hash != "" {
		t := torr.GetTorrent(hash)
		if t == nil {
			return c.Send(tr(c.Sender().ID, "torrent_not_found") + ":\n<code>" + hash + "</code>")
		}
		log.TLogln("tg status cmd", logUser(c.Sender()), logSafeStr(t.Title, 40), hash)
		if !t.WaitInfo() {
			msg, err := c.Bot().Send(c.Sender(), tr(c.Sender().ID, "status_waiting"))
			if err != nil {
				return err
			}
			go waitForInfoAndUpdateStatus(c.Bot(), msg, hash, c.Sender().ID)
			return nil
		}
		return sendStatus(c, t)
	}

	return sendStatusAllPage(c, 0)
}

const statusAllPageSize = 5

func sendStatusAllPage(c tele.Context, page int) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}

	totalPages := (len(torrents) + statusAllPageSize - 1) / statusAllPageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * statusAllPageSize
	end := start + statusAllPageSize
	if end > len(torrents) {
		end = len(torrents)
	}
	pageTorrents := torrents[start:end]

	uid := c.Sender().ID
	var sb strings.Builder
	for _, t := range pageTorrents {
		txt := formatTorrentStatus(uid, t)
		if txt != "" {
			sb.WriteString(txt)
			sb.WriteString("\n\n")
		}
	}
	if sb.Len() == 0 {
		return c.Send(tr(uid, "status_no_active"))
	}
	msg := strings.TrimSuffix(sb.String(), "\n\n")

	navRow := []tele.InlineButton{}
	if totalPages > 1 {
		if page > 0 {
			navRow = append(navRow, tele.InlineButton{Text: "◀️", Unique: "fstatusall", Data: strconv.Itoa(page - 1)})
		}
		navRow = append(navRow, tele.InlineButton{Text: strconv.Itoa(page+1) + "/" + strconv.Itoa(totalPages), Unique: "fnop", Data: ""})
		if page < totalPages-1 {
			navRow = append(navRow, tele.InlineButton{Text: "▶️", Unique: "fstatusall", Data: strconv.Itoa(page + 1)})
		}
	}
	navRow = append(navRow, tele.InlineButton{Text: "🔄", Unique: "fstatusallrefresh", Data: strconv.Itoa(page)})

	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{navRow}}
	if err := c.Send(msg, kbd); err != nil {
		log.TLogln("tg status all send err", err)
		return err
	}
	return nil
}

func callbackStatusAllPage(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(data); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendStatusAllPage(c, page)
}

func callbackStatusAllRefresh(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(data); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{Text: "🔄"})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendStatusAllPage(c, page)
}

func sendStatus(c tele.Context, t *torr.Torrent) error {
	uid := c.Sender().ID
	txt := formatTorrentStatus(uid, t)
	if txt == "" && t != nil {
		txt = "<b>" + escapeHtml(t.Title) + "</b>\n" + tr(uid, "status_label") + ": " + t.Stat.String()
	}
	hash := ""
	if t != nil {
		hash = t.Hash().HexString()
	}
	kbd := statusKeyboard(uid, hash, true)
	msg, err := c.Bot().Send(c.Sender(), txt, kbd)
	if err != nil {
		return err
	}
	if t != nil {
		log.TLogln("tg status sent", logUserID(uid), logSafeStr(t.Title, 40), hash)
		go refreshStatusLoop(c.Bot(), msg, hash, uid)
	}
	return nil
}

func statusKeyboard(uid int64, hash string, active bool) *tele.ReplyMarkup {
	if active {
		return &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
			{
				{Text: "🔄", Unique: "fstatusrefresh", Data: hash},
				{Text: tr(uid, "status_stop_btn"), Unique: "fstatusstop", Data: hash},
			},
		}}
	}
	return &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
		{{Text: tr(uid, "status_refresh_btn"), Unique: "fstatusrefresh", Data: hash}},
	}}
}

func refreshStatusLoop(api tele.API, msg *tele.Message, hash string, uid int64) {
	const interval = 5 * time.Second
	const duration = 2 * time.Minute
	stopCh := make(chan struct{})
	statusStopChansMu.Lock()
	statusStopChans[msg.ID] = stopCh
	statusStopChansMu.Unlock()
	defer func() {
		statusStopChansMu.Lock()
		delete(statusStopChans, msg.ID)
		statusStopChansMu.Unlock()
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	deadline := time.Now().Add(duration)
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if time.Now().After(deadline) {
				t := torr.GetTorrent(hash)
				txt := ""
				if t != nil {
					txt = formatTorrentStatus(uid, t)
					if txt == "" {
						txt = "<b>" + escapeHtml(t.Title) + "</b>\n" + tr(uid, "status_label") + ": " + t.Stat.String()
					}
					txt += "\n\n" + tr(uid, "status_auto_ended")
				} else {
					txt = "<code>" + hash + "</code>\n\n" + tr(uid, "status_torrent_gone")
				}
				_, _ = api.Edit(msg, txt, statusKeyboard(uid, hash, false), tele.ModeHTML)
				return
			}
			t := torr.GetTorrent(hash)
			if t == nil {
				txt := "<code>" + hash + "</code>\n\n" + tr(uid, "status_torrent_gone")
				_, _ = api.Edit(msg, txt, statusKeyboard(uid, hash, false), tele.ModeHTML)
				return
			}
			txt := formatTorrentStatus(uid, t)
			if txt == "" {
				txt = "<b>" + escapeHtml(t.Title) + "</b>\n" + tr(uid, "status_label") + ": " + t.Stat.String()
			}
			if _, err := api.Edit(msg, txt, statusKeyboard(uid, hash, true), tele.ModeHTML); err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "message is not modified") {
					continue
				}
				if strings.Contains(errStr, "message to edit not found") {
					return
				}
				log.TLogln("tg status refresh err", err)
				return
			}
		}
	}
}

func stopStatusRefresh(msgID int) {
	statusStopChansMu.Lock()
	ch := statusStopChans[msgID]
	delete(statusStopChans, msgID)
	statusStopChansMu.Unlock()
	if ch != nil {
		close(ch)
	}
}

const waitForInfoTimeout = 2 * time.Minute

func waitForInfoAndUpdateStatus(api tele.API, msg *tele.Message, hash string, uid int64) {
	deadline := time.Now().Add(waitForInfoTimeout)
	for {
		t := torr.GetTorrent(hash)
		if t == nil {
			_, _ = api.Edit(msg, tr(uid, "torrent_not_found")+":\n<code>"+hash+"</code>", tele.ModeHTML)
			return
		}
		if t.WaitInfo() {
			break
		}
		if time.Now().After(deadline) {
			_, _ = api.Edit(msg, tr(uid, "status_waiting")+"\n\n"+tr(uid, "status_auto_ended"), tele.ModeHTML)
			return
		}
		time.Sleep(time.Second)
	}
	t := torr.GetTorrent(hash)
	if t == nil {
		_, _ = api.Edit(msg, tr(uid, "torrent_not_found")+":\n<code>"+hash+"</code>", tele.ModeHTML)
		return
	}
	txt := formatTorrentStatus(uid, t)
	if txt == "" {
		txt = "<b>" + escapeHtml(t.Title) + "</b>\n" + tr(uid, "status_label") + ": " + t.Stat.String()
	}
	if _, err := api.Edit(msg, txt, statusKeyboard(uid, hash, true), tele.ModeHTML); err != nil {
		log.TLogln("tg status wait edit err", err)
		return
	}
	go refreshStatusLoop(api, msg, hash, uid)
}

func callbackStatusRefresh(c tele.Context, hash string) error {
	uid := c.Sender().ID
	if hash == "" {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "callback_unknown")})
	}
	t := torr.GetTorrent(hash)
	if t != nil {
		log.TLogln("tg status refresh", logUserID(uid), logSafeStr(t.Title, 40), hash)
	}
	if t == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
	}
	if c.Callback().Message != nil {
		stopStatusRefresh(c.Callback().Message.ID)
		_ = c.Bot().Delete(c.Callback().Message)
	}
	_ = c.Respond(&tele.CallbackResponse{})
	return sendStatus(c, t)
}

func callbackStatusStop(c tele.Context, hash string) error {
	uid := c.Sender().ID
	if hash != "" {
		if t := torr.GetTorrent(hash); t != nil {
			log.TLogln("tg status stop", logUserID(uid), logSafeStr(t.Title, 40), hash)
		}
	}
	if c.Callback().Message != nil {
		stopStatusRefresh(c.Callback().Message.ID)
		if hash != "" {
			msg := c.Callback().Message
			t := torr.GetTorrent(hash)
			txt := ""
			if t != nil {
				txt = formatTorrentStatus(uid, t)
				if txt == "" {
					txt = "<b>" + escapeHtml(t.Title) + "</b>\n" + tr(uid, "status_label") + ": " + t.Stat.String()
				}
			} else {
				txt = "<code>" + hash + "</code>"
			}
			txt += "\n\n" + tr(uid, "status_stopped")
			_, _ = c.Bot().Edit(msg, txt, statusKeyboard(uid, hash, false), tele.ModeHTML)
		}
	}
	return c.Respond(&tele.CallbackResponse{Text: "🛑"})
}

func callbackStatus(c tele.Context, hash string) error {
	uid := c.Sender().ID
	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
	}
	_ = c.Respond(&tele.CallbackResponse{})
	if !t.WaitInfo() {
		msg, err := c.Bot().Send(c.Sender(), tr(uid, "status_waiting"))
		if err != nil {
			return err
		}
		go waitForInfoAndUpdateStatus(c.Bot(), msg, hash, uid)
		return nil
	}
	return sendStatus(c, t)
}

func formatTorrentStatus(uid int64, t *torr.Torrent) string {
	if t == nil {
		return ""
	}
	st := t.Status()
	if st == nil {
		return "<b>" + escapeHtml(t.Title) + "</b>\n" + tr(uid, "status_label") + ": " + t.Stat.String()
	}

	// For streaming: size + cache info (progress is misleading — we stream, not download sequentially)
	sizeLine := fmt.Sprintf("%s: %s", tr(uid, "status_size"), humanize.IBytes(uint64(st.TorrentSize)))
	if cache := t.CacheState(); cache != nil {
		sizeLine += fmt.Sprintf(" | %s: %s / %s · %d %s",
			tr(uid, "status_cache"),
			humanize.IBytes(uint64(cache.Filled)),
			humanize.IBytes(uint64(cache.Capacity)),
			len(cache.Readers),
			tr(uid, "status_streams"))
	}

	txt := fmt.Sprintf("<b>%s</b>\n", escapeHtml(st.Title))
	txt += fmt.Sprintf("%s: %s\n", tr(uid, "status_label"), st.StatString)
	txt += sizeLine + "\n"
	txt += fmt.Sprintf("%s: %s | %s: %s\n",
		tr(uid, "status_download"), humanizeSpeedBits(uid, st.DownloadSpeed),
		tr(uid, "status_upload"), humanizeSpeedBits(uid, st.UploadSpeed))
	txt += fmt.Sprintf("%s: %d %s, %d %s\n",
		tr(uid, "stats_peers"), st.ActivePeers, tr(uid, "stats_active"),
		st.ConnectedSeeders, tr(uid, "stats_seeds"))
	txt += fmt.Sprintf("<code>%s</code>", st.Hash)
	return txt
}
