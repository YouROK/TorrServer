package tgbot

import (
	"encoding/json"
	"errors"
	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
	"net/http"
	"os"
	"path/filepath"
	"server/log"
	"server/settings"
	"server/torr"
	"server/web"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	WhiteIds []int64
}

var cfg *Config

func init() {
	cfg = &Config{}
	fn := filepath.Join(settings.Path, "tg.cfg")
	buf, err := os.ReadFile(fn)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &cfg)
	if err != nil {
		log.TLogln("Error read tg config:", err)
	}
}

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

	if len(cfg.WhiteIds) > 0 {
		b.Use(middleware.Whitelist(cfg.WhiteIds...))
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
				msg, err := c.Bot().Send(c.Sender(), "Подключение к торренту...")
				t := torr.GetTorrent(args[1])
				if t == nil {
					c.Edit(msg, "Torrent not connected: "+args[1])
					return nil
				}
				if err == nil {
					go func() {
						for !t.WaitInfo() {
							time.Sleep(time.Second)
							t = torr.GetTorrent(args[1])
						}
						c.Bot().Delete(msg)
						host := settings.PubIPv4
						if host == "" {
							ips := web.GetLocalIps()
							if len(ips) == 0 {
								host = "127.0.0.1"
							} else {
								host = ips[0]
							}
						}

						t = torr.GetTorrent(args[1])
						st := t.Status()
						txt := "Файлы:\n"
						for _, file := range st.FileStats {
							ff := "<b>" + filepath.Base(file.Path) + "</b> <i>" + humanize.Bytes(uint64(file.Length)) + "</i> " +
								"<a href=\"http://" + host + ":" + settings.Port + "/stream/" + filepath.Base(file.Path) + "?link=" + t.Hash().HexString() + "&index=" + strconv.Itoa(file.Id) + "&play\">Download</a>\n\n"
							if len(txt+ff) > 4096 {
								c.Send(txt)
								txt = ""
							}
							txt += ff
						}
						if len(txt) > 0 {
							c.Send(txt)
						}
					}()
				}
				return err
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
		c.Send(t.Title+" "+humanize.Bytes(uint64(t.Size)), torrKbd)
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
