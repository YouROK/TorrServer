package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const wafXPath = "Settings"
const wafConfigKey = "waf"

const WAFConfigVersion = 1

type WAFConfig struct {
	Version   int      `json:"version"`
	Whitelist []string `json:"whitelist"`
	Blacklist []string `json:"blacklist"`
	Referers  []string `json:"referers"`
}

// GetWAFConfig reads the top-level waf object from settings.json.
func GetWAFConfig() (WAFConfig, bool, error) {
	db, ok := NewJsonDB().(*JsonDB)
	if !ok {
		return WAFConfig{}, false, errors.New("settings JSON database unavailable")
	}
	filename, err := db.xPathToFilename(wafXPath)
	if err != nil {
		return WAFConfig{}, false, err
	}

	db.lock(filename)
	defer db.unlock(filename)
	root, err := readWAFSettings(db, filename)
	if err != nil {
		return WAFConfig{}, false, err
	}
	raw, found := root[wafConfigKey]
	if !found {
		return normalizeWAFConfig(WAFConfig{}), false, nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return WAFConfig{}, false, errors.New("waf setting must be an object")
	}
	var cfg WAFConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return WAFConfig{}, false, err
	}
	if cfg.Version == 0 {
		cfg.Version = WAFConfigVersion
	}
	return normalizeWAFConfig(cfg), true, nil
}

// SetWAFConfig replaces the complete waf object in settings.json while
// preserving every unrelated top-level setting.
func SetWAFConfig(cfg WAFConfig) error {
	if ReadOnly {
		return errors.New("read-only mode")
	}
	db, ok := NewJsonDB().(*JsonDB)
	if !ok {
		return errors.New("settings JSON database unavailable")
	}
	cfg = normalizeWAFConfig(cfg)
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	filename, err := db.xPathToFilename(wafXPath)
	if err != nil {
		return err
	}
	db.lock(filename)
	defer db.unlock(filename)
	root, err := readWAFSettings(db, filename)
	if err != nil {
		return err
	}
	root[wafConfigKey] = data
	return writeWAFSettings(db, filename, root)
}

func readWAFSettings(db *JsonDB, filename string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(filepath.Join(db.Path, filename))
	if errors.Is(err, os.ErrNotExist) {
		return map[string]json.RawMessage{}, nil
	}
	if err != nil {
		return nil, err
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if root == nil {
		root = map[string]json.RawMessage{}
	}
	return root, nil
}

func writeWAFSettings(db *JsonDB, filename string, root map[string]json.RawMessage) error {
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(db.Path, filename), data, db.fileMode)
}

func normalizeWAFConfig(cfg WAFConfig) WAFConfig {
	cfg.Version = WAFConfigVersion
	if cfg.Whitelist == nil {
		cfg.Whitelist = []string{}
	}
	if cfg.Blacklist == nil {
		cfg.Blacklist = []string{}
	}
	if cfg.Referers == nil {
		cfg.Referers = []string{}
	}
	return cfg
}
