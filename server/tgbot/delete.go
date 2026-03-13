package tgbot

import (
	"fmt"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func deleteTorrent(c tele.Context) {
	args := c.Args()
	if len(args) < 2 {
		return
	}
	hash := args[1]
	if !isHash(hash) {
		return
	}
	torr.RemTorrent(hash)
}

func clear(c tele.Context) error {
	torrents := torr.ListTorrent()
	count := len(torrents)
	if count == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}
	uid := c.Sender().ID
	btnYes := tele.InlineButton{Text: tr(uid, "btn_yes"), Unique: "fclear", Data: "1"}
	btnNo := tele.InlineButton{Text: tr(uid, "btn_no"), Unique: "fclear", Data: "0"}
	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnYes, btnNo}}}
	return c.Send(fmt.Sprintf(tr(uid, "clear_confirm"), count), kbd)
}

func clearConfirm(c tele.Context, confirm string) error {
	uid := c.Sender().ID
	if confirm != "1" {
		_ = c.Respond(&tele.CallbackResponse{Text: tr(uid, "canceled")})
		return c.Bot().Delete(c.Callback().Message)
	}
	torrents := torr.ListTorrent()
	count := len(torrents)
	for _, t := range torrents {
		torr.RemTorrent(t.Hash().HexString())
	}
	_ = c.Respond(&tele.CallbackResponse{Text: tr(uid, "deleted")})
	_ = c.Bot().Delete(c.Callback().Message)
	return c.Send(fmt.Sprintf(tr(uid, "clear_done"), count))
}
