package tgbot

import (
	"fmt"
	"sort"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdCategories(c tele.Context) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}
	uid := c.Sender().ID
	catCount := make(map[string]int)
	for _, t := range torrents {
		cat := t.Category
		if cat == "" {
			cat = tr(uid, "categories_uncategorized")
		}
		catCount[cat]++
	}
	var cats []string
	for c := range catCount {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	var sb strings.Builder
	fmt.Fprintf(&sb, "📁 <b>%s</b>\n\n", tr(uid, "categories_title"))
	for _, cat := range cats {
		fmt.Fprintf(&sb, "• %s: %d\n", escapeHtml(cat), catCount[cat])
	}
	return c.Send(strings.TrimSuffix(sb.String(), "\n"))
}
