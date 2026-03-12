package tgbot

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func cmdExport(c tele.Context) error {
	torrents := torr.ListTorrent()
	if len(torrents) == 0 {
		return c.Send(tr(c.Sender().ID, "no_torrents"))
	}
	uid := c.Sender().ID

	var magnets strings.Builder
	var hashes strings.Builder
	fmt.Fprintf(&hashes, "📁 <b>%s</b> (%d)\n\n", tr(uid, "export_title"), len(torrents))

	for i, t := range torrents {
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
		fmt.Fprintf(&hashes, "%d. %s\n<code>%s</code>\n\n", i+1, escapeHtml(title), hash)
	}

	msg := strings.TrimSuffix(hashes.String(), "\n\n")
	if len(msg) > 4000 {
		msg = msg[:4000] + "\n..."
	}
	if err := c.Send(msg); err != nil {
		return err
	}

	doc := &tele.Document{}
	doc.FileName = "torrents.txt"
	doc.FileReader = bytes.NewReader([]byte(strings.TrimSuffix(magnets.String(), "\n")))
	doc.Caption = "📁 " + tr(uid, "export_file_caption")
	return c.Send(doc)
}
