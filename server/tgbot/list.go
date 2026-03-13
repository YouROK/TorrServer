package tgbot

import (
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/log"
	"server/torr"
)

const listPageSize = 5

func list(c tele.Context) error {
	args := c.Args()
	compact := len(args) > 0 && strings.ToLower(args[0]) == "compact"
	return sendListPage(c, 0, compact)
}

func sendListPage(c tele.Context, page int, compact bool) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}

	totalPages := (len(torrents) + listPageSize - 1) / listPageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * listPageSize
	end := start + listPageSize
	if end > len(torrents) {
		end = len(torrents)
	}
	pageTorrents := torrents[start:end]

	uid := c.Sender().ID
	for _, t := range pageTorrents {
		hash := t.Hash().HexString()
		var rows [][]tele.InlineButton
		if compact {
			rows = [][]tele.InlineButton{
				{
					tele.InlineButton{Text: tr(uid, "btn_files"), Unique: "files", Data: hash},
					tele.InlineButton{Text: tr(uid, "btn_status"), Unique: "fstatus", Data: hash},
					tele.InlineButton{Text: tr(uid, "btn_delete"), Unique: "delete", Data: hash},
				},
			}
		} else {
			rows = [][]tele.InlineButton{
				{
					tele.InlineButton{Text: tr(uid, "btn_files"), Unique: "files", Data: hash},
					tele.InlineButton{Text: tr(uid, "btn_delete"), Unique: "delete", Data: hash},
					tele.InlineButton{Text: tr(uid, "btn_status"), Unique: "fstatus", Data: hash},
					tele.InlineButton{Text: tr(uid, "btn_m3u"), Unique: "fm3u", Data: hash},
				},
				{
					tele.InlineButton{Text: tr(uid, "btn_link"), Unique: "flink", Data: hash},
					tele.InlineButton{Text: tr(uid, "btn_drop"), Unique: "fdrop", Data: hash},
				},
			}
		}
		torrKbd := &tele.ReplyMarkup{InlineKeyboard: rows}
		msg := "<b>" + escapeHtml(t.Title) + "</b>"
		if t.Size > 0 {
			msg += " <i>" + humanize.IBytes(uint64(t.Size)) + "</i>"
		}
		msg += "\n<code>" + hash + "</code>"
		if err := c.Send(msg, torrKbd); err != nil {
			log.TLogln("tg list send err", err)
			return err
		}
	}

	compactStr := "0"
	if compact {
		compactStr = "1"
	}
	navRow := []tele.InlineButton{}
	if totalPages > 1 {
		if page > 0 {
			navRow = append(navRow, tele.InlineButton{Text: "◀️", Unique: "flist", Data: strconv.Itoa(page-1) + "|" + compactStr})
		}
		navRow = append(navRow, tele.InlineButton{Text: strconv.Itoa(page+1) + "/" + strconv.Itoa(totalPages), Unique: "fnop", Data: ""})
		if page < totalPages-1 {
			navRow = append(navRow, tele.InlineButton{Text: "▶️", Unique: "flist", Data: strconv.Itoa(page+1) + "|" + compactStr})
		}
	}
	navRow = append(navRow, tele.InlineButton{Text: "🔄", Unique: "frefresh", Data: strconv.Itoa(page) + "|" + compactStr})
	if len(navRow) > 1 || totalPages == 1 {
		if err := c.Send(tr(uid, "page")+" "+strconv.Itoa(page+1)+"/"+strconv.Itoa(totalPages), &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{navRow}}); err != nil {
			log.TLogln("tg list nav err", err)
			return err
		}
	}
	return nil
}

func callbackListPage(c tele.Context, data string) error {
	parts := strings.Split(data, "|")
	page := 0
	compact := false
	if len(parts) > 0 && parts[0] != "" {
		if p, err := strconv.Atoi(parts[0]); err == nil {
			page = p
		}
	}
	if len(parts) > 1 && parts[1] == "1" {
		compact = true
	}
	_ = c.Respond(&tele.CallbackResponse{})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendListPage(c, page, compact)
}

func callbackListRefresh(c tele.Context, data string) error {
	parts := strings.Split(data, "|")
	page := 0
	compact := false
	if len(parts) > 0 && parts[0] != "" {
		if p, err := strconv.Atoi(parts[0]); err == nil {
			page = p
		}
	}
	if len(parts) > 1 && parts[1] == "1" {
		compact = true
	}
	_ = c.Respond(&tele.CallbackResponse{Text: "🔄"})
	if c.Callback().Message != nil {
		_ = c.Bot().Delete(c.Callback().Message)
	}
	return sendListPage(c, page, compact)
}
