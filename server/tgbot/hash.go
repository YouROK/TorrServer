package tgbot

import (
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

// resolveHash returns hash from: 1) full hash string, 2) numeric index from list, 3) reply-to message
func resolveHash(c tele.Context, arg string) string {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return extractHashFromReply(c)
	}
	if isHash(arg) {
		return arg
	}
	if idx, err := strconv.Atoi(arg); err == nil && idx > 0 {
		torrents := torr.ListTorrent()
		if idx <= len(torrents) {
			return torrents[idx-1].Hash().HexString()
		}
	}
	return ""
}

func extractHashFromReply(c tele.Context) string {
	if c.Message() == nil || c.Message().ReplyTo == nil {
		return ""
	}
	reply := c.Message().ReplyTo
	if reply.ReplyMarkup == nil || len(reply.ReplyMarkup.InlineKeyboard) == 0 {
		return ""
	}
	for _, row := range reply.ReplyMarkup.InlineKeyboard {
		for _, btn := range row {
			if btn.Data == "" {
				continue
			}
			if isHash(btn.Data) {
				return btn.Data
			}
			if idx := strings.Index(btn.Data, "all|"); idx >= 0 {
				h := btn.Data[idx+4:]
				if len(h) >= 40 && isHash(h[:40]) {
					return h[:40]
				}
				if isHash(h) {
					return h
				}
			}
		}
	}
	return ""
}

func cmdHash(c tele.Context) error {
	args := c.Args()
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}

	if len(args) > 0 {
		idx, err := strconv.Atoi(strings.TrimSpace(args[0]))
		if err != nil || idx < 1 || idx > len(torrents) {
			return c.Send(tr(c.Sender().ID, "invalid_index"))
		}
		hash := torrents[idx-1].Hash().HexString()
		return c.Send("🔑 <code>" + hash + "</code>")
	}

	var sb strings.Builder
	sb.WriteString("🔑 <b>" + tr(c.Sender().ID, "hash_title") + "</b>\n\n")
	for i, t := range torrents {
		sb.WriteString(strconv.Itoa(i+1) + ". <code>" + t.Hash().HexString() + "</code>\n")
		sb.WriteString("   " + t.Title + "\n\n")
	}
	msg := sb.String()
	if len(msg) > 4000 {
		msg = msg[:4000] + "\n..."
	}
	return c.Send(msg)
}
