package tgbot

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/log"
	"server/torr"
)

const exportPageSize = 10

func cmdExport(c tele.Context) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}
	uid := c.Sender().ID

	var magnets strings.Builder
	for _, t := range torrents {
		hash := t.Hash().HexString()
		title := t.Title
		if title == "" {
			title = t.Name()
		}
		magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", hash)
		if title != "" {
			magnet += "&dn=" + url.QueryEscape(title)
		}
		magnets.WriteString(magnet + "\n")
	}

	doc := &tele.Document{}
	doc.FileName = "torrents.txt"
	doc.FileReader = bytes.NewReader([]byte(strings.TrimSuffix(magnets.String(), "\n")))
	doc.Caption = "📁 " + tr(uid, "export_file_caption")
	if err := c.Send(doc); err != nil {
		return err
	}

	return sendExportPage(c, 0)
}

func sendExportPage(c tele.Context, page int) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}

	totalPages := (len(torrents) + exportPageSize - 1) / exportPageSize
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	start := page * exportPageSize
	end := start + exportPageSize
	if end > len(torrents) {
		end = len(torrents)
	}
	pageTorrents := torrents[start:end]

	uid := c.Sender().ID
	var hashes strings.Builder
	fmt.Fprintf(&hashes, "📁 <b>%s</b> (%d)\n\n", tr(uid, "export_title"), len(torrents))
	for i, t := range pageTorrents {
		hash := t.Hash().HexString()
		title := t.Title
		if title == "" {
			title = t.Name()
		}
		fmt.Fprintf(&hashes, "%d. %s\n<code>%s</code>\n\n", start+i+1, escapeHtml(title), hash)
	}
	msg := strings.TrimSuffix(hashes.String(), "\n\n")

	navRow := []tele.InlineButton{}
	if totalPages > 1 {
		if page > 0 {
			navRow = append(navRow, tele.InlineButton{Text: "◀️", Unique: "fexport", Data: strconv.Itoa(page - 1)})
		}
		navRow = append(navRow, tele.InlineButton{Text: strconv.Itoa(page+1) + "/" + strconv.Itoa(totalPages), Unique: "fnop", Data: ""})
		if page < totalPages-1 {
			navRow = append(navRow, tele.InlineButton{Text: "▶️", Unique: "fexport", Data: strconv.Itoa(page + 1)})
		}
	}
	navRow = append(navRow, tele.InlineButton{Text: "🔄", Unique: "fexportrefresh", Data: strconv.Itoa(page)})

	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{navRow}}
	if err := c.Send(msg, kbd); err != nil {
		log.TLogln("tg export send err", err)
		return err
	}
	return nil
}

func callbackExportPage(c tele.Context, data string) error {
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
	return sendExportPage(c, page)
}

func callbackExportRefresh(c tele.Context, data string) error {
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
	return sendExportPage(c, page)
}
