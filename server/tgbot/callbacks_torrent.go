package tgbot

import (
	"strconv"

	tele "gopkg.in/telebot.v4"
	up "server/tgbot/upload"
)

func handleCallbackTorrent(c tele.Context, args []string) error {
	switch args[0] {
	case "\ffiles":
		return files(c)
	case "\fdelete":
		if len(args) < 2 {
			return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
		}
		deleteTorrent(c)
		_ = c.Bot().Delete(c.Callback().Message)
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "deleted")})
	case "\fupload":
		return upload(c)
	case "\fuploadall", "\ffall":
		return uploadall(c)
	case "\fcancel":
		if len(args) > 1 {
			if num, err := strconv.Atoi(args[1]); err == nil {
				up.Cancel(num)
				_ = c.Bot().Delete(c.Callback().Message)
				return c.Respond(&tele.CallbackResponse{})
			}
		}
		return c.Respond(&tele.CallbackResponse{})
	case "\ffstatus", "\ffm3u", "\fflink", "\ffdrop", "\ffstatusrefresh", "\ffstatusstop":
		hash := ""
		if len(args) >= 2 {
			hash = args[1]
		}
		switch args[0] {
		case "\ffstatus":
			if hash == "" {
				return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
			}
			return callbackStatus(c, hash)
		case "\ffstatusrefresh":
			return callbackStatusRefresh(c, hash)
		case "\ffstatusstop":
			return callbackStatusStop(c, hash)
		case "\ffm3u":
			if hash == "" {
				return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
			}
			return callbackM3u(c, hash)
		case "\fflink":
			if len(args) < 2 {
				return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
			}
			return callbackLink(c, args[1])
		case "\ffdrop":
			if hash == "" {
				return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
			}
			return callbackDrop(c, hash)
		}
	case "\fflist":
		if len(args) > 1 {
			return callbackListPage(c, args[1])
		}
	case "\ffrefresh":
		if len(args) > 1 {
			return callbackListRefresh(c, args[1])
		}
	case "\ffitems":
		if len(args) >= 3 {
			return callbackFileListPage(c, args[1], args[2])
		}
	case "\ffifresh":
		if len(args) >= 3 {
			return callbackFileListRefresh(c, args[1], args[2])
		}
	case "\ffnop":
		return c.Respond(&tele.CallbackResponse{})
	case "\ffpreload":
		if len(args) >= 3 {
			return callbackPreload(c, args[1], args[2])
		}
	case "\ffsnakerefresh", "\ffsnakestop":
		data := ""
		if len(args) >= 2 {
			data = args[1]
		}
		switch args[0] {
		case "\ffsnakerefresh":
			return callbackSnakeRefresh(c, data)
		case "\ffsnakestop":
			return callbackSnakeStop(c, data)
		}
	}
	return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
}
