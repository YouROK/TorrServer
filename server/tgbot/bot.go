package tgbot

import (
	"errors"
	tele "gopkg.in/telebot.v4"
	"net/http"
	"server/log"
	"server/torr"
	"strings"
	"time"
)

func Start(token string) {
	pref := tele.Settings{
		Token:     token,
		Poller:    &tele.LongPoller{Timeout: 5 * time.Minute},
		ParseMode: tele.ModeHTML,
		Client:    &http.Client{Timeout: 5 * time.Minute},
	}

	log.TLogln("Starting Telegram Bot")

	b, err := tele.NewBot(pref)
	if err != nil {
		log.TLogln("Error start tg bot:", err)
		return
	}

	//Commands

	b.Handle("help", help)
	b.Handle("Help", help)
	b.Handle("/help", help)
	b.Handle("/Help", help)
	b.Handle("/start", help)

	b.Handle("/list", list)

	//Text
	b.Handle(tele.OnText, func(c tele.Context) error {
		txt := c.Text()
		if strings.HasPrefix(strings.ToLower(txt), "magnet:") || isHash(txt) {
			return addTorrent(c, txt)
		} else {
			return c.Send("Вставьте магнет/хэш торрента чтоб добавить его на сервер")
		}
	})

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		args := c.Args()
		if len(args) > 0 {
			if args[0] == "\ffiles" {
				t := torr.GetTorrent(args[1])
				if t == nil {
					c.Send("Torrent not connected:", args[1])
				} else {

				}
				return nil
			}
			if args[0] == "\fdelete" {
				return nil
			}
		}
		return errors.New("Ошибка кнопка не распознана")
	})

	go b.Start()
}

func help(c tele.Context) error {
	return c.Send("Бот для управления TorrServer\n\n" +
		"Список комманд:\n" +
		"  /help - эта справка\n" +
		"  /list - показать список торрентов на сервере")
}

func list(c tele.Context) error {
	list := torr.ListTorrent()

	for _, t := range list {
		btnFiles := tele.InlineButton{Text: "Файлы", Unique: "files", Data: t.Hash().String()}
		btnDelete := tele.InlineButton{Text: "Удалить", Unique: "delete", Data: t.Hash().String()}
		torrKbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnFiles, btnDelete}}}
		c.Send(t.Title, torrKbd)
	}
	return nil
}

func isHash(txt string) bool {
	if len(txt) == 40 {
		for _, c := range strings.ToLower(txt) {
			switch c {
			case 'a', 'b', 'c', 'd', 'e', 'f', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			default:
				return false
			}
		}
		return true
	}
	return false
}
