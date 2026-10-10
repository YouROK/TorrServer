package job

import (
	"errors"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// State описывает состояние процесса транскодирования.
type State string

const (
	StateStarting State = "starting"
	StateRunning  State = "running"
	StateStopping State = "stopping"
	StateDone     State = "done"
	StateFailed   State = "failed"
	StateCanceled State = "canceled"
)

// Snapshot - снимок состояния задания для внешнего кода.
type Snapshot struct {
	ID         string        `json:"id"`
	UserID     string        `json:"user_id,omitempty"`
	Hash       string        `json:"hash,omitempty"`
	FileIdx    int           `json:"file_idx"`
	ProfileID  string        `json:"profile_id,omitempty"`
	Protocol   string        `json:"protocol"`
	State      State         `json:"state"`
	PID        int           `json:"pid"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at,omitempty"`
	Duration   time.Duration `json:"duration"`
	Position   time.Duration `json:"position"`
	Percent    float64       `json:"percent"`
	FPS        float64       `json:"fps"`
	BitrateBps int64         `json:"bitrate_bps"`
	OutputSize int64         `json:"output_size"`
	Speed      float64       `json:"speed"`
	ExitCode   int           `json:"exit_code"`
	Error      string        `json:"error,omitempty"`
	Command    string        `json:"command,omitempty"`
	// Paused показывает, что вывод приостановлен
	Paused bool `json:"paused"`
	// Throttled показывает, что паузу поставил ограничитель скорости
	Throttled bool `json:"throttled"`
	// PauseSupport показывает, умеет ли сборка приостанавливать вывод
	PauseSupport bool `json:"pause_support"`
}

// Job описывает один запущенный процесс ffmpeg.
type Job struct {
	ID        string
	UserID    string
	Hash      string
	FileIdx   int
	ProfileID string
	Protocol  string

	// WorkDir - каталог с файлами задания, PlaylistPath - плейлист HLS
	WorkDir      string
	PlaylistPath string

	Command   []string
	StartedAt time.Time

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stopMux sync.Mutex

	// finished закрывается после завершения процесса
	finished chan struct{}
	// stopOnce гарантирует однократную остановку
	stopOnce sync.Once
	// doneOnce гарантирует однократное закрытие канала завершения
	doneOnce sync.Once

	mu         sync.RWMutex
	state      State
	progress   Progress
	duration   time.Duration
	lastErr    error
	exitCode   int
	finishedAt time.Time

	// stderr хранит последние строки вывода ffmpeg для диагностики
	stderr *ringBuffer

	// paused показывает, приостановлен ли процесс клавишей p
	paused atomic.Bool
	// throttled показывает, что пауза поставлена ограничителем, а не плеером
	throttled atomic.Bool
	// segmentsServed считает сегменты, запрошенные плеером
	segmentsServed atomic.Int64

	// pauseOK хранит поддержку паузы сборкой ffmpeg
	pauseOK atomic.Bool

	// throttler следит за отставанием плеера при сегментированном выводе
	throttler *throttler
}

// newJob создаёт задание для запущенного процесса.
func newJob(cmd *exec.Cmd, opts StartRequest) *Job {
	return &Job{
		ID:           opts.ID,
		UserID:       opts.UserID,
		Hash:         opts.Hash,
		FileIdx:      opts.FileIdx,
		ProfileID:    opts.ProfileID,
		Protocol:     string(opts.Options.Protocol),
		WorkDir:      opts.WorkDir,
		PlaylistPath: opts.PlaylistPath,
		Command:      append([]string(nil), cmd.Args...),
		StartedAt:    time.Now(),
		cmd:          cmd,
		finished:     make(chan struct{}),
		state:        StateStarting,
		duration:     opts.Duration,
		stderr:       newRingBuffer(defaultStderrLines),
	}
}

// closePipes закрывает каналы ввода и вывода процесса.
func (j *Job) closePipes() {
	j.stopMux.Lock()
	defer j.stopMux.Unlock()
	if j.stdin != nil {
		_ = j.stdin.Close()
	}
	if j.stdout != nil {
		_ = j.stdout.Close()
	}
}

// setProcess сохраняет каналы процесса, полученные после запуска.
func (j *Job) setProcess(stdin io.WriteCloser, stdout io.ReadCloser) {
	j.stopMux.Lock()
	j.stdin = stdin
	j.stdout = stdout
	j.stopMux.Unlock()
}

// Output возвращает поток данных от ffmpeg.
func (j *Job) Output() io.ReadCloser {
	j.stopMux.Lock()
	defer j.stopMux.Unlock()
	return j.stdout
}

// State возвращает текущее состояние задания.
func (j *Job) State() State {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.state
}

// setState переводит задание в новое состояние.
func (j *Job) setState(s State) {
	j.mu.Lock()
	j.state = s
	j.mu.Unlock()
}

// setProgress сохраняет разобранный прогресс.
func (j *Job) setProgress(p Progress) {
	j.mu.Lock()
	j.progress = p
	j.mu.Unlock()
}

// Done возвращает канал, закрывающийся при завершении процесса.
func (j *Job) Done() <-chan struct{} { return j.finished }

// Err возвращает ошибку завершения, если она была.
func (j *Job) Err() error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.lastErr
}

// ExitCode возвращает код возврата процесса.
func (j *Job) ExitCode() int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.exitCode
}

// Stderr возвращает последние строки вывода ffmpeg.
func (j *Job) Stderr() []string {
	return j.stderr.Lines()
}

// Snapshot собирает снимок состояния задания.
func (j *Job) Snapshot() Snapshot {
	j.mu.RLock()
	state := j.state
	progress := j.progress
	lastErr := j.lastErr
	exitCode := j.exitCode
	finishedAt := j.finishedAt
	j.mu.RUnlock()

	cmd := ""
	if len(j.Command) > 0 {
		cmd = j.Command[0]
	}
	if len(j.Command) > 1 {
		cmd = joinCommand(j.Command[1:])
	}

	snap := Snapshot{
		Paused:       j.Paused(),
		Throttled:    j.Throttled(),
		PauseSupport: j.pauseSupported(),
		ID:           j.ID,
		UserID:       j.UserID,
		Hash:         j.Hash,
		FileIdx:      j.FileIdx,
		ProfileID:    j.ProfileID,
		Protocol:     j.Protocol,
		State:        state,
		PID:          j.pid(),
		StartedAt:    j.StartedAt,
		FinishedAt:   finishedAt,
		Position:     progress.OutTime,
		Percent:      progress.Percent(j.duration),
		FPS:          progress.FPS,
		BitrateBps:   progress.BitrateBps,
		OutputSize:   progress.TotalSize,
		Speed:        progress.Speed,
		ExitCode:     exitCode,
		Command:      cmd,
	}
	if lastErr != nil {
		snap.Error = lastErr.Error()
	}
	return snap
}

// pid возвращает идентификатор процесса.
func (j *Job) pid() int {
	j.stopMux.Lock()
	defer j.stopMux.Unlock()
	if j.cmd == nil || j.cmd.Process == nil {
		return 0
	}
	return j.cmd.Process.Pid
}

// Stop мягко завершает ffmpeg: просит выйти и убивает процесс при таймауте.
func (j *Job) Stop(timeout time.Duration) {
	j.stopOnce.Do(func() {
		j.setState(StateStopping)

		j.stopMux.Lock()
		stdin := j.stdin
		j.stopMux.Unlock()

		if stdin != nil {
			_, _ = io.WriteString(stdin, "q")
		}

		select {
		case <-j.finished:
			return
		case <-time.After(timeout):
		}

		j.stopMux.Lock()
		cmd := j.cmd
		j.stopMux.Unlock()
		if cmd != nil {
			_ = killProcess(cmd)
		}
	})
}

// Pause приостанавливает вывод ffmpeg клавишей p.
// Поддержка паузы зависит от сборки, поэтому вызывающий код проверяет её заранее.
func (j *Job) Pause() error {
	if !j.pauseSupported() {
		return errors.New("ffmpeg build does not support pause")
	}
	if j.paused.Load() {
		return nil
	}

	j.stopMux.Lock()
	stdin := j.stdin
	j.stopMux.Unlock()

	if stdin == nil {
		return errors.New("process input is not available")
	}
	if _, err := io.WriteString(stdin, "p"); err != nil {
		return err
	}

	j.paused.Store(true)
	return nil
}

// Resume возобновляет вывод ffmpeg клавишей u.
func (j *Job) Resume() error {
	if !j.paused.Load() {
		return nil
	}

	j.stopMux.Lock()
	stdin := j.stdin
	j.stopMux.Unlock()

	if stdin == nil {
		return errors.New("process input is not available")
	}
	if _, err := io.WriteString(stdin, "u"); err != nil {
		return err
	}

	j.paused.Store(false)
	return nil
}

// Paused сообщает, приостановлен ли процесс.
func (j *Job) Paused() bool { return j.paused.Load() }

// Throttled сообщает, что процесс приостановлен ограничителем скорости.
func (j *Job) Throttled() bool { return j.throttled.Load() }

// setThrottled запоминает причину паузы.
func (j *Job) setThrottled(v bool) { j.throttled.Store(v) }

// setPauseSupport сохраняет признак поддержки паузы сборкой.
func (j *Job) setPauseSupport(supported bool) { j.pauseOK.Store(supported) }

// pauseSupported сообщает, поддерживает ли сборка паузу через stdin.
func (j *Job) pauseSupported() bool { return j.pauseOK.Load() }

// AddServedSegment отмечает сегмент, отданный плееру.
func (j *Job) AddServedSegment() { j.segmentsServed.Add(1) }

// SegmentsServed возвращает число отданных плееру сегментов.
func (j *Job) SegmentsServed() int64 { return j.segmentsServed.Load() }

// markFinished фиксирует завершение процесса.
func (j *Job) markFinished(err error, exitCode int, canceled bool) {
	j.mu.Lock()
	j.finishedAt = time.Now()
	j.exitCode = exitCode
	switch {
	case canceled:
		j.state = StateCanceled
	case err != nil:
		j.lastErr = err
		j.state = StateFailed
	default:
		j.state = StateDone
	}
	j.mu.Unlock()
}

// signalDone снимает ожидание завершения процесса.
func (j *Job) signalDone() {
	j.doneOnce.Do(func() { close(j.finished) })
}
