package tgbot

import tele "gopkg.in/telebot.v4"

func handleCallbackExport(c tele.Context, args []string) error {
	switch args[0] {
	case "\ffexport":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackExportPage(c, data)
	case "\ffexportrefresh":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackExportRefresh(c, data)
	case "\ffhash":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackHashPage(c, data)
	case "\ffhashrefresh":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackHashRefresh(c, data)
	case "\ffstatusall":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackStatusAllPage(c, data)
	case "\ffstatusallrefresh":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackStatusAllRefresh(c, data)
	case "\ffdb":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackDbPage(c, data)
	case "\ffdbrefresh":
		data := ""
		if len(args) > 1 {
			data = args[1]
		}
		return callbackDbRefresh(c, data)
	}
	return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
}
