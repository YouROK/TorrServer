package tgbot

import (
	"fmt"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func callbackDrop(c tele.Context, hash string) error {
	torr.DropTorrent(hash)
	_ = c.Bot().Delete(c.Callback().Message)
	return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "drop_done")})
}

func cmdDrop(c tele.Context) error {
	arg := ""
	if args := c.Args(); len(args) > 0 {
		arg = args[0]
	}
	hash := resolveHash(c, arg)
	if hash == "" {
		return c.Send(tr(c.Sender().ID, "remove_usage"))
	}

	torr.DropTorrent(hash)
	return c.Send(fmt.Sprintf(tr(c.Sender().ID, "drop_done_hash"), hash))
}
