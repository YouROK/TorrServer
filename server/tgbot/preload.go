package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdPreload(c tele.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Send(tr(c.Sender().ID, "preload_usage"))
	}
	uid := c.Sender().ID
	hash := resolveHash(c, args[0])
	if hash == "" {
		return c.Send(tr(uid, "invalid_hash"))
	}
	index, err := strconv.Atoi(strings.TrimSpace(args[1]))
	if err != nil || index < 1 {
		return c.Send(tr(uid, "preload_invalid"))
	}

	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Send(tr(uid, "torrent_not_found") + ":\n<code>" + hash + "</code>")
	}

	torr.Preload(t, index)
	return c.Send(fmt.Sprintf(tr(uid, "preload_started"), args[1]))
}

func callbackPreload(c tele.Context, hash, indexStr string) error {
	uid := c.Sender().ID
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 1 {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "error")})
	}
	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "torrent_not_found")})
	}
	torr.Preload(t, index)
	return c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf(tr(uid, "preload_btn"), indexStr)})
}
