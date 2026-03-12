package tgbot

import (
	"fmt"
	"regexp"
	"strings"

	tele "gopkg.in/telebot.v4"
)

var magnetRegex = regexp.MustCompile(`magnet:\?[^\s]+`)
var torrsRegex = regexp.MustCompile(`torrs://[^\s]+`)
var hashRegex = regexp.MustCompile(`\b([a-fA-F0-9]{40})\b`)

func cmdImport(c tele.Context) error {
	text := ""
	if c.Message() != nil && c.Message().Text != "" {
		text = strings.TrimPrefix(strings.TrimSpace(c.Message().Text), "/import")
		text = strings.TrimSpace(text)
	}
	if text == "" {
		return c.Send(tr(c.Sender().ID, "import_usage"))
	}
	var links []string
	seen := make(map[string]bool)
	for _, m := range magnetRegex.FindAllString(text, -1) {
		m = strings.TrimSpace(m)
		if m != "" && !seen[m] {
			seen[m] = true
			links = append(links, m)
		}
	}
	for _, m := range torrsRegex.FindAllString(text, -1) {
		m = strings.TrimSpace(m)
		if m != "" && !seen[m] {
			seen[m] = true
			links = append(links, m)
		}
	}
	for _, m := range hashRegex.FindAllString(text, -1) {
		h := strings.ToLower(strings.TrimSpace(m))
		if h != "" && !seen[h] {
			seen[h] = true
			links = append(links, h)
		}
	}
	if len(links) == 0 {
		return c.Send(tr(c.Sender().ID, "import_no_links"))
	}
	uid := c.Sender().ID
	added := 0
	for _, link := range links {
		if err := addTorrent(c, link); err != nil {
			_ = c.Send(fmt.Sprintf(tr(uid, "add_error"), err.Error()))
			continue
		}
		added++
	}
	return c.Send(fmt.Sprintf(tr(uid, "import_done"), added, len(links)))
}
