package tgbot

import (
	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/torr"
)

func list(c tele.Context) error {
	torrents := torr.ListTorrent()

	for _, t := range torrents {
		btnFiles := tele.InlineButton{Text: "Файлы", Unique: "files", Data: t.Hash().String()}
		btnDelete := tele.InlineButton{Text: "Удалить", Unique: "delete", Data: t.Hash().String()}
		torrKbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnFiles, btnDelete}}}
		if t.Size > 0 {
			c.Send("<b>"+t.Title+"</b> <i>"+humanize.Bytes(uint64(t.Size))+"</i>", torrKbd)
		} else {
			c.Send("<b>"+t.Title+"</b>", torrKbd)
		}
	}

	if len(torrents) == 0 {
		c.Send("Нет торрентов")
	}

	return nil
}
