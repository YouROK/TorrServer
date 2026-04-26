package tgbot

import (
	"bytes"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdStat(c tele.Context) error {
	var buf bytes.Buffer
	torr.WriteStatus(&buf)
	msg := buf.String()
	msg = strings.ReplaceAll(msg, "<", "&lt;")
	msg = strings.ReplaceAll(msg, ">", "&gt;")
	if len(msg) > 4000 {
		msg = msg[:4000] + "\n..."
	}
	return c.Send("📋 <b>" + tr(c.Sender().ID, "help_stat") + "</b>\n\n<pre>" + msg + "</pre>")
}
