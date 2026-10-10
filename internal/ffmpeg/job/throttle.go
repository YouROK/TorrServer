package job

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Протоколы, для которых применяется ограничение скорости кодирования.
const (
	// defaultThrottleInterval - период проверки отставания
	defaultThrottleInterval = 5 * time.Second
	// defaultKeepAhead - сколько сегментов разрешено держать невостребованными
	defaultKeepAhead = 4
)

// throttler приостанавливает ffmpeg, когда он убегает вперёд плеера.
// Для сегментированного вывода обратного давления нет: процесс пишет файлы
// на диск с максимальной скоростью, поэтому его приходится останавливать вручную.
type throttler struct {
	job       *Job
	interval  time.Duration
	keepAhead int

	stopOnce sync.Once
	stop     chan struct{}
	done     chan struct{}
}

// newThrottler создаёт ограничитель для задания.
func newThrottler(j *Job, keepAhead int) *throttler {
	if keepAhead <= 0 {
		keepAhead = defaultKeepAhead
	}
	return &throttler{
		job:       j,
		interval:  defaultThrottleInterval,
		keepAhead: keepAhead,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}
}

// run проверяет отставание плеера и управляет паузой процесса.
func (t *throttler) run() {
	defer close(t.done)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stop:
			return
		case <-t.job.Done():
			return
		case <-ticker.C:
			t.check()
		}
	}
}

// check сравнивает число готовых сегментов с востребованными плеером.
func (t *throttler) check() {
	produced := countSegments(t.job.WorkDir)
	served := int(t.job.SegmentsServed())
	ahead := produced - served

	switch {
	case ahead >= t.keepAhead && !t.job.Paused():
		if err := t.job.Pause(); err == nil {
			t.job.setThrottled(true)
		}
	case ahead <= t.keepAhead/2 && t.job.Paused():
		if err := t.job.Resume(); err == nil {
			t.job.setThrottled(false)
		}
	}
}

// shutdown останавливает ограничитель, возобновляя процесс.
func (t *throttler) shutdown() {
	t.stopOnce.Do(func() { close(t.stop) })
	<-t.done

	if t.job.Paused() {
		_ = t.job.Resume()
		t.job.setThrottled(false)
	}
}

// countSegments считает готовые сегменты в каталоге задания.
func countSegments(dir string) int {
	if dir == "" {
		return 0
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".ts", ".m4s":
			count++
		}
	}
	return count
}
