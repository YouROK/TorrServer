package tgbot

import (
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/log"
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

const hashPageSize = 10

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

	return sendHashPage(c, 0)
}

func sendHashPage(c tele.Context, page int) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}

	totalPages := (len(torrents) + hashPageSize - 1) / hashPageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * hashPageSize
	end := start + hashPageSize
	if end > len(torrents) {
		end = len(torrents)
	}
	pageTorrents := torrents[start:end]

	uid := c.Sender().ID
	var sb strings.Builder
	sb.WriteString("🔑 <b>" + tr(uid, "hash_title") + "</b> (" + strconv.Itoa(len(torrents)) + ")\n\n")
	for i, t := range pageTorrents {
		sb.WriteString(strconv.Itoa(start+i+1) + ". <code>" + t.Hash().HexString() + "</code>\n")
		sb.WriteString("   " + escapeHtml(t.Title) + "\n\n")
	}
	msg := strings.TrimSuffix(sb.String(), "\n\n")

	navRow := []tele.InlineButton{}
	if totalPages > 1 {
		if page > 0 {
			navRow = append(navRow, tele.InlineButton{Text: "◀️", Unique: "fhash", Data: strconv.Itoa(page - 1)})
		}
		navRow = append(navRow, tele.InlineButton{Text: strconv.Itoa(page+1) + "/" + strconv.Itoa(totalPages), Unique: "fnop", Data: ""})
		if page < totalPages-1 {
			navRow = append(navRow, tele.InlineButton{Text: "▶️", Unique: "fhash", Data: strconv.Itoa(page + 1)})
		}
	}
	navRow = append(navRow, tele.InlineButton{Text: "🔄", Unique: "fhashrefresh", Data: strconv.Itoa(page)})

	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{navRow}}
	if err := c.Send(msg, kbd); err != nil {
		log.TLogln("tg hash send err", err)
		return err
	}
	return nil
}

func callbackHashPage(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(data); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendHashPage(c, page)
}

func callbackHashRefresh(c tele.Context, data string) error {
	page := 0
	if data != "" {
		if p, err := strconv.Atoi(data); err == nil {
			page = p
		}
	}
	_ = c.Respond(&tele.CallbackResponse{Text: "🔄"})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendHashPage(c, page)
}
