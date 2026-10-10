package job

import (
	"strconv"
	"strings"
	"time"
)

// Progress описывает текущее положение транскодирования.
type Progress struct {
	Frame      int64         `json:"frame"`
	FPS        float64       `json:"fps"`
	BitrateBps int64         `json:"bitrate_bps"`
	TotalSize  int64         `json:"total_size"`
	OutTime    time.Duration `json:"out_time"`
	Speed      float64       `json:"speed"`
	Done       bool          `json:"done"`
}

// Percent возвращает долю выполненного от общей длительности.
func (p Progress) Percent(duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	v := float64(p.OutTime) / float64(duration) * 100
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// parseProgressLine разбирает пару ключ-значение из вывода -progress.
func parseProgressLine(p *Progress, line string) {
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return
	}
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	switch key {
	case "frame":
		p.Frame = parseI64(value, p.Frame)
	case "fps":
		p.FPS = parseFloat(value, p.FPS)
	case "bitrate":
		p.BitrateBps = parseBitrate(value)
	case "total_size":
		p.TotalSize = parseI64(value, p.TotalSize)
	case "out_time_us":
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			p.OutTime = time.Duration(v) * time.Microsecond
		}
	case "out_time_ms":
		// Значение приходит в микросекундах, несмотря на имя ключа
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			p.OutTime = time.Duration(v) * time.Microsecond
		}
	case "out_time":
		if d, err := parseTime(value); err == nil {
			p.OutTime = d
		}
	case "speed":
		p.Speed = parseSpeed(value)
	case "progress":
		p.Done = value == "end"
	}
}

// parseI64 разбирает целое число, сохраняя прежнее значение при ошибке.
func parseI64(s string, fallback int64) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fallback
	}
	return v
}

// parseFloat разбирает дробное число, сохраняя прежнее значение при ошибке.
func parseFloat(s string, fallback float64) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return v
}

// parseBitrate переводит значение вида "1234.5kbits/s" в биты в секунду.
func parseBitrate(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "kbits/s")
	s = strings.TrimSuffix(s, "bits/s")
	s = strings.TrimSuffix(s, "kb/s")
	s = strings.TrimSuffix(s, "k")
	s = strings.TrimSuffix(s, "N/A")
	if s == "" {
		return 0
	}

	multiplier := int64(1)
	if strings.Contains(s, ".") {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return int64(v * 1000 * float64(multiplier))
	}

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v * 1000 * multiplier
}

// parseSpeed переводит значение вида "1.23x" в число.
func parseSpeed(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "x")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// parseTime разбирает значение вида "00:01:02.345000".
func parseTime(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, strconv.ErrSyntax
	}

	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}

	total := hours*3600 + minutes*60 + seconds
	return time.Duration(total * float64(time.Second)), nil
}
