package tgbot

import (
	"strings"

	tele "gopkg.in/telebot.v4"
)

func cmdLang(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) == 0 {
		lang := getUserLang(uid)
		if lang == LangEN {
			return c.Send(tr(uid, "lang_current_en") + "\n/lang RU — " + tr(uid, "lang_switch_ru"))
		}
		return c.Send(tr(uid, "lang_current_ru") + "\n/lang EN — " + tr(uid, "lang_switch_en"))
	}
	lang := strings.ToUpper(strings.TrimSpace(args[0]))
	if lang == "EN" {
		setUserLang(uid, LangEN)
		return c.Send(tr(uid, "lang_set_en"))
	}
	if lang == "RU" || lang == "РУ" {
		setUserLang(uid, LangRU)
		return c.Send(tr(uid, "lang_set"))
	}
	return c.Send(tr(uid, "lang_usage"))
}
