package tgbot

import (
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	up "server/tgbot/upload"
)

func upload(c tele.Context) error {
	args := c.Args()
	if len(args) < 3 {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	hash := args[1]
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	id, err := strconv.Atoi(args[2])
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	up.AddRange(c, hash, id, id)
	return nil
}

func uploadall(c tele.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	hash := ""
	if len(args) >= 3 && isHash(args[2]) {
		hash = args[2]
	} else {
		hash = strings.TrimPrefix(args[1], "all|")
	}
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	up.AddRange(c, hash, 1, -1)
	return nil
}
