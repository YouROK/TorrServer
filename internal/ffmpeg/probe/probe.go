package probe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"silo/internal/ffmpeg/media"
	"silo/internal/log"
)

const Tag = "[FFmpeg:Probe]"

var (
	ErrEmptyResult = errors.New("ffprobe returned no streams and no format")
	ErrTimeout     = errors.New("ffprobe timed out")
)

// Config описывает параметры запуска ffprobe.
type Config struct {
	Timeout        time.Duration // Ограничение на разбор одного файла
	AnalyzeDurMs   int           // Значение -analyzeduration в миллисекундах
	ProbeSizeBytes int64         // Значение -probesize в байтах
	HTTPHeaders    []string      // Дополнительные заголовки для HTTP-источника
	CacheSize      int           // Сколько результатов держать в памяти, 0 - без кэша
	CacheTTL       time.Duration // Время жизни записи кэша
}

// Prober запускает ffprobe и разбирает его вывод.
type Prober struct {
	binary string
	cfg    Config
	cache  *cache

	mu      sync.Mutex
	running map[*exec.Cmd]struct{}
}

// New создаёт Prober для указанного бинарника ffprobe.
func New(probePath string, cfg Config) *Prober {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.AnalyzeDurMs <= 0 {
		cfg.AnalyzeDurMs = 5000
	}
	if cfg.ProbeSizeBytes <= 0 {
		cfg.ProbeSizeBytes = 5 << 20
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 6 * time.Hour
	}

	p := &Prober{
		binary:  probePath,
		cfg:     cfg,
		running: make(map[*exec.Cmd]struct{}),
	}
	if cfg.CacheSize > 0 {
		p.cache = newCache(cfg.CacheSize, cfg.CacheTTL)
	}
	return p
}

// Available сообщает, задан ли путь к ffprobe.
func (p *Prober) Available() bool { return p != nil && p.binary != "" }

// Probe разбирает файл по ссылке и возвращает информацию о потоках.
func (p *Prober) Probe(ctx context.Context, input string) (*media.MediaInfo, error) {
	if !p.Available() {
		return nil, errors.New("ffprobe is not available")
	}

	if p.cache != nil {
		if info, ok := p.cache.get(input); ok {
			return info, nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	info, err := p.run(ctx, input)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w after %s", ErrTimeout, p.cfg.Timeout)
		}
		return nil, err
	}

	if p.cache != nil {
		p.cache.put(input, info)
	}
	return info, nil
}

// run выполняет ffprobe и нормализует вывод.
func (p *Prober) run(ctx context.Context, input string) (*media.MediaInfo, error) {
	args := p.buildArgs(input)

	cmd := exec.CommandContext(ctx, p.binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	p.track(cmd)
	defer p.untrack(cmd)

	if err := cmd.Run(); err != nil {
		msg := firstLine(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("ffprobe failed: %s", msg)
	}

	res, err := decode(stdout.Bytes())
	if err != nil {
		return nil, err
	}
	if len(res.Streams) == 0 && res.Format.FormatName == "" {
		return nil, ErrEmptyResult
	}

	info := normalize(res)
	log.Debugf("%s Parsed %s: %s, %d stream(s), %s", Tag, input, info.Container, len(info.Streams), media.FormatDuration(info.Duration))
	return info, nil
}

// buildArgs собирает аргументы запуска ffprobe.
func (p *Prober) buildArgs(input string) []string {
	args := []string{
		"-hide_banner",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-show_chapters",
		"-analyzeduration", strconv.Itoa(p.cfg.AnalyzeDurMs * 1000),
		"-probesize", strconv.FormatInt(p.cfg.ProbeSizeBytes, 10),
	}
	for _, h := range p.cfg.HTTPHeaders {
		args = append(args, "-headers", h)
	}
	return append(args, input)
}

// track регистрирует запущенный процесс.
func (p *Prober) track(cmd *exec.Cmd) {
	p.mu.Lock()
	p.running[cmd] = struct{}{}
	p.mu.Unlock()
}

// untrack снимает процесс с учёта.
func (p *Prober) untrack(cmd *exec.Cmd) {
	p.mu.Lock()
	delete(p.running, cmd)
	p.mu.Unlock()
}

// StopAll завершает все запущенные процессы ffprobe.
func (p *Prober) StopAll() {
	p.mu.Lock()
	cmds := make([]*exec.Cmd, 0, len(p.running))
	for cmd := range p.running {
		cmds = append(cmds, cmd)
	}
	p.mu.Unlock()

	for _, cmd := range cmds {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
}

// firstLine возвращает первую непустую строку текста.
func firstLine(s string) string {
	for _, line := range bytes.Split([]byte(s), []byte{'\n'}) {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 {
			return string(trimmed)
		}
	}
	return ""
}
