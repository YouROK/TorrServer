package tgbot

import tele "gopkg.in/telebot.v4"

func handleCallbackSearch(c tele.Context, args []string) error {
	switch args[0] {
	case "\ffadd":
		if len(args) > 1 {
			return callbackSearchAdd(c, args[1])
		}
	case "\ffmore":
		if len(args) > 1 {
			return callbackSearchMore(c, args[1])
		}
	}
	return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
}
