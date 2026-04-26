package tgbot

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdSet(c tele.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Send(tr(c.Sender().ID, "set_usage"))
	}
	hash := resolveHash(c, args[0])
	if hash == "" {
		return c.Send(tr(c.Sender().ID, "invalid_hash"))
	}
	title := strings.TrimSpace(strings.Join(args[1:], " "))
	if title == "" {
		return c.Send(tr(c.Sender().ID, "set_title_required"))
	}

	torr.SetTorrent(hash, title, "", "", "")
	return c.Send(fmt.Sprintf(tr(c.Sender().ID, "set_done"), escapeHtml(title)))
}
