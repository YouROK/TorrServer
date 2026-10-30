package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"

	"server/library"
	"server/log"
	"server/torr"
)

// Keep page small so reply_markup stays under Telegram limits with long titles.
const listPageSize = 6

// Max runes for a torrent title line in the hub message body (full title on card).
const listTitleMaxRunes = 72

func listNavToken(page int, cat string) string {
	if cat == "" {
		return strconv.Itoa(page)
	}
	return strconv.Itoa(page) + "|" + cat
}

func parseListNav(s string) (page int, cat string) {
	parts := strings.SplitN(s, "|", 2)
	if len(parts) > 0 {
		page, _ = strconv.Atoi(parts[0])
	}
	if len(parts) == 2 {
		cat = parts[1]
	}
	return
}

func filterLibrary(cat string) []*torr.Torrent {
	all := torr.ListTorrent()
	if cat == "" || cat == "all" {
		return all
	}
	var out []*torr.Torrent
	for _, t := range all {
		if _, _, category, _ := t.Meta(); library.MatchCategory(category, cat) {
			out = append(out, t)
		}
	}
	return out
}

func list(c tele.Context) error {
	cat := ""
	if args := c.Args(); len(args) > 0 {
		cat = library.NormalizeCategory(args[0])
		if cat == "all" {
			cat = ""
		}
	}
	return sendListHub(c, 0, cat, false)
}

// sendListHub renders one message: short numbered rows + pick/nav/refresh buttons.
func sendListHub(c tele.Context, page int, cat string, edit bool) error {
	uid := c.Sender().ID
	torrents := filterLibrary(cat)
	if len(torrents) == 0 {
		msg := tr(uid, "no_torrents")
		if cat != "" {
			msg = fmt.Sprintf(tr(uid, "list_empty_cat"), categoryLabel(uid, cat))
		}
		if edit && c.Callback() != nil && c.Callback().Message != nil {
			_, err := c.Bot().Edit(c.Callback().Message, msg, listFilterKeyboard(uid, cat), tele.ModeHTML)
			if isMessageNotModified(err) {
				return nil
			}
			return err
		}
		if cat != "" {
			return c.Send(msg, listFilterKeyboard(uid, cat), tele.ModeHTML)
		}
		return sendWithMenu(c, msg)
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
	navTok := listNavToken(page, cat)

	var b strings.Builder
	b.WriteString("<b>" + tr(uid, "menu_library") + "</b>")
	if cat != "" {
		fmt.Fprintf(&b, " · %s", escapeHtml(categoryLabel(uid, cat)))
	}
	fmt.Fprintf(&b, " — %s %d/%d\n\n", tr(uid, "page"), page+1, totalPages)
	if peekPendingTool(uid) != "" {
		b.WriteString("<i>" + tr(uid, "menu_pick_torrent") + "</i>\n\n")
	}
	for i, t := range pageTorrents {
		n := start + i + 1
		title, _, category, _ := t.Meta()
		if title == "" {
			title = t.Hash().HexString()
		}
		size := ""
		if t.Size > 0 {
			size = " <i>" + humanize.IBytes(uint64(t.Size)) + "</i>"
		}
		catMark := ""
		if category != "" {
			catMark = " <i>[" + escapeHtml(category) + "]</i>"
		}
		fmt.Fprintf(&b, "<b>%d.</b> %s%s%s\n", n, escapeHtml(shortListTitle(title)), catMark, size)
	}

	m := &tele.ReplyMarkup{}
	var rows []tele.Row
	rows = append(rows, listFilterRow(m, uid, cat)...)
	for i, t := range pageTorrents {
		n := start + i + 1
		hash := t.Hash().HexString()
		title := torrentTitle(t)
		if title == "" {
			title = hash[:8] + "…"
		}
		label := numberedBtnLabel(n, title)
		rows = append(rows, m.Row(m.Data(label, "ftpick", hash, navTok)))
	}

	var nav []tele.Btn
	if totalPages > 1 {
		if page > 0 {
			nav = append(nav, m.Data("◀️", "flist", listNavToken(page-1, cat)))
		}
		nav = append(nav, m.Data(strconv.Itoa(page+1)+"/"+strconv.Itoa(totalPages), "fnop"))
		if page < totalPages-1 {
			nav = append(nav, m.Data("▶️", "flist", listNavToken(page+1, cat)))
		}
	}
	nav = append(nav, m.Data("🔄", "frefresh", navTok))
	rows = append(rows, m.Row(nav...))
	if botUsername != "" {
		rows = append(rows, m.Row(m.QueryChat(tr(uid, "menu_search_inline"), inlineSwitchQuery)))
	}
	m.Inline(rows...)

	txt := b.String()
	if edit && c.Callback() != nil && c.Callback().Message != nil {
		_, err := c.Bot().Edit(c.Callback().Message, txt, m, tele.ModeHTML)
		if isMessageNotModified(err) {
			return nil
		}
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

func listFilterRow(m *tele.ReplyMarkup, uid int64, current string) []tele.Row {
	mark := func(key, label string) string {
		if current == key || (key == "all" && current == "") {
			return "• " + label
		}
		return label
	}
	return []tele.Row{m.Row(
		m.Data(mark("all", tr(uid, "cat_all")), "flistcat", "all"),
		m.Data(mark("movie", tr(uid, "cat_movie")), "flistcat", "movie"),
		m.Data(mark("tv", tr(uid, "cat_tv")), "flistcat", "tv"),
		m.Data(mark("music", tr(uid, "cat_music")), "flistcat", "music"),
		m.Data(mark("other", tr(uid, "cat_other")), "flistcat", "other"),
	)}
}

func listFilterKeyboard(uid int64, current string) *tele.ReplyMarkup {
	m := &tele.ReplyMarkup{}
	m.Inline(listFilterRow(m, uid, current)...)
	return m
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

func showTorrentCard(c tele.Context, hash string, nav string, edit bool) error {
	uid := c.Sender().ID
	t := torr.GetTorrent(hash)
	if t == nil {
		if c.Callback() != nil {
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
		}
		return c.Send(tr(uid, "torrent_not_found"))
	}
	metaTitle, poster, category, _ := t.Meta()
	title := escapeHtml(metaTitle)
	if title == "" {
		title = hash
	}
	msg := "<b>" + title + "</b>"
	if t.Size > 0 {
		msg += " <i>" + humanize.IBytes(uint64(t.Size)) + "</i>"
	}
	msg += "\n" + tr(uid, "card_category") + ": <i>" + escapeHtml(categoryLabel(uid, category)) + "</i>"
	msg += "\n<code>" + hash + "</code>"

	m := &tele.ReplyMarkup{}
	rows := []tele.Row{
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
	}
	rows = append(rows, categoryButtonRows(m, uid, hash)...)
	playURL := library.ShortPlayURL(getHost(), hash, 1)
	rows = append(rows, appendCopyRow(m, uid, hash, playURL, magnetForHash(hash))...)
	if nav == "" {
		nav = "0"
	}
	rows = append(rows, m.Row(m.Data(tr(uid, "btn_back_list"), "fbacklist", nav)))
	m.Inline(rows...)

	if edit && c.Callback() != nil && c.Callback().Message != nil {
		_ = c.Respond(&tele.CallbackResponse{})
		cbMsg := c.Callback().Message
		if cbMsg.Photo != nil {
			_, err := c.Bot().EditCaption(cbMsg, msg, m, tele.ModeHTML)
			return err
		}
		_, err := c.Bot().Edit(cbMsg, msg, m, tele.ModeHTML)
		return err
	}
	if !edit && isPosterURL(poster) {
		photo := &tele.Photo{File: tele.FromURL(poster), Caption: msg}
		if err := c.Send(photo, m, tele.ModeHTML); err != nil {
			log.TLogln("tg poster send err", err)
		} else {
			return nil
		}
	}
	return c.Send(msg, m, tele.ModeHTML)
}

func callbackListPage(c tele.Context, data string) error {
	page, cat := parseListNav(data)
	_ = c.Respond(&tele.CallbackResponse{})
	return sendListHub(c, page, cat, true)
}

func callbackListRefresh(c tele.Context, data string) error {
	page, cat := parseListNav(data)
	_ = c.Respond(&tele.CallbackResponse{Text: "🔄"})
	return sendListHub(c, page, cat, true)
}

func callbackListCat(c tele.Context, cat string) error {
	norm := library.NormalizeCategory(cat)
	if norm == "all" {
		norm = ""
	}
	_ = c.Respond(&tele.CallbackResponse{})
	return sendListHub(c, 0, norm, true)
}

func callbackTorrentPick(c tele.Context, hash, pageStr string) error {
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	if kind := takePendingTool(c.Sender().ID); kind != "" {
		_ = c.Respond(&tele.CallbackResponse{})
		return runPendingTool(c, kind, hash)
	}
	return showTorrentCard(c, hash, pageStr, true)
}

func callbackBackList(c tele.Context, pageStr string) error {
	page, cat := parseListNav(pageStr)
	_ = c.Respond(&tele.CallbackResponse{})
	return sendListHub(c, page, cat, true)
}
