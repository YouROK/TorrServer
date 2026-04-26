package tgbot

import tele "gopkg.in/telebot.v4"

func handleCallbackAdmin(c tele.Context, args []string) error {
	switch args[0] {
	case "\ffclear":
		if len(args) > 1 {
			return clearConfirm(c, args[1])
		}
	case "\ffshutdown":
		if len(args) > 1 {
			return shutdownConfirm(c, args[1])
		}
	case "\ffpreset":
		if len(args) > 1 {
			return presetConfirm(c, args[1])
		}
	case "\ffset":
		if len(args) > 1 {
			action := args[1]
			for i := 2; i < len(args); i++ {
				action += "|" + args[i]
			}
			return settingsCallback(c, action)
		}
	}
	return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
}
