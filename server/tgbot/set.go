package tgbot

import (
	"fmt"
	"strings"

	"server/library"
	"server/torr"

	tele "gopkg.in/telebot.v4"
)

func cmdSet(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) < 2 {
		return c.Send(tr(uid, "set_usage"))
	}
	hash := resolveHash(c, args[0])
	if hash == "" {
		return c.Send(tr(uid, "invalid_hash"))
	}
	title := strings.TrimSpace(strings.Join(args[1:], " "))
	if title == "" {
		return c.Send(tr(uid, "set_title_required"))
	}
	if applyTorrentMeta(hash, &title, nil, nil) == nil {
		return c.Send(tr(uid, "torrent_not_found"))
	}
	return c.Send(fmt.Sprintf(tr(uid, "set_done"), escapeHtml(title)))
}

func cmdSetCat(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) < 1 {
		return c.Send(tr(uid, "setcat_usage"))
	}
	hash := resolveHash(c, args[0])
	if hash == "" {
		return c.Send(tr(uid, "invalid_hash"))
	}
	raw := ""
	if len(args) >= 2 {
		raw = strings.Join(args[1:], " ")
	}
	norm := library.NormalizeCategory(raw)
	if raw != "" && norm == "" {
		return c.Send(tr(uid, "setcat_usage"))
	}
	val := library.CategoryValue(norm)
	if applyTorrentMeta(hash, nil, nil, &val) == nil {
		return c.Send(tr(uid, "torrent_not_found"))
	}
	label := val
	if label == "" {
		label = tr(uid, "categories_uncategorized")
	}
	return c.Send(fmt.Sprintf(tr(uid, "setcat_done"), escapeHtml(label)))
}

func callbackSetCat(c tele.Context, hash, cat string) error {
	uid := c.Sender().ID
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "callback_unknown")})
	}
	norm := library.NormalizeCategory(cat)
	val := library.CategoryValue(norm)
	if applyTorrentMeta(hash, nil, nil, &val) == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
	}
	_ = c.Respond(&tele.CallbackResponse{Text: tr(uid, "setcat_saved")})
	return showTorrentCard(c, hash, "0", true)
}

func categoryButtonRows(m *tele.ReplyMarkup, uid int64, hash string) []tele.Row {
	return []tele.Row{
		m.Row(
			m.Data(tr(uid, "cat_movie"), "fsetcat", hash, "movie"),
			m.Data(tr(uid, "cat_tv"), "fsetcat", hash, "tv"),
			m.Data(tr(uid, "cat_music"), "fsetcat", hash, "music"),
			m.Data(tr(uid, "cat_other"), "fsetcat", hash, "other"),
		),
		m.Row(m.Data(tr(uid, "cat_clear"), "fsetcat", hash, "-")),
	}
}

func sendCategoryPicker(c tele.Context, hash string) error {
	uid := c.Sender().ID
	t := torr.GetTorrent(hash)
	title := hash
	if t != nil && torrentTitle(t) != "" {
		title = torrentTitle(t)
	}
	m := &tele.ReplyMarkup{}
	rows := categoryButtonRows(m, uid, hash)
	m.Inline(rows...)
	return c.Send(fmt.Sprintf(tr(uid, "cat_pick"), escapeHtml(title)), m, tele.ModeHTML)
}

func categoryLabel(uid int64, cat string) string {
	switch library.NormalizeCategory(cat) {
	case "movie":
		return tr(uid, "cat_movie")
	case "tv":
		return tr(uid, "cat_tv")
	case "music":
		return tr(uid, "cat_music")
	case "other":
		return tr(uid, "cat_other")
	default:
		if strings.TrimSpace(cat) == "" {
			return tr(uid, "categories_uncategorized")
		}
		return cat
	}
}
