package tgbot

import (
	"strings"

	tele "gopkg.in/telebot.v4"
)

func cmdAdd(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) == 0 {
		return c.Send(tr(uid, "add_usage"))
	}
	link := strings.TrimSpace(strings.Join(args, " "))
	if link == "" {
		return c.Send(tr(uid, "add_no_link"))
	}
	err := addTorrent(c, link)
	if err != nil {
		return err
	}
	return list(c)
}
