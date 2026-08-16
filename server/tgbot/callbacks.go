package tgbot

import tele "gopkg.in/telebot.v4"

// handleCallback routes callback queries to appropriate handlers
func handleCallback(c tele.Context) error {
	if c.Sender() == nil {
		return nil
	}
	args := c.Args()
	if len(args) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}

	switch args[0] {
	case "\ffiles", "\fdelete", "\fupload", "\fuploadall", "\ffall", "\fcancel",
		"\ffstatus", "\ffm3u", "\fflink", "\ffdrop", "\ffstatusrefresh", "\ffstatusstop",
		"\fflist", "\ffrefresh", "\ffnop", "\ffpreload", "\ffitems", "\ffifresh",
		"\ffsnakerefresh", "\ffsnakestop":
		return handleCallbackTorrent(c, args)
	case "\ffadd", "\ffmore":
		return handleCallbackSearch(c, args)
	case "\ffexport", "\ffexportrefresh", "\ffhash", "\ffhashrefresh",
		"\ffstatusall", "\ffstatusallrefresh", "\ffdb", "\ffdbrefresh":
		return handleCallbackExport(c, args)
	case "\ffclear", "\ffshutdown", "\ffpreset", "\ffset":
		return handleCallbackAdmin(c, args)
	default:
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
}
