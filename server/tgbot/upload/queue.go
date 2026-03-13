package upload

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/torr"
)

type DLQueue struct {
	id        int
	c         tele.Context
	hash      string
	fileID    string
	fileName  string
	updateMsg *tele.Message
}

var manager = &Manager{}

func Start() {
	manager.Start()
}

func ShowQueue(c tele.Context) error {
	msg := ""
	manager.queueLock.Lock()
	defer manager.queueLock.Unlock()
	if len(manager.queue) == 0 && len(manager.working) == 0 {
		return c.Send(tr(c.Sender().ID, "queue_empty"))
	}
	if len(manager.working) > 0 {
		msg += tr(c.Sender().ID, "upload_working") + ":\n"
		i := 0
		for _, dlQueue := range manager.working {
			s := "#" + strconv.Itoa(i+1) + ": <code>" + dlQueue.torrentHash + "</code>\n"
			if len(msg+s) > 1024 {
				c.Send(msg)
				msg = ""
			}
			msg += s
			i++
		}
		if len(msg) > 0 {
			c.Send(msg)
			msg = ""
		}
	}
	if len(manager.queue) > 0 {
		msg = tr(c.Sender().ID, "upload_in_queue") + ":\n"
		for i, dlQueue := range manager.queue {
			s := "#" + strconv.Itoa(i+1) + ": <code>" + dlQueue.torrentHash + "</code>\n"
			if len(msg+s) > 1024 {
				c.Send(msg)
				msg = ""
			}
			msg += s
		}
		if len(msg) > 0 {
			c.Send(msg)
			msg = ""
		}
	}
	return nil
}

func AddRange(c tele.Context, hash string, from, to int) {
	manager.AddRange(c, hash, from, to)
}

func Cancel(id int) {
	manager.Cancel(id)
}

func updateLoadStatus(wrk *Worker, file *TorrFile, fi, fc int) {
	if wrk.msg == nil {
		return
	}
	t := torr.GetTorrent(wrk.torrentHash)
	if t == nil {
		return
	}
	ti := t.Status()
	if wrk.isCancelled {
		wrk.c.Bot().Edit(wrk.msg, tr(wrk.c.Sender().ID, "upload_stopping"))
	} else {
		wrk.c.Send(tele.UploadingVideo)
		if ti.DownloadSpeed == 0 {
			ti.DownloadSpeed = 1.0
		}
		wait := time.Duration(float64(file.Remaining())/ti.DownloadSpeed) * time.Second
		speed := humanize.IBytes(uint64(ti.DownloadSpeed)) + "/sec"
		peers := fmt.Sprintf("%v · %v/%v", ti.ConnectedSeeders, ti.ActivePeers, ti.TotalPeers)
		prc := fmt.Sprintf("%.2f%% %v / %v", float64(file.offset)*100.0/float64(file.size), humanize.IBytes(uint64(file.offset)), humanize.IBytes(uint64(file.size)))

		name := file.name
		if name == ti.Title {
			name = ""
		}

		uid := wrk.c.Sender().ID
		msg := tr(uid, "upload_title") + ":\n" +
			"<b>" + escapeHtml(ti.Title) + "</b>\n"
		if name != "" {
			msg += "<i>" + escapeHtml(name) + "</i>\n"
		}
		msg += "<b>" + tr(uid, "upload_hash") + ":</b> <code>" + file.hash + "</code>\n"
		if file.offset < file.size {
			msg += "<b>" + tr(uid, "upload_speed") + ": </b>" + speed + "\n" +
				"<b>" + tr(uid, "upload_remaining") + ": </b>" + wait.String() + "\n" +
				"<b>" + tr(uid, "upload_peers") + ": </b>" + peers + "\n" +
				"<b>" + tr(uid, "upload_progress") + ": </b>" + prc
		}
		if fc > 1 {
			msg += "\n<b>" + tr(uid, "upload_files") + ": </b>" + strconv.Itoa(fi) + "/" + strconv.Itoa(fc)
		}
		if file.offset >= file.size {
			msg += "\n<b>" + tr(uid, "upload_finishing") + "</b>"
			wrk.c.Bot().Edit(wrk.msg, msg)
			return
		}

		torrKbd := &tele.ReplyMarkup{}
		torrKbd.Inline([]tele.Row{torrKbd.Row(torrKbd.Data(tr(wrk.c.Sender().ID, "upload_cancel"), "cancel", strconv.Itoa(wrk.id)))}...)
		wrk.c.Bot().Edit(wrk.msg, msg, torrKbd)
	}
}
