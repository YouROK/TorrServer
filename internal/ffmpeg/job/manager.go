package job

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"silo/internal/ffmpeg/args"
	"silo/internal/log"
)

const Tag = "[FFmpeg]"

var (
	ErrTooManySessions = errors.New("too many transcoding sessions")
	// ErrNotFound означает, что задание с таким идентификатором не запущено
	ErrNotFound = errors.New("transcoding session not found")
	// ErrPauseUnsupported означает, что сборка ffmpeg не умеет приостанавливать вывод
	ErrPauseUnsupported = errors.New("this ffmpeg build does not support pause")
	ErrNoBinary         = errors.New("ffmpeg binary is not configured")
	ErrShuttingDown     = errors.New("transcoding manager is shutting down")
)

// StartRequest описывает задание на запуск процесса.
type StartRequest struct {
	ID        string
	UserID    string
	Hash      string
	FileIdx   int
	ProfileID string

	// WorkDir и PlaylistPath заполняются для сегментированного вывода
	WorkDir      string
	PlaylistPath string

	Options  args.Options
	Duration time.Duration
}

// Config описывает параметры менеджера заданий.
type Config struct {
	Binary      string // Путь к ffmpeg
	Selector    args.EncoderSelector
	MaxSessions int // 0 - без ограничения
	StopTimeout time.Duration
	KeepStderr  bool
	// PauseSupport показывает, умеет ли сборка ставить вывод на паузу
	PauseSupport bool
	// ProgressPeriod - как часто публикуется прогресс
	ProgressPeriod time.Duration
	// KeepAhead - сколько сегментов разрешено держать невостребованными
	KeepAhead int
	OnEvent   func(topic string, payload any)
}

// Manager запускает и отслеживает процессы ffmpeg.
type Manager struct {
	cfg Config

	mu      sync.RWMutex
	jobs    map[string]*Job
	active  atomic.Int64
	stopped bool
	wg      sync.WaitGroup
}

// NewManager создаёт менеджер заданий.
func NewManager(cfg Config) *Manager {
	if cfg.StopTimeout <= 0 {
		cfg.StopTimeout = 5 * time.Second
	}
	return &Manager{
		cfg:  cfg,
		jobs: make(map[string]*Job),
	}
}

// Enabled сообщает, можно ли запускать процессы.
func (m *Manager) Enabled() bool { return m != nil && m.cfg.Binary != "" }

// Start запускает новый процесс ffmpeg.
func (m *Manager) Start(ctx context.Context, req StartRequest) (*Job, error) {
	if !m.Enabled() {
		return nil, ErrNoBinary
	}
	if req.Options.Input == "" {
		return nil, errors.New("input is required")
	}
	if req.ID == "" {
		return nil, errors.New("session id is required")
	}

	if err := m.reserve(); err != nil {
		return nil, err
	}

	a, err := args.Build(req.Options, m.cfg.Selector)
	if err != nil {
		m.releaseSlot()
		return nil, err
	}

	cmd := exec.Command(m.cfg.Binary, a...)
	setProcessGroup(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		m.releaseSlot()
		return nil, fmt.Errorf("failed to open stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.releaseSlot()
		return nil, fmt.Errorf("failed to open stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.releaseSlot()
		return nil, fmt.Errorf("failed to open stderr: %w", err)
	}

	j := newJob(cmd, req)
	j.setProcess(stdin, stdout)

	if err := cmd.Start(); err != nil {
		m.releaseSlot()
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	if !m.register(j) {
		_ = killProcess(cmd)
		m.releaseSlot()
		return nil, ErrShuttingDown
	}

	j.setState(StateRunning)
	j.setPauseSupport(m.cfg.PauseSupport)

	// Сегментированный вывод пишется на диск без обратного давления,
	// поэтому процесс придерживается, когда убегает вперёд плеера
	if j.Protocol == string(args.ProtocolHLS) && m.cfg.PauseSupport {
		j.throttler = newThrottler(j, m.cfg.KeepAhead)
		go j.throttler.run()
	}

	m.emit("transcode:session:started", j.Snapshot())

	m.wg.Add(1)
	go m.watch(ctx, j, stderr)

	log.Infof("%s Session %s started", Tag, j.ID)
	return j, nil
}

// watch читает вывод процесса, обновляет прогресс и фиксирует завершение.
func (m *Manager) watch(ctx context.Context, j *Job, stderr io.ReadCloser) {
	defer m.wg.Done()

	// Разрыв соединения клиента останавливает процесс
	watcherDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			j.Stop(m.cfg.StopTimeout)
		case <-watcherDone:
		}
	}()

	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		m.readProgress(j, stderr)
	}()

	err := j.cmd.Wait()
	close(watcherDone)
	<-stderrDone

	if j.throttler != nil {
		j.throttler.shutdown()
	}

	j.closePipes()
	m.unregister(j.ID)
	m.releaseSlot()

	canceled := j.State() == StateStopping || errors.Is(ctx.Err(), context.Canceled)
	exitCode := 0
	if j.cmd.ProcessState != nil {
		exitCode = j.cmd.ProcessState.ExitCode()
	}
	if ctx.Err() != nil && exitCode != 0 {
		canceled = true
	}
	j.markFinished(err, exitCode, canceled)

	snap := j.Snapshot()
	switch {
	case snap.State == StateCanceled:
		log.Infof("%s Session %s canceled", Tag, j.ID)
		m.emit("transcode:session:stopped", snap)
	case snap.State == StateFailed:
		log.Errorf("%s Session %s failed: %s", Tag, j.ID, j.failureReason())
		m.emit("transcode:session:failed", snap)
	default:
		log.Infof("%s Session %s finished", Tag, j.ID)
		m.emit("transcode:session:stopped", snap)
	}

	// Ожидание снимается после событий, чтобы подписчики увидели финальный снимок
	j.signalDone()
}

// readProgress разбирает вывод ffmpeg в машинном формате -progress.
func (m *Manager) readProgress(j *Job, stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	progress := Progress{}
	var lastEmit time.Time

	for scanner.Scan() {
		line := scanner.Text()

		if isProgressLine(line) {
			parseProgressLine(&progress, line)
			j.setProgress(progress)
			if time.Since(lastEmit) >= m.progressPeriod() {
				m.emit("transcode:session:progress", j.Snapshot())
				lastEmit = time.Now()
			}
			continue
		}

		j.stderr.Write([]byte(line))
		if m.cfg.KeepStderr {
			log.Debugf("%s Session %s: %s", Tag, j.ID, line)
		}
	}
}

// progressPeriod возвращает период публикации прогресса.
func (m *Manager) progressPeriod() time.Duration {
	if m.cfg.ProgressPeriod > 0 {
		return m.cfg.ProgressPeriod
	}
	return time.Second
}

// isProgressLine определяет строку машинного прогресса.
func isProgressLine(line string) bool {
	if line == "" || line[0] == '[' {
		return false
	}
	for _, prefix := range progressKeys {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

var progressKeys = []string{
	"frame=", "fps=", "bitrate=", "total_size=", "out_time_us=", "out_time_ms=",
	"out_time=", "dup_frames=", "drop_frames=", "speed=", "progress=", "stream_",
}

// SetEventHandler задаёт обработчик событий менеджера.
func (m *Manager) SetEventHandler(fn func(topic string, payload any)) {
	m.cfg.OnEvent = fn
}

// Stop останавливает задание по идентификатору.
func (m *Manager) Stop(id string) bool {
	j := m.Get(id)
	if j == nil {
		return false
	}
	j.Stop(m.cfg.StopTimeout)
	return true
}

// Get возвращает задание по идентификатору.
func (m *Manager) Get(id string) *Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobs[id]
}

// List возвращает снимки всех активных заданий.
// Pause приостанавливает вывод задания по идентификатору.
func (m *Manager) Pause(id string) error {
	j := m.Get(id)
	if j == nil {
		return ErrNotFound
	}
	return j.Pause()
}

// Resume возобновляет вывод задания по идентификатору.
func (m *Manager) Resume(id string) error {
	j := m.Get(id)
	if j == nil {
		return ErrNotFound
	}
	return j.Resume()
}

func (m *Manager) List() []Snapshot {
	m.mu.RLock()
	out := make([]Snapshot, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, j.Snapshot())
	}
	m.mu.RUnlock()
	return out
}

// Count возвращает число активных заданий.
func (m *Manager) Count() int { return int(m.active.Load()) }

// StopAll останавливает все активные задания.
func (m *Manager) StopAll() {
	m.mu.Lock()
	m.stopped = true
	jobs := make([]*Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		jobs = append(jobs, j)
	}
	m.mu.Unlock()

	for _, j := range jobs {
		j.Stop(m.cfg.StopTimeout)
	}
}

// Wait ожидает завершения всех отслеживающих горутин.
func (m *Manager) Wait() { m.wg.Wait() }

// reserve проверяет и занимает слот под новое задание.
func (m *Manager) reserve() error {
	m.mu.RLock()
	stopped := m.stopped
	m.mu.RUnlock()
	if stopped {
		return ErrShuttingDown
	}

	n := m.active.Add(1)
	if m.cfg.MaxSessions > 0 && n > int64(m.cfg.MaxSessions) {
		m.active.Add(-1)
		return fmt.Errorf("%w: limit is %d", ErrTooManySessions, m.cfg.MaxSessions)
	}
	return nil
}

// releaseSlot освобождает слот задания.
func (m *Manager) releaseSlot() {
	if n := m.active.Add(-1); n < 0 {
		m.active.Store(0)
	}
}

// register добавляет задание в реестр, если менеджер ещё работает.
func (m *Manager) register(j *Job) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return false
	}
	m.jobs[j.ID] = j
	return true
}

// unregister убирает задание из реестра.
func (m *Manager) unregister(id string) {
	m.mu.Lock()
	delete(m.jobs, id)
	m.mu.Unlock()
}

// emit отправляет событие во внешний обработчик.
func (m *Manager) emit(topic string, payload any) {
	if m.cfg.OnEvent != nil {
		m.cfg.OnEvent(topic, payload)
	}
}

// failureReason описывает причину отказа процесса.
func (j *Job) failureReason() string {
	if err := j.Err(); err != nil {
		return err.Error()
	}
	if lines := j.Stderr(); len(lines) > 0 {
		return lines[len(lines)-1]
	}
	return "process exited unexpectedly"
}
