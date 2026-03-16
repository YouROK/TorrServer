package tgbot

import (
	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdShutdown(c tele.Context) error {
	uid := c.Sender().ID
	btnYes := tele.InlineButton{Text: tr(uid, "btn_yes"), Unique: "fshutdown", Data: "1"}
	btnNo := tele.InlineButton{Text: tr(uid, "btn_no"), Unique: "fshutdown", Data: "0"}
	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnYes, btnNo}}}
	return c.Send(tr(uid, "shutdown_confirm"), kbd)
}

func shutdownConfirm(c tele.Context, confirm string) error {
	uid := c.Sender().ID
	if !isAdmin(uid) {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "admin_only")})
	}
	if confirm != "1" {
		_ = c.Respond(&tele.CallbackResponse{Text: tr(uid, "canceled")})
		return c.Bot().Delete(c.Callback().Message)
	}
	_ = c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "server_stopped")})
	_ = c.Bot().Delete(c.Callback().Message)
	_ = c.Send(tr(c.Sender().ID, "server_stopped"))
	go func() {
		torr.Shutdown()
	}()
	return nil
}
