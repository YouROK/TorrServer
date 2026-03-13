package tgbot

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/anacrolix/torrent"
	tele "gopkg.in/telebot.v4"
	"server/log"
	set "server/settings"
	"server/torr"
	"server/web/api/utils"
)

func addTorrentFromSpec(c tele.Context, torrSpec *torrent.TorrentSpec, displayLabel string) error {
	msg, err := c.Bot().Send(c.Sender(), tr(c.Sender().ID, "connecting"))
	if err != nil {
		return err
	}

	tor, err := torr.AddTorrent(torrSpec, "", "", "", "")
	if err != nil {
		log.TLogln("tg add err", err)
		_, _ = c.Bot().Edit(msg, fmt.Sprintf(tr(c.Sender().ID, "add_error"), err.Error()))
		return err
	}
	if tor == nil {
		_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "add_not_created"))
		return errors.New("torrent not created")
	}

	if set.BTsets != nil && set.BTsets.EnableDebug {
		if tor.Data != "" {
			log.TLogln("tg add data", logSafeStr(tor.Data, 60))
		}
		if tor.Category != "" {
			log.TLogln("tg add category", logSafeStr(tor.Category, 40))
		}
	}

	_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "add_getting_meta"))
	if !tor.GotInfo() {
		log.TLogln("tg add err", "timeout get torrent info")
		_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "add_timeout"))
		return errors.New("timeout connection get torrent info")
	}

	if tor.Title == "" {
		tor.Title = torrSpec.DisplayName
		tor.Title = strings.ReplaceAll(tor.Title, "rutor.info", "")
		tor.Title = strings.ReplaceAll(tor.Title, "_", " ")
		tor.Title = strings.Trim(tor.Title, " ")
		if tor.Title == "" {
			tor.Title = tor.Name()
		}
	}

	torr.SaveTorrentToDB(tor)

	if len(displayLabel) > 80 {
		displayLabel = displayLabel[:77] + "..."
	}
	_, _ = c.Bot().Edit(msg, fmt.Sprintf(tr(c.Sender().ID, "add_success"), displayLabel))

	return nil
}

func addTorrent(c tele.Context, link string) error {
	log.TLogln("tg add torrent", logHashOrTruncate(link))
	link = strings.ReplaceAll(link, "&amp;", "&")
	var torrSpec *torrent.TorrentSpec
	var err error
	if strings.HasPrefix(strings.ToLower(link), "torrs://") {
		torrSpec, _, err = utils.ParseTorrsHash(link)
	} else {
		torrSpec, err = utils.ParseLink(link)
	}
	if err != nil {
		log.TLogln("tg add parse err", err)
		return err
	}
	return addTorrentFromSpec(c, torrSpec, link)
}

func addTorrentFromDocument(c tele.Context, doc *tele.Document) error {
	if doc == nil || doc.FileID == "" {
		return errors.New("no document")
	}
	reader, err := c.Bot().File(&doc.File)
	if err != nil {
		log.TLogln("tg add document getfile err", err)
		return err
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	if err != nil {
		log.TLogln("tg add document read err", err)
		return err
	}
	torrSpec, err := utils.ParseFromBytes(data)
	if err != nil {
		log.TLogln("tg add document parse err", err)
		return err
	}
	displayLabel := doc.FileName
	if displayLabel == "" {
		displayLabel = ".torrent"
	}
	return addTorrentFromSpec(c, torrSpec, displayLabel)
}

func cmdAdd(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) == 0 {
		return c.Send(tr(uid, "add_usage"))
	}
	link := strings.TrimSpace(strings.Join(args, " "))
	if link == "" {
		return c.Send(tr(uid, "add_no_link"))
	}
	err := addTorrent(c, link)
	if err != nil {
		return err
	}
	return list(c)
}
