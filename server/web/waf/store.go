package waf

import (
	"fmt"
	"strings"
	"sync"

	"server/log"
	"server/settings"
)

// Snapshot is an immutable view of current WAF configuration.
type Snapshot struct {
	WhiteIP        Ranger
	BlackIP        Ranger
	Referers       []string // defaults + user referer hosts
	WhitelistText  string
	BlacklistText  string
	ReferersText   string
	Warnings       []Warning
	IPEnabled      bool
	RefererEnabled bool
}

type Warning struct {
	List string `json:"list"`
	Line int    `json:"line,omitempty"`
	Code string `json:"code"`
}

// ListsUpdate is the payload for replacing list contents.
type ListsUpdate struct {
	Whitelist string
	Blacklist string
	Referers  string
}

type store struct {
	mu       sync.RWMutex
	updateMu sync.Mutex
	snap     Snapshot
}

var global = &store{}

// Load reads WAF lists from settings.json into the global store.
func Load() {
	global.reload()
}

// Reload re-reads WAF lists into the global store.
func Reload() {
	global.reload()
}

// GetSnapshot returns a copy of the current WAF state.
func GetSnapshot() Snapshot {
	return global.snapshot()
}

// Update validates, stores lists in settings.json, and reloads in-memory state.
func Update(u ListsUpdate) (Snapshot, error) {
	return global.update(u)
}

// DefaultReferers returns the built-in blocked referer hosts.
func DefaultReferers() []string {
	out := make([]string, len(defaultBlockedReferers))
	copy(out, defaultBlockedReferers)
	return out
}

func (s *store) snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneSnapshot(s.snap)
}

func (s *store) reload() {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()
	snap := loadLists()
	s.mu.Lock()
	s.snap = snap
	s.mu.Unlock()
}

func (s *store) update(u ListsUpdate) (Snapshot, error) {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	if err := settings.SetWAFConfig(settings.WAFConfig{
		Whitelist: textToList(u.Whitelist),
		Blacklist: textToList(u.Blacklist),
		Referers:  textToList(u.Referers),
	}); err != nil {
		return Snapshot{}, err
	}

	snap := loadLists()
	s.mu.Lock()
	s.snap = snap
	s.mu.Unlock()
	return cloneSnapshot(snap), nil
}

func loadLists() Snapshot {
	cfg, _, err := settings.GetWAFConfig()
	var storageWarnings []Warning
	if err != nil {
		log.TLogln("WAF: load config:", err)
		storageWarnings = append(storageWarnings, Warning{List: "storage", Code: "read_failed"})
	}
	whiteBuf := []byte(listToText(cfg.Whitelist))
	blackBuf := []byte(listToText(cfg.Blacklist))
	referBuf := []byte(listToText(cfg.Referers))

	white, wWarn := scanBuf(whiteBuf)
	black, bWarn := scanBuf(blackBuf)
	userReferers, rWarn := scanRefererBufValidated(referBuf)
	referers := mergeReferers(defaultBlockedReferers, userReferers)

	warnings := storageWarnings
	for _, w := range wWarn {
		warnings = append(warnings, Warning{List: "whitelist", Line: w.Line, Code: w.Code})
	}
	for _, w := range bWarn {
		warnings = append(warnings, Warning{List: "blacklist", Line: w.Line, Code: w.Code})
	}
	for _, w := range rWarn {
		warnings = append(warnings, Warning{List: "referers", Line: w.Line, Code: w.Code})
	}
	for _, w := range warnings {
		log.TLogln("WAF warning:", fmt.Sprintf("%s line %d: %s", w.List, w.Line, w.Code))
	}

	return Snapshot{
		WhiteIP:        white,
		BlackIP:        black,
		Referers:       referers,
		WhitelistText:  string(whiteBuf),
		BlacklistText:  string(blackBuf),
		ReferersText:   string(referBuf),
		Warnings:       warnings,
		IPEnabled:      white.NumRanges() > 0 || black.NumRanges() > 0,
		RefererEnabled: len(referers) > 0,
	}
}

func cloneSnapshot(s Snapshot) Snapshot {
	out := s
	if s.Referers != nil {
		out.Referers = append([]string(nil), s.Referers...)
	}
	if s.Warnings != nil {
		out.Warnings = append([]Warning(nil), s.Warnings...)
	}
	return out
}

func textToList(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

func listToText(lines []string) string {
	return strings.Join(lines, "\n")
}
