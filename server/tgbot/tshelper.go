package tgbot

import (
	tele "gopkg.in/telebot.v4"
	"server/log"
	set "server/settings"
	"server/torr"
	"server/web/api/utils"
	"strings"
)

func addTorrent(c tele.Context, link string) error {
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
		return err
	}

	go func() {
		if !tor.GotInfo() {
			log.TLogln("tg error add torrent: timeout connection get torrent info")
			c.Send("Ошибка при добаваления торрента: timeout connection get torrent info")
			return
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
		c.Send("Торрент добавлен: <code>" + link + "</code>")
	}()

	return nil
}
