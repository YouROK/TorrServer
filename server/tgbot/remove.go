package tgbot

import (
	"fmt"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdRemove(c tele.Context) error {
	arg := ""
	if args := c.Args(); len(args) > 0 {
		arg = args[0]
	}
	hash := resolveHash(c, arg)
	if hash == "" {
		return c.Send(tr(c.Sender().ID, "remove_usage"))
	}

	torrents := torr.ListTorrent()
	var found bool
	for _, t := range torrents {
		if t.Hash().HexString() == hash {
			found = true
			break
		}
	}
	if !found {
		return c.Send(tr(c.Sender().ID, "torrent_not_found") + ":\n<code>" + hash + "</code>")
	}

	torr.RemTorrent(hash)
	return c.Send(fmt.Sprintf(tr(c.Sender().ID, "remove_done"), hash))
}
