package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"

	"server/log"
	"server/torr"
)

// Keep page small so reply_markup stays under Telegram limits with long titles.
const listPageSize = 6

// Max runes for a torrent title line in the hub message body (full title on card).
const listTitleMaxRunes = 72

func list(c tele.Context) error {
	return sendListHub(c, 0, false)
}

// sendListHub renders one message: short numbered rows + pick/nav/refresh buttons.
// When edit is true, updates the callback message in place.
// Do not pass mainMenuKeyboard here — telebot keeps only the last ReplyMarkup.
func sendListHub(c tele.Context, page int, edit bool) error {
	uid := c.Sender().ID
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		if edit && c.Callback() != nil && c.Callback().Message != nil {
			_, err := c.Bot().Edit(c.Callback().Message, tr(uid, "no_torrents"), &tele.ReplyMarkup{})
			return err
		}
		return sendWithMenu(c, tr(uid, "no_torrents"))
	}

	totalPages := (len(torrents) + listPageSize - 1) / listPageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * listPageSize
	end := start + listPageSize
	if end > len(torrents) {
		end = len(torrents)
	}
	pageTorrents := torrents[start:end]

	var b strings.Builder
	// menu_library already includes 📚 — do not prefix another.
	b.WriteString("<b>" + tr(uid, "menu_library") + "</b>")
	fmt.Fprintf(&b, " — %s %d/%d\n\n", tr(uid, "page"), page+1, totalPages)
	for i, t := range pageTorrents {
		n := start + i + 1
		title := t.Title
		if title == "" {
			title = t.Hash().HexString()
		}
		size := ""
		if t.Size > 0 {
			size = " <i>" + humanize.IBytes(uint64(t.Size)) + "</i>"
		}
		fmt.Fprintf(&b, "<b>%d.</b> %s%s\n", n, escapeHtml(shortListTitle(title)), size)
	}

	m := &tele.ReplyMarkup{}
	var rows []tele.Row
	for i, t := range pageTorrents {
		n := start + i + 1
		hash := t.Hash().HexString()
		title := t.Title
		if title == "" {
			title = hash[:8] + "…"
		}
		label := numberedBtnLabel(n, title)
		rows = append(rows, m.Row(m.Data(label, "ftpick", hash, strconv.Itoa(page))))
	}

	var nav []tele.Btn
	if totalPages > 1 {
		if page > 0 {
			nav = append(nav, m.Data("◀️", "flist", strconv.Itoa(page-1)))
		}
		nav = append(nav, m.Data(strconv.Itoa(page+1)+"/"+strconv.Itoa(totalPages), "fnop"))
		if page < totalPages-1 {
			nav = append(nav, m.Data("▶️", "flist", strconv.Itoa(page+1)))
		}
	}
	nav = append(nav, m.Data("🔄", "frefresh", strconv.Itoa(page)))
	rows = append(rows, m.Row(nav...))
	m.Inline(rows...)

	txt := b.String()
	if edit && c.Callback() != nil && c.Callback().Message != nil {
		_, err := c.Bot().Edit(c.Callback().Message, txt, m, tele.ModeHTML)
		if err != nil {
			log.TLogln("tg list hub edit err", err)
		}
		return err
	}
	if err := c.Send(txt, m, tele.ModeHTML); err != nil {
		log.TLogln("tg list hub send err", err)
		return err
	}
	return nil
}

func shortListTitle(s string) string {
	r := []rune(s)
	if len(r) <= listTitleMaxRunes {
		return s
	}
	return string(r[:listTitleMaxRunes-1]) + "…"
}

func numberedBtnLabel(n int, title string) string {
	prefix := strconv.Itoa(n) + ". "
	if title == "" {
		return truncateBtnText(prefix)
	}
	return truncateBtnText(prefix + title)
}

func showTorrentCard(c tele.Context, hash string, listPage int, edit bool) error {
	uid := c.Sender().ID
	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
	}
	title := escapeHtml(t.Title)
	if title == "" {
		title = hash
	}
	msg := "<b>" + title + "</b>"
	if t.Size > 0 {
		msg += " <i>" + humanize.IBytes(uint64(t.Size)) + "</i>"
	}
	msg += "\n<code>" + hash + "</code>"

	pageStr := strconv.Itoa(listPage)
	m := &tele.ReplyMarkup{}
	m.Inline(
		m.Row(
			m.Data(tr(uid, "btn_files"), "files", hash),
			m.Data(tr(uid, "btn_status"), "fstatus", hash),
			m.Data(tr(uid, "btn_m3u"), "fm3u", hash),
		),
		m.Row(
			m.Data(tr(uid, "btn_link"), "flink", hash),
			m.Data(tr(uid, "btn_drop"), "fdrop", hash),
			m.Data(tr(uid, "btn_delete"), "delete", hash),
		),
		m.Row(m.Data(tr(uid, "btn_back_list"), "fbacklist", pageStr)),
	)

	if edit && c.Callback() != nil && c.Callback().Message != nil {
		_ = c.Respond(&tele.CallbackResponse{})
		_, err := c.Bot().Edit(c.Callback().Message, msg, m, tele.ModeHTML)
		return err
	}
	return c.Send(msg, m, tele.ModeHTML)
}

func callbackListPage(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(strings.Split(data, "|")[0]); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{})
	return sendListHub(c, page, true)
}

func callbackListRefresh(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(strings.Split(data, "|")[0]); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{Text: "🔄"})
	return sendListHub(c, page, true)
}

func callbackTorrentPick(c tele.Context, hash, pageStr string) error {
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	page := 0
	if p, err := strconv.Atoi(pageStr); err == nil {
		page = p
	}
	return showTorrentCard(c, hash, page, true)
}

func callbackBackList(c tele.Context, pageStr string) error {
	page := 0
	if p, err := strconv.Atoi(pageStr); err == nil {
		page = p
	}
	_ = c.Respond(&tele.CallbackResponse{})
	return sendListHub(c, page, true)
}
