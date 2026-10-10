package job

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"silo/internal/ffmpeg/args"
)

// helperEnv помечает процесс как подставной ffmpeg.
const helperEnv = "SILO_FAKE_FFMPEG"

func TestMain(m *testing.M) {
	if os.Getenv(helperEnv) != "" {
		fakeFFmpeg()
		return
	}
	os.Exit(m.Run())
}

// fakeFFmpeg подменяет реальный ffmpeg: пишет прогресс в stderr и данные в stdout.
func fakeFFmpeg() {
	mode := os.Getenv(helperEnv)

	// В режиме segments ввод читает сам обработчик, остальным нужен только q
	stop := make(chan struct{})
	if mode != "segments" {
		go func() {
			buf := make([]byte, 1)
			for {
				if _, err := os.Stdin.Read(buf); err != nil {
					return
				}
				if buf[0] == 'q' {
					close(stop)
					return
				}
			}
		}()
	}

	switch mode {
	case "fail":
		fmt.Fprintln(os.Stderr, "[error] fake ffmpeg failed")
		os.Exit(2)

	case "noexit":
		// Не реагирует на q, проверяет принудительное завершение
		time.Sleep(30 * time.Second)
		os.Exit(0)

	case "segments":
		// Пишет сегменты и слушает клавиши паузы, как настоящий ffmpeg
		fakeSegments(stop)
		return
	}

	payload := make([]byte, 1024)
	for i := 0; i < 5; i++ {
		fmt.Fprintf(os.Stderr, "frame=%d\n", i*25)
		fmt.Fprintf(os.Stderr, "fps=25.0\n")
		fmt.Fprintf(os.Stderr, "bitrate=1000.5kbits/s\n")
		fmt.Fprintf(os.Stderr, "total_size=%d\n", (i+1)*1024)
		fmt.Fprintf(os.Stderr, "out_time_us=%d\n", (i+1)*1000000)
		fmt.Fprintf(os.Stderr, "speed=1.5x\n")
		fmt.Fprintf(os.Stderr, "progress=continue\n")
		fmt.Fprintln(os.Stderr, "[info] fake ffmpeg working")

		if _, err := os.Stdout.Write(payload); err != nil {
			os.Exit(0)
		}

		select {
		case <-stop:
			fmt.Fprintln(os.Stderr, "progress=end")
			os.Exit(0)
		case <-time.After(200 * time.Millisecond):
		}
	}

	fmt.Fprintln(os.Stderr, "progress=end")
	os.Exit(0)
}

// fakeSegments создаёт сегменты в каталоге задания и реагирует на паузу.
func fakeSegments(stop <-chan struct{}) {
	dir := os.Getenv("SILO_FAKE_FFMPEG_DIR")
	if dir == "" {
		fmt.Fprintln(os.Stderr, "[error] SILO_FAKE_FFMPEG_DIR is not set")
		os.Exit(1)
	}

	paused := make(chan struct{}, 1)
	keys := make(chan byte, 16)

	go func() {
		buf := make([]byte, 1)
		for {
			if _, err := os.Stdin.Read(buf); err != nil {
				return
			}
			select {
			case keys <- buf[0]:
			default:
			}
		}
	}()

	for i := 0; ; i++ {
		// Проверяем команды паузы и завершения
		select {
		case key := <-keys:
			switch key {
			case 'q':
				fmt.Fprintln(os.Stderr, "progress=end")
				os.Exit(0)
			case 'p':
				paused <- struct{}{}
			case 'u':
				select {
				case <-paused:
				default:
				}
			}
		case <-stop:
			os.Exit(0)
		default:
		}

		// На паузе сегменты не создаются, как и у настоящего ffmpeg
		select {
		case <-paused:
			paused <- struct{}{}
			time.Sleep(50 * time.Millisecond)
			continue
		default:
		}

		name := fmt.Sprintf("seg%d.ts", i)
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, 1024), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "[error] %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "frame=%d\n", i*25)
		fmt.Fprintf(os.Stderr, "out_time_us=%d\n", i*1000000)
		fmt.Fprintln(os.Stderr, "progress=continue")
		time.Sleep(30 * time.Millisecond)
	}
}

// helperBinary возвращает путь к текущему тестовому бинарю.
func helperBinary(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("failed to locate test binary: %v", err)
	}
	return exe
}

// fakeSelector подменяет список доступных кодировщиков.
type fakeSelector struct{}

func (fakeSelector) AutoVideoEncoder(codec, hwaccel string) string {
	if codec == "h264" {
		return "libx264"
	}
	return ""
}

func (fakeSelector) AutoAudioEncoder(codec string) string {
	if codec == "aac" {
		return "aac"
	}
	return ""
}

// newTestManager создаёт менеджер с подставным ffmpeg.
func newTestManager(t *testing.T, mode string, cfg Config) *Manager {
	t.Helper()

	if cfg.Binary == "" {
		cfg.Binary = helperBinary(t)
	}
	if cfg.Selector == nil {
		cfg.Selector = fakeSelector{}
	}
	if cfg.StopTimeout == 0 {
		cfg.StopTimeout = 2 * time.Second
	}

	m := NewManager(cfg)
	// Подставной процесс запускается тем же бинарём с переменной окружения
	t.Setenv(helperEnv, mode)
	return m
}

// eventRecorder собирает события менеджера, безопасно для параллельных вызовов.
type eventRecorder struct {
	mu     sync.Mutex
	topics []string
	snaps  map[string]Snapshot
}

func newEventRecorder() *eventRecorder {
	return &eventRecorder{snaps: make(map[string]Snapshot)}
}

func (r *eventRecorder) handler(topic string, payload any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.topics = append(r.topics, topic)
	if snap, ok := payload.(Snapshot); ok {
		r.snaps[topic] = snap
	}
}

func (r *eventRecorder) has(topic string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return contains(r.topics, topic)
}

func (r *eventRecorder) snapshot(topic string) (Snapshot, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	snap, ok := r.snaps[topic]
	return snap, ok
}

func (r *eventRecorder) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.topics))
	copy(out, r.topics)
	return out
}

// startRequest собирает простое задание для тестов.
func startRequest(id string) StartRequest {
	o := args.DefaultOptions("fake-input.mkv")
	o.Output = "pipe:1"

	return StartRequest{
		ID:       id,
		UserID:   "owner",
		Hash:     "0123456789abcdef0123456789abcdef01234567",
		FileIdx:  0,
		Options:  o,
		Duration: 5 * time.Second,
	}
}

func TestManagerStartAndFinish(t *testing.T) {
	rec := newEventRecorder()
	m := newTestManager(t, "ok", Config{OnEvent: rec.handler})

	j, err := m.Start(context.Background(), startRequest("job-1"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if j.State() != StateRunning {
		t.Errorf("state = %s, want %s", j.State(), StateRunning)
	}

	select {
	case <-j.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("job did not finish in time")
	}

	if got := j.State(); got != StateDone {
		t.Errorf("final state = %s, want %s", got, StateDone)
	}
	if m.Count() != 0 {
		t.Errorf("active jobs = %d, want 0", m.Count())
	}

	snap := j.Snapshot()
	if snap.Position < 5*time.Second {
		t.Errorf("position = %v, want at least 5s", snap.Position)
	}
	if snap.Percent != 100 {
		t.Errorf("percent = %v, want 100", snap.Percent)
	}
	if snap.FPS != 25 {
		t.Errorf("fps = %v, want 25", snap.FPS)
	}
	if snap.OutputSize == 0 {
		t.Error("output size was not tracked")
	}

	if !rec.has("transcode:session:started") {
		t.Errorf("started event missing: %v", rec.all())
	}
	if !rec.has("transcode:session:progress") {
		t.Errorf("progress event missing: %v", rec.all())
	}
	if !rec.has("transcode:session:stopped") {
		t.Errorf("stopped event missing: %v", rec.all())
	}
	if final, ok := rec.snapshot("transcode:session:stopped"); ok && final.State != StateDone {
		t.Errorf("final event state = %s, want %s", final.State, StateDone)
	}
}

func TestManagerOutputIsReadable(t *testing.T) {
	m := newTestManager(t, "ok", Config{})

	j, err := m.Start(context.Background(), startRequest("job-output"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	buf := make([]byte, 4096)
	total := 0
	deadline := time.Now().Add(10 * time.Second)
	for total < 2048 && time.Now().Before(deadline) {
		n, err := j.Output().Read(buf)
		total += n
		if err != nil {
			break
		}
	}

	if total < 2048 {
		t.Errorf("read %d bytes from the output, want at least 2048", total)
	}

	select {
	case <-j.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("job did not finish in time")
	}
}

func TestManagerFailure(t *testing.T) {
	rec := newEventRecorder()
	m := newTestManager(t, "fail", Config{OnEvent: rec.handler})

	j, err := m.Start(context.Background(), startRequest("job-fail"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("job did not finish in time")
	}

	if j.State() != StateFailed {
		t.Errorf("state = %s, want %s", j.State(), StateFailed)
	}

	failed, ok := rec.snapshot("transcode:session:failed")
	if !ok {
		t.Fatalf("failed event was not emitted: %v", rec.all())
	}
	if failed.ExitCode != 2 {
		t.Errorf("exit code = %d, want 2", failed.ExitCode)
	}
	if failed.Error == "" {
		t.Error("failure reason is empty")
	}
}

func TestManagerStop(t *testing.T) {
	m := newTestManager(t, "noexit", Config{StopTimeout: 500 * time.Millisecond})

	j, err := m.Start(context.Background(), startRequest("job-stop"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	start := time.Now()
	if !m.Stop("job-stop") {
		t.Fatal("Stop returned false for an active job")
	}

	select {
	case <-j.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("job did not stop in time")
	}

	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("stop took %v, expected a forced kill", elapsed)
	}
	if j.State() != StateCanceled {
		t.Errorf("state = %s, want %s", j.State(), StateCanceled)
	}
}

func TestManagerStopUnknown(t *testing.T) {
	m := newTestManager(t, "ok", Config{})
	if m.Stop("missing") {
		t.Error("Stop returned true for an unknown job")
	}
}

func TestManagerMaxSessions(t *testing.T) {
	m := newTestManager(t, "noexit", Config{MaxSessions: 2, StopTimeout: 300 * time.Millisecond})
	defer m.StopAll()

	var jobs []*Job
	for i := 0; i < 2; i++ {
		j, err := m.Start(context.Background(), startRequest(fmt.Sprintf("job-limit-%d", i)))
		if err != nil {
			t.Fatalf("Start %d: %v", i, err)
		}
		jobs = append(jobs, j)
	}

	_, err := m.Start(context.Background(), startRequest("job-over-limit"))
	if !errors.Is(err, ErrTooManySessions) {
		t.Fatalf("error = %v, want %v", err, ErrTooManySessions)
	}

	for _, j := range jobs {
		j.Stop(time.Second)
		<-j.Done()
	}

	// После освобождения слота запуск должен снова проходить
	j, err := m.Start(context.Background(), startRequest("job-after-limit"))
	if err != nil {
		t.Fatalf("Start after release: %v", err)
	}
	j.Stop(time.Second)
	<-j.Done()
}

func TestManagerContextCancelStopsJob(t *testing.T) {
	m := newTestManager(t, "noexit", Config{StopTimeout: 300 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	j, err := m.Start(ctx, startRequest("job-ctx"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-j.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("job was not stopped by context cancellation")
	}
}

func TestManagerRejectsNoBinary(t *testing.T) {
	m := NewManager(Config{})
	if m.Enabled() {
		t.Error("manager without a binary must be disabled")
	}
	if _, err := m.Start(context.Background(), startRequest("x")); !errors.Is(err, ErrNoBinary) {
		t.Errorf("error = %v, want %v", err, ErrNoBinary)
	}
}

func TestManagerRejectsEmptyInput(t *testing.T) {
	m := newTestManager(t, "ok", Config{})

	req := startRequest("job-empty")
	req.Options.Input = ""
	if _, err := m.Start(context.Background(), req); err == nil {
		t.Error("expected an error for an empty input")
	}
	if m.Count() != 0 {
		t.Errorf("active jobs = %d, want 0 after a failed start", m.Count())
	}
}

func TestManagerList(t *testing.T) {
	m := newTestManager(t, "noexit", Config{StopTimeout: 300 * time.Millisecond})
	defer m.StopAll()

	j, err := m.Start(context.Background(), startRequest("job-list"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		j.Stop(time.Second)
		<-j.Done()
	}()

	list := m.List()
	if len(list) != 1 {
		t.Fatalf("list length = %d, want 1", len(list))
	}
	if list[0].ID != "job-list" {
		t.Errorf("job id = %q, want job-list", list[0].ID)
	}
	if list[0].UserID != "owner" {
		t.Errorf("user id = %q, want owner", list[0].UserID)
	}
	if list[0].PID == 0 {
		t.Error("pid was not reported")
	}
}

func TestManagerStopAll(t *testing.T) {
	m := newTestManager(t, "noexit", Config{StopTimeout: 300 * time.Millisecond})

	var jobs []*Job
	for i := 0; i < 3; i++ {
		j, err := m.Start(context.Background(), startRequest(fmt.Sprintf("job-all-%d", i)))
		if err != nil {
			t.Fatalf("Start %d: %v", i, err)
		}
		jobs = append(jobs, j)
	}

	m.StopAll()

	for _, j := range jobs {
		select {
		case <-j.Done():
		case <-time.After(10 * time.Second):
			t.Fatalf("job %s did not stop", j.ID)
		}
	}

	if m.Count() != 0 {
		t.Errorf("active jobs = %d, want 0", m.Count())
	}

	if _, err := m.Start(context.Background(), startRequest("job-after-stopall")); !errors.Is(err, ErrShuttingDown) {
		t.Errorf("error = %v, want %v", err, ErrShuttingDown)
	}
}

func TestManagerNoZombieProcesses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process check in short mode")
	}

	m := newTestManager(t, "noexit", Config{StopTimeout: 300 * time.Millisecond})

	j, err := m.Start(context.Background(), startRequest("job-zombie"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	pid := j.Snapshot().PID

	j.Stop(time.Second)
	<-j.Done()

	// Процесс должен исчезнуть из таблицы процессов
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Errorf("process %d is still alive after Stop", pid)
}

// processExists проверяет, существует ли процесс с указанным идентификатором.
func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid))
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// contains проверяет наличие строки в списке.
func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
