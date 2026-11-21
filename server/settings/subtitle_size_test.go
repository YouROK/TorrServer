package settings

import (
	"testing"
	"time"
)

func TestMaxSubtitleSize(t *testing.T) {
	old := BTsets
	defer func() { BTsets = old }()

	cases := []struct {
		mb   int
		want int64
	}{
		{0, 16 << 20},
		{-5, 16 << 20},
		{2, 2 << 20},
		{8, 8 << 20},
		{32, 32 << 20},
		{500, 32 << 20},
	}
	for _, c := range cases {
		BTsets = &BTSets{MaxSubtitleSizeMB: c.mb}
		if got := MaxSubtitleSize(); got != c.want {
			t.Errorf("MaxSubtitleSizeMB=%d: MaxSubtitleSize() = %d, want %d", c.mb, got, c.want)
		}
	}

	BTsets = nil
	if got := MaxSubtitleSize(); got != 16<<20 {
		t.Errorf("no settings: MaxSubtitleSize() = %d, want %d", got, 16<<20)
	}
}

func TestExternalSubtitleTimeout(t *testing.T) {
	old := BTsets
	defer func() { BTsets = old }()

	cases := []struct {
		sec  int
		want time.Duration
	}{
		{0, 15 * time.Second},
		{-1, 15 * time.Second},
		{1, time.Second},
		{30, 30 * time.Second},
		{60, 60 * time.Second},
		{600, 60 * time.Second},
	}
	for _, c := range cases {
		BTsets = &BTSets{ExternalSubtitleTimeoutSec: c.sec}
		if got := ExternalSubtitleTimeout(); got != c.want {
			t.Errorf("ExternalSubtitleTimeoutSec=%d: ExternalSubtitleTimeout() = %v, want %v", c.sec, got, c.want)
		}
	}

	BTsets = nil
	if got := ExternalSubtitleTimeout(); got != 15*time.Second {
		t.Errorf("no settings: ExternalSubtitleTimeout() = %v, want 15s", got)
	}
}

func TestExternalSubtitleAttempts(t *testing.T) {
	old := BTsets
	defer func() { BTsets = old }()

	cases := []struct{ n, want int }{
		{0, 3}, {-1, 3}, {1, 1}, {5, 5}, {10, 10}, {11, 10},
	}
	for _, c := range cases {
		BTsets = &BTSets{ExternalSubtitleAttempts: c.n}
		if got := ExternalSubtitleAttempts(); got != c.want {
			t.Errorf("ExternalSubtitleAttempts=%d: ExternalSubtitleAttempts() = %d, want %d", c.n, got, c.want)
		}
	}

	BTsets = nil
	if got := ExternalSubtitleAttempts(); got != 3 {
		t.Errorf("no settings: ExternalSubtitleAttempts() = %d, want 3", got)
	}
}

func TestNormalizeSubtitles(t *testing.T) {
	s := &BTSets{}
	s.normalizeSubtitles()
	if s.MaxSubtitleSizeMB != 16 || s.ExternalSubtitleTimeoutSec != 15 || s.ExternalSubtitleAttempts != 3 {
		t.Errorf("zero settings: got size=%d timeout=%d attempts=%d, want 16, 15, 3",
			s.MaxSubtitleSizeMB, s.ExternalSubtitleTimeoutSec, s.ExternalSubtitleAttempts)
	}

	s = &BTSets{MaxSubtitleSizeMB: 33, ExternalSubtitleTimeoutSec: 61, ExternalSubtitleAttempts: 11}
	s.normalizeSubtitles()
	if s.MaxSubtitleSizeMB != 32 || s.ExternalSubtitleTimeoutSec != 60 || s.ExternalSubtitleAttempts != 10 {
		t.Errorf("too big settings: got size=%d timeout=%d attempts=%d, want 32, 60, 10",
			s.MaxSubtitleSizeMB, s.ExternalSubtitleTimeoutSec, s.ExternalSubtitleAttempts)
	}

	s = &BTSets{MaxSubtitleSizeMB: 4, ExternalSubtitleTimeoutSec: 30, ExternalSubtitleAttempts: 2}
	s.normalizeSubtitles()
	if s.MaxSubtitleSizeMB != 4 || s.ExternalSubtitleTimeoutSec != 30 || s.ExternalSubtitleAttempts != 2 {
		t.Errorf("valid settings changed: got size=%d timeout=%d attempts=%d, want 4, 30, 2",
			s.MaxSubtitleSizeMB, s.ExternalSubtitleTimeoutSec, s.ExternalSubtitleAttempts)
	}
}
