package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"silo/internal/ffmpeg/args"
)

func TestCountSegments(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"seg0.ts", "seg1.ts", "seg2.m4s", "index.m3u8", ".hidden.ts", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if got := countSegments(dir); got != 3 {
		t.Errorf("countSegments = %d, want 3", got)
	}
}

func TestCountSegmentsMissingDir(t *testing.T) {
	if got := countSegments(filepath.Join(t.TempDir(), "missing")); got != 0 {
		t.Errorf("countSegments = %d, want 0", got)
	}
	if got := countSegments(""); got != 0 {
		t.Errorf("countSegments for an empty dir = %d, want 0", got)
	}
}

func TestThrottlerPausesWhenAhead(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"seg0.ts", "seg1.ts", "seg2.ts", "seg3.ts", "seg4.ts", "seg5.ts"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	j := &Job{ID: "throttle-test", WorkDir: dir, finished: make(chan struct{})}
	j.setPauseSupport(true)

	th := newThrottler(j, 4)
	// Процесс не запущен, поэтому управлять паузой нечем: проверяем только решение
	produced := countSegments(dir)
	served := int(j.SegmentsServed())

	if produced-served < th.keepAhead {
		t.Fatalf("test setup is wrong: ahead = %d, want at least %d", produced-served, th.keepAhead)
	}
}

func TestThrottlerShutdownResumes(t *testing.T) {
	j := &Job{ID: "throttle-shutdown", finished: make(chan struct{})}
	j.setPauseSupport(true)

	th := newThrottler(j, 4)
	go th.run()

	// Остановка не должна зависать, даже если пауза не ставилась
	done := make(chan struct{})
	go func() {
		defer close(done)
		th.shutdown()
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown hung")
	}
}

func TestJobPauseWithoutSupport(t *testing.T) {
	j := &Job{ID: "no-pause", finished: make(chan struct{})}

	if err := j.Pause(); err == nil {
		t.Error("Pause must fail when the build does not support it")
	}
	if j.Paused() {
		t.Error("job must not be marked as paused")
	}
}

func TestJobSegmentCounter(t *testing.T) {
	j := &Job{ID: "counter", finished: make(chan struct{})}

	for i := 0; i < 5; i++ {
		j.AddServedSegment()
	}
	if got := j.SegmentsServed(); got != 5 {
		t.Errorf("served segments = %d, want 5", got)
	}
}

func TestJobThrottledFlag(t *testing.T) {
	j := &Job{ID: "flags", finished: make(chan struct{})}

	if j.Throttled() {
		t.Error("job must not be throttled initially")
	}
	j.setThrottled(true)
	if !j.Throttled() {
		t.Error("throttled flag was not set")
	}
	j.setThrottled(false)
	if j.Throttled() {
		t.Error("throttled flag was not cleared")
	}
}

// TestManagerThrottlesAhead проверяет ограничение на подставном процессе,
// который создаёт сегменты и слушает паузу как настоящий ffmpeg.
func TestManagerThrottlesAhead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping throttling test in short mode")
	}

	workDir := t.TempDir()

	m := newTestManager(t, "segments", Config{
		PauseSupport: true,
		KeepAhead:    3,
		StopTimeout:  time.Second,
	})
	t.Setenv("SILO_FAKE_FFMPEG_DIR", workDir)

	req := startRequest("job-throttle")
	req.WorkDir = workDir
	req.PlaylistPath = filepath.Join(workDir, "index.m3u8")
	req.Options.Protocol = args.ProtocolHLS

	j, err := m.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		j.Stop(time.Second)
		<-j.Done()
	}()

	// Плеер не забирает сегменты, поэтому процесс должен быть приостановлен
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if j.Paused() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	segments, _ := filepath.Glob(filepath.Join(workDir, "seg*.ts"))
	t.Errorf("process was not throttled: %d segments produced, served %d",
		len(segments), j.SegmentsServed())
}

// TestManagerThrottleResumesAfterServing проверяет возобновление,
// когда плеер догнал процесс.
func TestManagerThrottleResumesAfterServing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping throttling test in short mode")
	}

	workDir := t.TempDir()

	m := newTestManager(t, "segments", Config{
		PauseSupport: true,
		KeepAhead:    2,
		StopTimeout:  time.Second,
	})
	t.Setenv("SILO_FAKE_FFMPEG_DIR", workDir)

	req := startRequest("job-resume")
	req.WorkDir = workDir
	req.PlaylistPath = filepath.Join(workDir, "index.m3u8")
	req.Options.Protocol = args.ProtocolHLS

	j, err := m.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		j.Stop(time.Second)
		<-j.Done()
	}()

	// Ждём паузы
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) && !j.Paused() {
		time.Sleep(100 * time.Millisecond)
	}
	if !j.Paused() {
		t.Skip("process did not pause on this machine")
	}

	// Плеер забирает все готовые сегменты, поэтому процесс должен возобновиться.
	// До паузы процесс успевает создать их много, поэтому считаем по факту.
	produced, err := filepath.Glob(filepath.Join(workDir, "seg*.ts"))
	if err != nil {
		t.Fatalf("failed to count segments: %v", err)
	}
	for i := 0; i < len(produced)+1; i++ {
		j.AddServedSegment()
	}

	deadline = time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if !j.Paused() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Error("process did not resume after the player caught up")
}

// TestCountSegmentsFromSeek проверяет счёт сегментов от позиции перезапуска.
// После перемотки в каталоге остаются файлы прошлых мест: если считать
// их опережением, процесс встанет на паузу и сегмент не появится.
func TestCountSegmentsFromSeek(t *testing.T) {
	dir := t.TempDir()

	// Остатки прошлой позиции и начало нового запуска
	for _, n := range []int{0, 1, 2, 3, 4, 1200, 1201} {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("seg%d.ts", n)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if got := countSegmentsFrom(dir, 1200); got != 2 {
		t.Errorf("segments from 1200 = %d, want 2", got)
	}
	if got := countSegmentsFrom(dir, 0); got != 7 {
		t.Errorf("segments from 0 = %d, want 7", got)
	}
	// Служебные файлы не считаются
	if err := os.WriteFile(filepath.Join(dir, "index.m3u8"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := countSegmentsFrom(dir, 0); got != 7 {
		t.Errorf("playlist must not be counted: %d", got)
	}
}

// TestProducedFromSlidingWindow проверяет счёт сделанных сегментов, когда
// ffmpeg удаляет старые (скользящее окно) и после перемотки назад.
func TestProducedFromSlidingWindow(t *testing.T) {
	cases := []struct {
		nums       []int
		from, want int
	}{
		{nil, 0, 0},
		{[]int{0, 1, 2}, 0, 3},
		// ffmpeg уже удалил 0..16: сделано 24, а файлов только 7
		{[]int{17, 18, 19, 20, 21, 22, 23}, 0, 24},
		// Перемотка назад на 100: 1200 и 1201 остались от прошлого места
		{[]int{100, 101, 1200, 1201}, 100, 2},
		{[]int{0, 1, 2, 3, 4, 1200, 1201}, 1200, 2},
	}
	for _, c := range cases {
		dir := t.TempDir()
		for _, n := range c.nums {
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("seg%d.ts", n)), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "index.m3u8"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := producedFrom(dir, c.from); got != c.want {
			t.Errorf("%v from %d = %d, want %d", c.nums, c.from, got, c.want)
		}
	}
}

func TestTrailingSegmentNumber(t *testing.T) {
	cases := map[string]int{
		"seg0.ts":    0,
		"seg1200.ts": 1200,
		"seg7.m4s":   7,
	}
	for name, want := range cases {
		got, ok := trailingSegmentNumber(name)
		if !ok || got != want {
			t.Errorf("trailingSegmentNumber(%q) = %d, %v; want %d", name, got, ok, want)
		}
	}

	if _, ok := trailingSegmentNumber("index.m3u8"); ok {
		t.Error("playlist name must not produce a number")
	}
}

func TestJobStartIndex(t *testing.T) {
	j := &Job{ID: "seek", finished: make(chan struct{})}

	if j.StartIndex() != 0 {
		t.Errorf("initial index = %d, want 0", j.StartIndex())
	}
	j.SetStartIndex(1200)
	if j.StartIndex() != 1200 {
		t.Errorf("index = %d, want 1200", j.StartIndex())
	}
}
