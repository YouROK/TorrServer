package tgbot

import (
	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func deleteTorrent(c tele.Context) {
	args := c.Args()
	hash := args[1]
	torr.RemTorrent(hash)
	return
}

func clear(c tele.Context) error {
	torrents := torr.ListTorrent()
	for _, t := range torrents {
		torr.RemTorrent(t.TorrentSpec.InfoHash.HexString())
	}
	return nil
}
