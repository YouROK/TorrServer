package tgbot

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"

	"server/log"
	sets "server/settings"
	"server/torr"
)

func files(c tele.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	hash := args[1]
	if !isHash(hash) {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	msg, err := c.Bot().Send(c.Sender(), tr(c.Sender().ID, "connecting"))
	t := torr.GetTorrent(hash)
	if t == nil {
		if err == nil {
			_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "torrent_not_found")+":\n<code>"+hash+"</code>")
		}
		return nil
	}
	if err == nil {
		api := c.Bot()
		recipient := c.Sender()
		uid := c.Sender().ID
		go sendFilesList(api, recipient, msg, hash, uid)
	}
	return err
}

func sendFilesList(api tele.API, recipient tele.Recipient, statusMsg *tele.Message, hash string, uid int64) {
	t := torr.GetTorrent(hash)
	for t != nil && !t.WaitInfo() {
		time.Sleep(time.Second)
		t = torr.GetTorrent(hash)
	}
	_ = api.Delete(statusMsg)
	t = torr.GetTorrent(hash)
	if t == nil {
		return
	}
	ti := t.Status()
	if ti == nil {
		return
	}

	host := getHost()
	viewedSet := make(map[int]struct{})
	for _, v := range sets.ListViewed(ti.Hash) {
		viewedSet[v.FileIndex] = struct{}{}
	}

	txt := "📁 <b>" + escapeHtml(ti.Title) + "</b> " +
		"<i>" + humanize.IBytes(uint64(ti.TorrentSize)) + "</i>\n\n" +
		"<code>" + ti.Hash + "</code>"

	filesKbd := &tele.ReplyMarkup{}
	var files []tele.Row

	for _, f := range ti.FileStats {
		viewedMark := ""
		if _, ok := viewedSet[f.Id]; ok {
			viewedMark = "✓ "
		}
		fileLabel := viewedMark + "#" + strconv.Itoa(f.Id) + ": " + humanize.IBytes(uint64(f.Length)) + "\n" + filepath.Base(f.Path)
		btn := filesKbd.Data(fileLabel, "upload", ti.Hash, strconv.Itoa(f.Id))
		linkBtn := filesKbd.URL(tr(uid, "files_link"), host+"/stream/"+filepath.Base(f.Path)+"?link="+t.Hash().HexString()+"&index="+strconv.Itoa(f.Id)+"&play")
		btnPreload := filesKbd.Data("⏳", "fpreload", ti.Hash, strconv.Itoa(f.Id))
		files = append(files, filesKbd.Row(btn, linkBtn, btnPreload))
		if len(files) > 99 {
			sendKbd := &tele.ReplyMarkup{}
			sendKbd.Inline(files...)
			if _, err := api.Send(recipient, txt, sendKbd); err != nil {
				log.TLogln("tg files send err", err)
				return
			}
			files = files[:0]
		}
	}

	if len(files) > 0 {
		filesKbd.Inline(files...)
		if _, err := api.Send(recipient, txt, filesKbd); err != nil {
			log.TLogln("tg files send err", err)
			return
		}
	}

	if len(ti.FileStats) > 1 {
		txt = "📁 <b>" + escapeHtml(ti.Title) + "</b> " +
			"<i>" + humanize.IBytes(uint64(ti.TorrentSize)) + "</i>\n\n" +
			"<code>" + ti.Hash + "</code>\n\n" +
			fmt.Sprintf(tr(uid, "files_range_hint"), len(ti.FileStats))
		files = files[:0]
		files = append(files, filesKbd.Row(filesKbd.Data(tr(uid, "files_download_all"), "fall", "all|"+ti.Hash)))
		filesKbd.Inline(files...)
		if _, err := api.Send(recipient, txt, filesKbd); err != nil {
			log.TLogln("tg files send err", err)
		}
	}
}
