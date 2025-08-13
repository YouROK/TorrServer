package tgbot

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"

	"server/log"
	"server/tgbot/config"
	up "server/tgbot/upload"
)

func Start(token string) {
	config.LoadConfig()

	pref := tele.Settings{
		URL:       config.Cfg.HostTG,
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

	if len(config.Cfg.WhiteIds) > 0 {
		b.Use(middleware.Whitelist(config.Cfg.WhiteIds...))
	}
	if len(config.Cfg.BlackIds) > 0 {
		b.Use(middleware.Blacklist(config.Cfg.BlackIds...))
	}

	// Commands
	b.Handle("help", help)
	b.Handle("Help", help)
	b.Handle("/help", help)
	b.Handle("/Help", help)
	b.Handle("/start", help)
	b.Handle("/id", help)

	b.Handle("/list", list)
	b.Handle("/clear", clear)

	// Text
	b.Handle(tele.OnText, func(c tele.Context) error {
		txt := c.Text()
		if strings.HasPrefix(strings.ToLower(txt), "magnet:") || isHash(txt) {
			err := addTorrent(c, txt)
			if err != nil {
				return err
			}
			return list(c)
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.ReplyMarkup != nil && len(c.Message().ReplyTo.ReplyMarkup.InlineKeyboard) > 0 {
			data := c.Message().ReplyTo.ReplyMarkup.InlineKeyboard[0][0].Data
			if strings.HasPrefix(strings.ToLower(data), "\fall|") {
				hash := strings.TrimPrefix(data, "\fall|")
				from, to, err := ParseRange(c.Message().Text)
				if err != nil {
					c.Send("Ошибка, нужно указывать числа, пример: 2-12")
					return err
				}
				up.AddRange(c, hash, from, to)
			}
			return nil
		} else {
			return c.Send("Вставьте магнет/хэш торрента чтоб добавить его на сервер")
		}
	})

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		args := c.Args()
		if len(args) > 0 {
			if args[0] == "\ffiles" {
				return files(c)
			}
			if args[0] == "\fdelete" {
				deleteTorrent(c)
				return list(c)
			}
			if args[0] == "\fupload" {
				return upload(c)
			}
			if args[0] == "\fuploadall" {
				uploadall(c)
				return nil
			}
			if args[0] == "\fcancel" {
				if num, err := strconv.Atoi(args[1]); err == nil {
					up.Cancel(num)
					c.Bot().Delete(c.Callback().Message)
					return nil
				}
			}
		}
		return errors.New("Ошибка кнопка не распознана")
	})

	up.Start()

	go b.Start()
}

func help(c tele.Context) error {
	id := strconv.FormatInt(c.Sender().ID, 10)
	var arr []string
	if c.Sender().Username != "" {
		arr = append(arr, c.Sender().Username)
	}
	if c.Sender().FirstName != "" {
		arr = append(arr, c.Sender().FirstName)
	}
	if c.Sender().LastName != "" {
		arr = append(arr, c.Sender().LastName)
	}
	return c.Send("Бот для управления TorrServer\n\n" +
		"Список комманд:\n" +
		"  /help - Эта справка\n" +
		"  /list - Показать список торрентов на сервере\n" +
		"  /clear - Удалить все торренты\n\n" +
		"Ваш id: <code>" + id + "</code>, " + strings.Join(arr, ", "))
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

func ParseRange(rng string) (int, int, error) {
	parts := strings.Split(rng, "-")

	if len(parts) != 2 {
		return -1, -1, errors.New("Неверный формат строки")
	}

	num1, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err1 != nil {
		return -1, -1, err1
	}

	num2, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err2 != nil {
		return -1, -1, err2
	}
	return num1, num2, nil
}
