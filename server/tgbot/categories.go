package tgbot

import (
	"fmt"
	"sort"
	"strings"

	"server/torr"

	tele "gopkg.in/telebot.v4"
)

func cmdCategories(c tele.Context) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}
	uid := c.Sender().ID
	catCount := make(map[string]int)
	for _, t := range torrents {
		_, _, cat, _ := t.Meta()
		if cat == "" {
			cat = "uncategorized"
		}
		catCount[cat]++
	}
	var cats []string
	for k := range catCount {
		cats = append(cats, k)
	}
	sort.Strings(cats)
	var sb strings.Builder
	fmt.Fprintf(&sb, "📁 <b>%s</b>\n\n", tr(uid, "categories_title"))
	for _, cat := range cats {
		fmt.Fprintf(&sb, "• %s: %d\n", escapeHtml(categoryLabel(uid, cat)), catCount[cat])
	}
	sb.WriteString("\n")
	sb.WriteString(tr(uid, "categories_filter_hint"))

	m := &tele.ReplyMarkup{}
	m.Inline(listFilterRow(m, uid, "")...)
	return c.Send(strings.TrimSuffix(sb.String(), "\n"), m, tele.ModeHTML)
}
