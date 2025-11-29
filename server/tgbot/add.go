package tgbot

import (
	"errors"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/log"
	set "server/settings"
	"server/torr"
	"server/web/api/utils"
)

func addTorrent(c tele.Context, link string) error {
	msg, err := c.Bot().Send(c.Sender(), "Подключение к торренту...")
	if err != nil {
		return err
	}
	log.TLogln("tg add torrent", link)
	link = strings.ReplaceAll(link, "&amp;", "&")
	torrSpec, err := utils.ParseLink(link)
	if err != nil {
		log.TLogln("tg error parse link:", err)
		return err
	}

	tor, err := torr.AddTorrent(torrSpec, "", "", "", "")

	if tor.Data != "" && set.BTsets.EnableDebug {
		log.TLogln("torrent data:", tor.Data)
	}
	if tor.Category != "" && set.BTsets.EnableDebug {
		log.TLogln("torrent category:", tor.Category)
	}

	if err != nil {
		log.TLogln("tg error add torrent:", err)
		c.Bot().Edit(msg, "Ошибка при подключении: "+err.Error())
		return err
	}

	if !tor.GotInfo() {
		log.TLogln("tg error add torrent: timeout connection get torrent info")
		c.Bot().Edit(msg, "Ошибка при добаваления торрента: timeout connection get torrent info")
		return errors.New("timeout connection get torrent info")
	}

	if tor.Title == "" {
		tor.Title = torrSpec.DisplayName // prefer dn over name
		tor.Title = strings.ReplaceAll(tor.Title, "rutor.info", "")
		tor.Title = strings.ReplaceAll(tor.Title, "_", " ")
		tor.Title = strings.Trim(tor.Title, " ")
		if tor.Title == "" {
			tor.Title = tor.Name()
		}
	}

	torr.SaveTorrentToDB(tor)

	c.Bot().Edit(msg, "Торрент добавлен:\n<code>"+link+"</code>")

	return nil
}
