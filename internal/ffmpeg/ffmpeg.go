package ffmpeg

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/job"
	"silo/internal/ffmpeg/media"
	"silo/internal/ffmpeg/probe"
	"silo/internal/log"
)

const ProbeTag = "[FFmpeg:Probe]"

// StartRequest описывает запрос на транскодирование.
type StartRequest struct {
	ID        string // Пусто - идентификатор создаётся автоматически
	UserID    string
	Hash      string
	FileIdx   int
	ProfileID string
	Options   args.Options
	Duration  time.Duration
}

// FFmpeg объединяет поиск бинарей, разбор медиа и запуск процессов.
type FFmpeg struct {
	cfg     Config
	dataDir string
	binary  *Binary
	prober  *probe.Prober
	jobs    *job.Manager

	mu      sync.RWMutex
	lastErr error
}

// New создаёт модуль транскодинга и ищет подходящий ffmpeg.
func New(cfg Config, dataDir string) *FFmpeg {
	cfg.Normalize()

	f := &FFmpeg{cfg: cfg, dataDir: dataDir}

	if cfg.validate() != nil || !cfg.Enabled {
		f.lastErr = errors.New("transcoding is disabled")
		return f
	}

	bin, err := probeBinary()
	if err != nil {
		f.lastErr = err
		log.Errorf("%s Transcoding disabled: %v", Tag, err)
		return f
	}

	return f.setup(bin)
}

// newWithBinary создаёт модуль на заданном бинарнике, минуя поиск.
func newWithBinary(cfg Config, dataDir, ffmpegPath string) *FFmpeg {
	cfg.Normalize()

	f := &FFmpeg{cfg: cfg, dataDir: dataDir}
	if !cfg.Enabled {
		f.lastErr = errors.New("transcoding is disabled")
		return f
	}

	bin, err := inspect(ffmpegPath, cfg)
	if err != nil {
		f.lastErr = err
		log.Errorf("%s Transcoding disabled: %v", Tag, err)
		return f
	}
	return f.setup(bin)
}

// setup достраивает модуль после успешной проверки бинарника.
func (f *FFmpeg) setup(bin *Binary) *FFmpeg {
	cacheDir := f.cfg.ResolveCacheDir(f.dataDir)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		f.lastErr = fmt.Errorf("failed to create cache dir: %w", err)
		log.Errorf("%s Transcoding disabled: %v", Tag, f.lastErr)
		return f
	}

	f.binary = bin

	// Режим auto выбирает ускорение по фактически работающему устройству
	if f.cfg.HWAccel == "auto" {
		resolved := bin.PreferredHWAccel()
		f.cfg.HWAccel = resolved
		if resolved == "none" {
			log.Warnf("%s No usable hardware acceleration found, using software", Tag)
		} else {
			log.Infof("%s Hardware acceleration selected: %s", Tag, resolved)
		}
	}

	f.prober = probe.New(bin.FFprobe, probe.Config{
		Timeout:   f.cfg.probeTimeout(),
		CacheSize: 128,
	})
	// Поддержку паузы проверяем один раз: от неё зависит ограничение скорости
	pauseSupport := bin.DetectPause()

	f.jobs = job.NewManager(job.Config{
		Binary:         bin.FFmpeg,
		Selector:       bin,
		MaxSessions:    f.cfg.MaxSessions,
		StopTimeout:    DefaultStopTimeout,
		KeepStderr:     f.cfg.KeepStderr,
		PauseSupport:   pauseSupport && f.cfg.ThrottleEnabled(),
		KeepAhead:      f.cfg.KeepAhead,
		ProgressPeriod: DefaultProgressPeriod,
	})

	log.Infof("%s Using ffmpeg %s at %s", Tag, bin.Version.String(), bin.FFmpeg)
	log.Infof("%s ffprobe at %s", ProbeTag, bin.FFprobe)

	if pauseSupport {
		log.Infof("%s Pause and resume are supported via stdin", Tag)
	} else {
		log.Warnf("%s Pause is not supported by this build, throttling is disabled", Tag)
	}
	return f
}

// Enabled сообщает, готов ли модуль к работе.
func (f *FFmpeg) Enabled() bool {
	return f != nil && f.binary != nil && f.jobs != nil && f.jobs.Enabled()
}

// Reason возвращает причину, по которой транскодинг недоступен.
func (f *FFmpeg) Reason() string {
	if f == nil {
		return "not initialized"
	}
	if f.Enabled() {
		return ""
	}
	if f.lastErr != nil {
		return f.lastErr.Error()
	}
	return "unknown reason"
}

// Info описывает найденный бинарь и его возможности.
type Info struct {
	Enabled      bool     `json:"enabled"`
	Reason       string   `json:"reason,omitempty"`
	FFmpegPath   string   `json:"ffmpeg_path,omitempty"`
	FFprobePath  string   `json:"ffprobe_path,omitempty"`
	Version      string   `json:"version,omitempty"`
	PauseSupport bool     `json:"pause_support"`
	CacheDir     string   `json:"cache_dir"`
	MaxSessions  int      `json:"max_sessions"`
	ActiveJobs   int      `json:"active_jobs"`
	Encoders     []string `json:"encoders,omitempty"`
	Hwaccels     []string `json:"hwaccels,omitempty"`
}

// Info возвращает сведения о состоянии модуля.
func (f *FFmpeg) Info() Info {
	if f == nil {
		return Info{Reason: "not initialized"}
	}

	info := Info{Enabled: f.Enabled(), Reason: f.Reason(), MaxSessions: f.cfg.MaxSessions}
	if !f.Enabled() {
		return info
	}

	info.FFmpegPath = f.binary.FFmpeg
	info.FFprobePath = f.binary.FFprobe
	info.Version = f.binary.Version.String()
	info.PauseSupport = f.binary.PauseSupport
	info.ActiveJobs = f.jobs.Count()
	info.CacheDir = f.CacheDir()
	for name := range f.binary.Encoders {
		info.Encoders = append(info.Encoders, name)
	}
	for name := range f.binary.Hwaccels {
		info.Hwaccels = append(info.Hwaccels, name)
	}
	return info
}

// Probe разбирает медиафайл по ссылке.
func (f *FFmpeg) Probe(ctx context.Context, input string) (*media.MediaInfo, error) {
	if !f.Enabled() {
		return nil, fmt.Errorf("transcoding is unavailable: %s", f.Reason())
	}
	return f.prober.Probe(ctx, input)
}

// Start запускает процесс транскодирования.
func (f *FFmpeg) Start(ctx context.Context, req StartRequest) (*job.Job, error) {
	if !f.Enabled() {
		return nil, fmt.Errorf("transcoding is unavailable: %s", f.Reason())
	}

	if req.Options.Threads == 0 {
		req.Options.Threads = f.cfg.Threads
	}
	if req.Options.Video.HWAccel == "" {
		req.Options.Video.HWAccel = f.cfg.HWAccel
	}
	if req.Options.Video.HWDevice == "" {
		req.Options.Video.HWDevice = f.cfg.HWDevicePath()
	}

	id := req.ID
	if id == "" {
		id = newSessionID()
	}

	var workDir, playlistPath string
	if req.Options.Protocol == args.ProtocolHLS {
		workDir = filepath.Join(f.CacheDir(), id)
		if err := os.MkdirAll(workDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create session dir: %w", err)
		}
		playlistPath = filepath.Join(workDir, "index.m3u8")
		req.Options.Output = playlistPath

		// Расширение сегмента зависит от типа: плеер определяет формат по нему
		pattern := "seg%d.ts"
		if req.Options.HLS.SegmentType == "fmp4" {
			pattern = "seg%d.m4s"
			if req.Options.HLS.InitFilename == "" {
				req.Options.HLS.InitFilename = args.DefaultInitSegment
			}
		}
		req.Options.HLS.SegmentFilename = filepath.Join(workDir, pattern)
	}

	jobReq := job.StartRequest{
		ID:           id,
		UserID:       req.UserID,
		Hash:         req.Hash,
		FileIdx:      req.FileIdx,
		ProfileID:    req.ProfileID,
		WorkDir:      workDir,
		PlaylistPath: playlistPath,
		Options:      req.Options,
		Duration:     req.Duration,
	}
	return f.jobs.Start(ctx, jobReq)
}

// SessionDir возвращает каталог задания по идентификатору сессии.
func (f *FFmpeg) SessionDir(id string) string {
	if f == nil || id == "" {
		return ""
	}
	return filepath.Join(f.CacheDir(), id)
}

// Stop останавливает задание по идентификатору.
func (f *FFmpeg) Stop(id string) bool {
	if !f.Enabled() {
		return false
	}
	return f.jobs.Stop(id)
}

// Pause приостанавливает вывод задания.
// Поддержка зависит от сборки ffmpeg, поэтому отсутствие поддержки не считается ошибкой запуска.
func (f *FFmpeg) Pause(id string) error {
	if !f.Enabled() {
		return errors.New("transcoding is unavailable")
	}
	// Наличие сессии проверяется до поддержки паузы:
	// иначе несуществующая сессия выглядит как неподдерживаемая сборка
	if f.jobs.Get(id) == nil {
		return job.ErrNotFound
	}
	if !f.jobPauseSupport() {
		return job.ErrPauseUnsupported
	}
	return f.jobs.Pause(id)
}

// Resume возобновляет вывод задания.
func (f *FFmpeg) Resume(id string) error {
	if !f.Enabled() {
		return errors.New("transcoding is unavailable")
	}
	if f.jobs.Get(id) == nil {
		return job.ErrNotFound
	}
	if !f.jobPauseSupport() {
		return job.ErrPauseUnsupported
	}
	return f.jobs.Resume(id)
}

// jobPauseSupport сообщает, умеет ли сборка приостанавливать вывод.
func (f *FFmpeg) jobPauseSupport() bool {
	return f.binary != nil && f.binary.PauseSupport
}

// Jobs возвращает снимки активных заданий.
func (f *FFmpeg) Jobs() []job.Snapshot {
	if !f.Enabled() {
		return nil
	}
	return f.jobs.List()
}

// SelectVideoEncoder подбирает кодировщик для кодека и ускорения.
func (f *FFmpeg) SelectVideoEncoder(codec, hwaccel string) string {
	if !f.Enabled() {
		return ""
	}
	return f.binary.AutoVideoEncoder(codec, hwaccel)
}

// SelectAudioEncoder подбирает аудиокодировщик для кодека.
func (f *FFmpeg) SelectAudioEncoder(codec string) string {
	if !f.Enabled() {
		return ""
	}
	return f.binary.AutoAudioEncoder(codec)
}

// SetEventHandler задаёт обработчик событий модуля.
func (f *FFmpeg) SetEventHandler(fn func(topic string, payload any)) {
	if f == nil || f.jobs == nil {
		return
	}
	f.jobs.SetEventHandler(fn)
}

// StopAll останавливает все задания и разбор медиа.
func (f *FFmpeg) StopAll() {
	if f == nil {
		return
	}
	if f.jobs != nil {
		f.jobs.StopAll()
	}
	if f.prober != nil {
		f.prober.StopAll()
	}
}

// Wait ожидает завершения фоновых операций.
func (f *FFmpeg) Wait() {
	if f != nil && f.jobs != nil {
		f.jobs.Wait()
	}
}

// CacheDir возвращает каталог файлов транскодирования.
func (f *FFmpeg) CacheDir() string {
	if f == nil {
		return ""
	}
	return f.cfg.ResolveCacheDir(f.dataDir)
}

// DefaultTestOptions собирает тестовое задание для проверки транскодирования.
func (f *FFmpeg) DefaultTestOptions(input string) args.Options {
	o := args.DefaultOptions(input)
	if f.Enabled() {
		o.Video.Encoder = f.binary.AutoVideoEncoder(o.Video.Codec, f.cfg.HWAccel)
		o.Audio.Encoder = f.binary.AutoAudioEncoder(o.Audio.Codec)
	}
	return o
}

// newSessionID создаёт непредсказуемый идентификатор сессии транскодирования.
func newSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
