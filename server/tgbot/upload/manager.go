package upload

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
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
	queueLock sync.Mutex
	stop      chan struct{}
	done      chan struct{}
	stopOnce  sync.Once
}

func (m *Manager) Start() {
	m.working = make(map[int]*Worker)
	m.stop = make(chan struct{})
	m.done = make(chan struct{})
	// Stop uses stopOnce; a second Start/Stop (bot restart, tests) must arm it again.
	m.stopOnce = sync.Once{}
	go func() {
		m.work()
		close(m.done)
	}()
}

func (m *Manager) Stop() {
	if m.stop == nil {
		return
	}
	m.stopOnce.Do(func() { close(m.stop) })
	if m.done != nil {
		<-m.done
	}
}

func Stop() { manager.Stop() }

func (m *Manager) AddRange(c tele.Context, hash string, from, to int) {
	m.queueLock.Lock()
	defer m.queueLock.Unlock()

	if len(m.queue) > 50 {
		_, _ = c.Bot().Send(c.Recipient(), fmt.Sprintf(tr(c.Sender().ID, "upload_queue_full"), len(m.queue)))
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
		_, _ = c.Bot().Edit(msg, tr(c.Sender().ID, "torrent_not_found")+":\n<code>"+hash+"</code>")
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
			_ = w.c.Bot().Delete(w.msg)
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
		select {
		case <-m.stop:
			return
		default:
		}
		m.queueLock.Lock()
		if len(m.working) > 0 {
			m.queueLock.Unlock()
			m.sendQueueStatus()
			if !m.sleep() {
				return
			}
			continue
		}
		if len(m.queue) == 0 {
			m.queueLock.Unlock()
			if !m.sleep() {
				return
			}
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

func (m *Manager) sleep() bool {
	select {
	case <-m.stop:
		return false
	case <-time.After(time.Second):
		return true
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

		_, _ = wrk.c.Bot().Edit(wrk.msg, msg, torrKbd)
	}
}

func loading(wrk *Worker) {
	iserr := false

	t := torr.GetTorrent(wrk.torrentHash)
	if t == nil {
		_, _ = wrk.c.Bot().Edit(wrk.msg, tr(wrk.c.Sender().ID, "torrent_not_found")+":\n<code>"+wrk.torrentHash+"</code>")
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
			_, _ = wrk.c.Bot().Edit(wrk.msg, errstr)
			iserr = true
			break
		}
	}
	if !iserr {
		_ = wrk.c.Bot().Delete(wrk.msg)
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
	var complete atomic.Bool
	go func() {
		for !complete.Load() {
			updateLoadStatus(wrk, torrFile, fi, fc)
			time.Sleep(1 * time.Second)
		}
		wa.Done()
	}()

	d := &tele.Document{}
	d.FileName = file.Path
	d.Caption = caption
	d.FileReader = torrFile

	for i := 0; i < 20; i++ {
		err = wrk.c.Send(d)
		if err == nil || errors.Is(err, ErrStopped) {
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

	complete.Store(true)
	wa.Wait()
	torrFile.Close()
	if errors.Is(err, ErrStopped) {
		err = nil
	}
	return err
}
