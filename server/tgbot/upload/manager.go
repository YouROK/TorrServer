package upload

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"

	"server/log"
	"server/torr"
	"server/torr/state"
)

// TrFunc is set by tgbot for localization (avoids circular import)
var TrFunc func(int64, string) string

// EscapeFunc is set by tgbot for HTML escaping (avoids circular import)
var EscapeFunc func(string) string

func tr(uid int64, key string) string {
	if TrFunc != nil {
		return TrFunc(uid, key)
	}
	return key
}

func escapeHtml(s string) string {
	if EscapeFunc != nil {
		return EscapeFunc(s)
	}
	return s
}

type Worker struct {
	id          int
	c           tele.Context
	msg         *tele.Message
	torrentHash string
	isCancelled bool
	from        int
	to          int
	ti          *state.TorrentStatus
}

type Manager struct {
	queue     []*Worker
	working   map[int]*Worker
	ids       int
	wrkSync   sync.Mutex
	queueLock sync.Mutex
}

func (m *Manager) Start() {
	m.working = make(map[int]*Worker)
	go m.work()
}

func (m *Manager) AddRange(c tele.Context, hash string, from, to int) {
	m.queueLock.Lock()
	defer m.queueLock.Unlock()

	if len(m.queue) > 50 {
		c.Bot().Send(c.Recipient(), fmt.Sprintf(tr(c.Sender().ID, "upload_queue_full"), len(m.queue)))
		return
	}

	m.ids++
	if m.ids > math.MaxInt {
		m.ids = 0
	}

	var msg *tele.Message
	var err error

	for i := 0; i < 20; i++ {
		msg, err = c.Bot().Send(c.Recipient(), fmt.Sprintf(tr(c.Sender().ID, "upload_connecting"), hash))
		if err == nil {
			break
		}
		log.TLogln("tg upload retry", i+1, "/", 20)
		if i < 19 {
			backoff := time.Duration(1<<uint(i)) * 100 * time.Millisecond
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
			time.Sleep(backoff)
		}
	}

	if err != nil {
		log.TLogln("tg upload send err", err)
		return
	}

	t := torr.GetTorrent(hash)
	if t == nil {
		c.Bot().Edit(msg, tr(c.Sender().ID, "torrent_not_found")+":\n<code>"+hash+"</code>")
		return
	}
	t.WaitInfo()
	for t.Status().Stat != state.TorrentWorking {
		time.Sleep(time.Second)
		t = torr.GetTorrent(hash)
		if t == nil {
			return
		}
	}
	ti := t.Status()

	if from == 1 && to == -1 {
		to = len(ti.FileStats)
	}
	if from < 1 {
		from = 1
	}
	if to > len(ti.FileStats) {
		to = len(ti.FileStats)
	}
	if from > to {
		from, to = to, from
	}
	if to > len(ti.FileStats) {
		to = len(ti.FileStats)
	}

	w := &Worker{
		id:          m.ids,
		c:           c,
		torrentHash: hash,
		msg:         msg,
		ti:          ti,
		from:        from,
		to:          to,
	}

	m.queue = append(m.queue, w)
}

func (m *Manager) Cancel(id int) {
	m.queueLock.Lock()
	defer m.queueLock.Unlock()
	for i, w := range m.queue {
		if w.id == id {
			w.isCancelled = true
			w.c.Bot().Delete(w.msg)
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return
		}
	}
	if wrk, ok := m.working[id]; ok {
		wrk.isCancelled = true
		return
	}
}

func (m *Manager) work() {
	for {
		m.queueLock.Lock()
		if len(m.working) > 0 {
			m.queueLock.Unlock()
			m.sendQueueStatus()
			time.Sleep(time.Second)
			continue
		}
		if len(m.queue) == 0 {
			m.queueLock.Unlock()
			time.Sleep(time.Second)
			continue
		}
		wrk := m.queue[0]
		m.queue = m.queue[1:]
		m.working[wrk.id] = wrk
		m.queueLock.Unlock()

		m.sendQueueStatus()

		loading(wrk)

		m.queueLock.Lock()
		delete(m.working, wrk.id)
		m.queueLock.Unlock()
	}
}

func (m *Manager) sendQueueStatus() {
	m.queueLock.Lock()
	defer m.queueLock.Unlock()
	for i, wrk := range m.queue {
		if wrk.msg == nil || wrk.c.Sender() == nil {
			continue
		}
		torrKbd := &tele.ReplyMarkup{}
		torrKbd.Inline([]tele.Row{torrKbd.Row(torrKbd.Data(tr(wrk.c.Sender().ID, "upload_cancel"), "cancel", strconv.Itoa(wrk.id)))}...)

		msg := fmt.Sprintf(tr(wrk.c.Sender().ID, "upload_queue_pos"), i+1)

		wrk.c.Bot().Edit(wrk.msg, msg, torrKbd)
	}
}

func loading(wrk *Worker) {
	iserr := false

	t := torr.GetTorrent(wrk.torrentHash)
	if t == nil {
		wrk.c.Bot().Edit(wrk.msg, tr(wrk.c.Sender().ID, "torrent_not_found")+":\n<code>"+wrk.torrentHash+"</code>")
		return
	}
	t.WaitInfo()
	for t.Status().Stat != state.TorrentWorking {
		time.Sleep(time.Second)
		t = torr.GetTorrent(wrk.torrentHash)
		if t == nil {
			return
		}
	}
	wrk.ti = t.Status()

	for i := wrk.from - 1; i <= wrk.to-1; i++ {
		file := wrk.ti.FileStats[i]
		if wrk.isCancelled {
			return
		}

		err := uploadFile(wrk, file, i+1, len(wrk.ti.FileStats))
		if err != nil {
			errstr := fmt.Sprintf(tr(wrk.c.Sender().ID, "upload_error"), err)
			wrk.c.Bot().Edit(wrk.msg, errstr)
			iserr = true
			break
		}
	}
	if !iserr {
		wrk.c.Bot().Delete(wrk.msg)
	}
}

func uploadFile(wrk *Worker, file *state.TorrentFileStat, fi, fc int) error {
	caption := filepath.Base(file.Path)
	torrFile, err := NewTorrFile(wrk, file)
	if err != nil {
		return err
	}

	var wa sync.WaitGroup
	wa.Add(1)
	complete := false
	go func() {
		for !complete {
			updateLoadStatus(wrk, torrFile, fi, fc)
			time.Sleep(1 * time.Second)
		}
		wa.Done()
	}()

	d := &tele.Document{}
	d.FileName = file.Path
	d.Caption = caption
	d.File.FileReader = torrFile

	for i := 0; i < 20; i++ {
		err = wrk.c.Send(d)
		if err == nil || errors.Is(err, ERR_STOPPED) {
			break
		}
		log.TLogln("tg upload retry", i+1, "/", 20)
		if i < 19 {
			backoff := time.Duration(1<<uint(i)) * 100 * time.Millisecond
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
			time.Sleep(backoff)
		}
	}

	complete = true
	wa.Wait()
	torrFile.Close()
	if errors.Is(err, ERR_STOPPED) {
		err = nil
	}
	return err
}
