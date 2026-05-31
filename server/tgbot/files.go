package tgbot

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"

	"server/log"
	sets "server/settings"
	"server/torr"
)

// Telegram limits the serialized reply_markup size; many file rows with long
// labels/URLs would exceed it (e.g. "reply markup is too long").
const filesPageSize = 5

// Inline button text is limited to 64 characters in the Bot API.
func truncateBtnText(s string) string {
	const max = 64
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

func files(c tele.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	hash := args[1]
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	msg, err := c.Bot().Send(c.Sender(), tr(c.Sender().ID, "connecting"))
	t := torr.GetTorrent(hash)
	if t == nil {
		if err == nil {
			_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "torrent_not_found")+":\n<code>"+hash+"</code>")
		}
		return nil
	}
	if err == nil {
		api := c.Bot()
		recipient := c.Sender()
		uid := c.Sender().ID
		go sendFilesList(api, recipient, msg, hash, uid, 0)
	}
	return err
}

// sendFilesList shows one page of per-file actions; fitems / fifresh change the page in-place.
func sendFilesList(api tele.API, recipient tele.Recipient, statusMsg *tele.Message, hash string, uid int64, page int) {
	t := torr.GetTorrent(hash)
	for t != nil && !t.WaitInfo() {
		time.Sleep(time.Second)
		t = torr.GetTorrent(hash)
	}
	_ = api.Delete(statusMsg)
	t = torr.GetTorrent(hash)
	if t == nil {
		return
	}
	ti := t.Status()
	if ti == nil {
		return
	}

	host := getHost()
	txt, kbd := buildFilesListView(t, host, uid, page)
	if kbd == nil {
		return
	}
	if _, err := api.Send(recipient, txt, kbd, tele.ModeHTML); err != nil {
		log.TLogln("tg files send err", err)
	}
}

func buildFilesListView(t *torr.Torrent, host string, uid int64, page int) (string, *tele.ReplyMarkup) {
	ti := t.Status()
	if ti == nil {
		return "", nil
	}
	hex := t.Hash().HexString()
	n := len(ti.FileStats)
	if n == 0 {
		return "", nil
	}

	totalPages := (n + filesPageSize - 1) / filesPageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * filesPageSize
	end := start + filesPageSize
	if end > n {
		end = n
	}
	pageFiles := ti.FileStats[start:end]

	viewedSet := make(map[int]struct{})
	for _, v := range sets.ListViewed(ti.Hash) {
		viewedSet[v.FileIndex] = struct{}{}
	}

	txt := "📁 <b>" + escapeHtml(ti.Title) + "</b> " +
		"<i>" + humanize.IBytes(uint64(ti.TorrentSize)) + "</i>\n\n" +
		"<code>" + ti.Hash + "</code>"
	if totalPages > 1 {
		txt += "\n\n" + tr(uid, "page") + " " + strconv.Itoa(page+1) + "/" + strconv.Itoa(totalPages)
	}
	if n > 1 {
		txt += "\n\n" + fmt.Sprintf(tr(uid, "files_range_hint"), n)
	}

	m := &tele.ReplyMarkup{}
	var rows []tele.Row

	for _, f := range pageFiles {
		viewedMark := ""
		if _, ok := viewedSet[f.Id]; ok {
			viewedMark = "✓ "
		}
		baseName := filepath.Base(f.Path)
		mline := viewedMark + "#" + strconv.Itoa(f.Id) + ": " + humanize.IBytes(uint64(f.Length)) + " — " + baseName
		fileLabel := truncateBtnText(mline)
		idStr := strconv.Itoa(f.Id)
		streamURL := host + "/stream/" + filepath.Base(f.Path) + "?link=" + hex + "&index=" + idStr + "&play"
		rows = append(rows, m.Row(
			m.Data(fileLabel, "upload", ti.Hash, idStr),
			m.URL(tr(uid, "files_link"), streamURL),
			m.Data("⏳", "fpreload", ti.Hash, idStr),
		))
	}

	if totalPages > 1 {
		var nav []tele.Btn
		if page > 0 {
			nav = append(nav, m.Data("◀️", "fitems", strconv.Itoa(page-1), ti.Hash))
		}
		nav = append(nav, m.Data(strconv.Itoa(page+1)+"/"+strconv.Itoa(totalPages), "fnop"))
		if page < totalPages-1 {
			nav = append(nav, m.Data("▶️", "fitems", strconv.Itoa(page+1), ti.Hash))
		}
		nav = append(nav, m.Data("🔄", "fifresh", strconv.Itoa(page), ti.Hash))
		rows = append(rows, m.Row(nav...))
	} else {
		rows = append(rows, m.Row(m.Data("🔄", "fifresh", strconv.Itoa(page), ti.Hash)))
	}
	if n > 1 {
		rows = append(rows, m.Row(m.Data(tr(uid, "files_download_all"), "fall", "all", ti.Hash)))
	}
	m.Inline(rows...)
	return txt, m
}

func callbackFileListPage(c tele.Context, pageStr, hash string) error {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	_ = c.Respond(&tele.CallbackResponse{})

	return editFilesListMessage(c, hash, c.Sender().ID, page)
}

func callbackFileListRefresh(c tele.Context, pageStr, hash string) error {
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	_ = c.Respond(&tele.CallbackResponse{Text: "🔄"})
	return editFilesListMessage(c, hash, c.Sender().ID, page)
}

func editFilesListMessage(c tele.Context, hash string, uid int64, page int) error {
	t := torr.GetTorrent(hash)
	if t == nil {
		_ = c.Send(tr(uid, "torrent_not_found") + ":\n<code>" + hash + "</code>")
		return nil
	}
	for t != nil && !t.WaitInfo() {
		time.Sleep(time.Second)
		t = torr.GetTorrent(hash)
	}
	t = torr.GetTorrent(hash)
	if t == nil {
		_ = c.Send(tr(uid, "torrent_not_found") + ":\n<code>" + hash + "</code>")
		return nil
	}
	host := getHost()
	txt, kbd := buildFilesListView(t, host, uid, page)
	if kbd == nil {
		log.TLogln("tg files: empty kbd for hash", logSafeStr(hash, 20))
		return nil
	}
	if c.Callback() == nil || c.Callback().Message == nil {
		_, err := c.Bot().Send(c.Sender(), txt, kbd, tele.ModeHTML)
		return err
	}
	_, err := c.Bot().Edit(c.Callback().Message, txt, kbd, tele.ModeHTML)
	if err != nil {
		if strings.Contains(err.Error(), "message is not modified") {
			return nil
		}
		log.TLogln("tg files edit err", err)
		_, _ = c.Bot().Send(c.Sender(), tr(uid, "error")+":\n"+escapeHtml(err.Error()), tele.ModeHTML)
		return err
	}
	return nil
}
