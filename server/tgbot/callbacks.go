package tgbot

import (
	"strconv"

	tele "gopkg.in/telebot.v4"
	up "server/tgbot/upload"
)

// handleCallback routes callback queries to appropriate handlers
func handleCallback(c tele.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}

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
	case "\ffadd":
		if len(args) > 1 {
			return callbackSearchAdd(c, args[1])
		}
	case "\ffmore":
		if len(args) > 1 {
			return callbackSearchMore(c, args[1])
		}
	case "\fflist":
		if len(args) > 1 {
			return callbackListPage(c, args[1])
		}
	case "\ffrefresh":
		if len(args) > 1 {
			return callbackListRefresh(c, args[1])
		}
	case "\ffnop":
		return c.Respond(&tele.CallbackResponse{})
	case "\ffpreload":
		if len(args) >= 3 {
			return callbackPreload(c, args[1], args[2])
		}
	case "\ffclear":
		if len(args) > 1 {
			return clearConfirm(c, args[1])
		}
	case "\ffshutdown":
		if len(args) > 1 {
			return shutdownConfirm(c, args[1])
		}
	case "\ffset":
		if len(args) > 1 {
			action := args[1]
			for i := 2; i < len(args); i++ {
				action += "|" + args[i]
			}
			return settingsCallback(c, action)
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
