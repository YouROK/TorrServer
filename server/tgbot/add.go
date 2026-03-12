package tgbot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/anacrolix/torrent"
	tele "gopkg.in/telebot.v4"
	"server/log"
	set "server/settings"
	"server/torr"
	"server/web/api/utils"
)

func addTorrent(c tele.Context, link string) error {
	msg, err := c.Bot().Send(c.Sender(), tr(c.Sender().ID, "connecting"))
	if err != nil {
		return err
	}
	log.TLogln("tg add torrent", logHashOrTruncate(link))
	link = strings.ReplaceAll(link, "&amp;", "&")
	var torrSpec *torrent.TorrentSpec
	if strings.HasPrefix(strings.ToLower(link), "torrs://") {
		torrSpec, _, err = utils.ParseTorrsHash(link)
	} else {
		torrSpec, err = utils.ParseLink(link)
	}
	if err != nil {
		log.TLogln("tg add parse err", err)
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

	displayLink := link
	if len(displayLink) > 80 {
		displayLink = displayLink[:77] + "..."
	}
	_, _ = c.Bot().Edit(msg, fmt.Sprintf(tr(c.Sender().ID, "add_success"), displayLink))

	return nil
}
