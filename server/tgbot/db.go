package tgbot

import (
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	sets "server/settings"
)

func cmdDb(c tele.Context) error {
	uid := c.Sender().ID
	dbList := sets.ListTorrent()
	if len(dbList) == 0 {
		return c.Send(tr(uid, "db_empty"))
	}

	var sb strings.Builder
	sb.WriteString("📁 <b>" + tr(uid, "db_title") + "</b>\n\n")
	for i, t := range dbList {
		hash := t.InfoHash.HexString()
		sb.WriteString(strconv.Itoa(i+1) + ". <b>" + escapeHtml(t.Title) + "</b>")
		if t.Size > 0 {
			sb.WriteString(" <i>" + humanize.Bytes(uint64(t.Size)) + "</i>")
		}
		sb.WriteString("\n<code>" + hash + "</code>\n\n")
	}
	msg := sb.String()
	if len(msg) > 4000 {
		msg = msg[:4000] + "\n..."
	}
	return c.Send(msg)
}
