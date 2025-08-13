package tgbot

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"

	"server/log"
	"server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/web"
)

func files(c tele.Context) error {
	args := c.Args()
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
			host := config.Cfg.HostWeb
			if host == "" {
				host = settings.PubIPv4
				if host == "" {
					ips := web.GetLocalIps()
					if len(ips) == 0 {
						host = "127.0.0.1"
					} else {
						host = ips[0]
					}
				}
			}
			if !strings.Contains(host, ":") {
				host += ":" + settings.Port
			}
			if !strings.HasPrefix(host, "http") {
				host = "http://" + host
			}

			t = torr.GetTorrent(args[1])
			ti := t.Status()

			txt := "<b>" + ti.Title + "</b> " +
				"<i>" + humanize.Bytes(uint64(ti.TorrentSize)) + "</i>\n\n" +
				"<code>" + ti.Hash + "</code>"

			filesKbd := &tele.ReplyMarkup{}
			var files []tele.Row

			i := len(txt)
			for _, f := range ti.FileStats {
				btn := filesKbd.Data("#"+strconv.Itoa(f.Id)+": "+humanize.Bytes(uint64(f.Length))+"\n"+filepath.Base(f.Path), "upload", ti.Hash, strconv.Itoa(f.Id))
				link := filesKbd.URL("Ссылка", host+"/stream/"+filepath.Base(f.Path)+"?link="+t.Hash().HexString()+"&index="+strconv.Itoa(f.Id)+"&play")
				files = append(files, filesKbd.Row(btn, link))
				if i+len(txt) > 1024 || len(files) > 99 {
					filesKbd := &tele.ReplyMarkup{}
					filesKbd.Inline(files...)
					err = c.Send(txt, filesKbd)
					if err != nil {
						log.TLogln("Error send message files:", err)
						return
					}
					files = files[:0]
					i = len(txt)
				}
				i += len(btn.Text + link.Text)
			}

			if len(files) > 0 {
				filesKbd.Inline(files...)
				err = c.Send(txt, filesKbd)
				if err != nil {
					log.TLogln("Error send message files:", err)
					return
				}
			}

			if len(files) > 1 {
				txt = "<b>" + ti.Title + "</b> " +
					"<i>" + humanize.Bytes(uint64(ti.TorrentSize)) + "</i>\n\n" +
					"<code>" + ti.Hash + "</code>\n\n" +
					"Чтобы скачать несколько файлов, ответьте на это сообщение, с какого файла скачать по какой, пример: 2-12\n\n" +
					"Скачать все файлы? Всего:" + strconv.Itoa(len(ti.FileStats))
				files = files[:0]
				files = append(files, filesKbd.Row(filesKbd.Data("Скачать все файлы", "uploadall", ti.Hash)))
				filesKbd.Inline(files...)
				err = c.Send(txt, filesKbd)
				if err != nil {
					log.TLogln("Error send message files:", err)
					return
				}
			}
		}()
	}
	return err
}
