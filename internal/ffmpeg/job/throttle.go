package job

import (
	"os"
	"path/filepath"
	"strconv"
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

// check сравнивает готовые сегменты с востребованными плеером.
// Отсчёт ведётся от номера первого сегмента запуска: после перемотки
// в каталоге остаются файлы прошлых позиций, и по ним опережение
// считалось бы мгновенно, останавливая процесс на новом месте.
func (t *throttler) check() {
	start := t.job.StartIndex()
	produced := countSegmentsFrom(t.job.WorkDir, start)
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

// countSegmentsFrom считает сегменты начиная с указанного номера.
func countSegmentsFrom(dir string, from int) int {
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
			if num, ok := trailingSegmentNumber(e.Name()); ok && num >= from {
				count++
			}
		}
	}
	return count
}

// trailingSegmentNumber извлекает номер из имени сегмента.
func trailingSegmentNumber(name string) (int, bool) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	digits := ""
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] < '0' || base[i] > '9' {
			break
		}
		digits = string(base[i]) + digits
	}
	if digits == "" {
		return 0, false
	}
	num, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return num, true
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
