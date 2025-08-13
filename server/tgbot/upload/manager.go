package upload

import (
	"errors"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"

	"server/torr"
	"server/torr/state"
)

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
		c.Bot().Send(c.Recipient(), "Очередь переполнена, попробуйте попозже\n\nЭлементов в очереди:"+strconv.Itoa(len(m.queue)))
		return
	}

	m.ids++
	if m.ids > math.MaxInt {
		m.ids = 0
	}

	var msg *tele.Message
	var err error

	for i := 0; i < 20; i++ {
		msg, err = c.Bot().Send(c.Recipient(), "<b>Подключение к торренту</b>\n<code>"+hash+"</code>")
		if err == nil {
			break
		} else {
			log.Println("Error send msg, try again:", i+1, "/", 20)
		}
	}

	if err != nil {
		log.Println("Error send msg:", err)
		return
	}

	t := torr.GetTorrent(hash)
	t.WaitInfo()
	for t.Status().Stat != state.TorrentWorking {
		time.Sleep(time.Second)
		t = torr.GetTorrent(hash)
	}
	ti := t.Status()

	if from == 1 && to == -1 {
		to = len(ti.FileStats)
	}
	if to > len(ti.FileStats) {
		to = len(ti.FileStats)
	}
	if from < 1 {
		from = 1
	}
	if to >= 0 && to < from {
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
	var rem []int
	for i, w := range m.queue {
		if w.id == id {
			w.isCancelled = true
			w.c.Bot().Delete(w.msg)
			rem = append(rem, i)
			return
		}
	}
	for _, i := range rem {
		m.queue = append(m.queue[:i], m.queue[i+1:]...)
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
		if wrk.msg == nil {
			continue
		}
		torrKbd := &tele.ReplyMarkup{}
		torrKbd.Inline([]tele.Row{torrKbd.Row(torrKbd.Data("Отмена", "cancel", strconv.Itoa(wrk.id)))}...)

		msg := "Номер в очереди " + strconv.Itoa(i+1)

		wrk.c.Bot().Edit(wrk.msg, msg, torrKbd)
	}
}

func loading(wrk *Worker) {
	iserr := false

	t := torr.GetTorrent(wrk.torrentHash)
	t.WaitInfo()
	for t.Status().Stat != state.TorrentWorking {
		time.Sleep(time.Second)
		t = torr.GetTorrent(wrk.torrentHash)
	}
	wrk.ti = t.Status()

	for i := wrk.from - 1; i <= wrk.to-1; i++ {
		file := wrk.ti.FileStats[i]
		if wrk.isCancelled {
			return
		}

		err := uploadFile(wrk, file, i+1, len(wrk.ti.FileStats))
		if err != nil {
			errstr := fmt.Sprintf("Ошибка загрузки в телеграм: %v", err.Error())
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
		} else {
			log.Println("Error send msg, try again:", i+1, "/", 20)
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
