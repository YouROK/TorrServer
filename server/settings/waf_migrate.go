package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"server/log"
)

const (
	legacyWIPFile = "wip.txt"
	legacyBIPFile = "bip.txt"
)

// MigrateWAFLists performs a one-shot import of legacy HTTP ACL files
// (wip.txt whitelist, bip.txt blacklist) into settings.json when no waf
// object exists yet. After a successful write, each source file is renamed
// to *.bak and is not read again.
func MigrateWAFLists() {
	if ReadOnly {
		return
	}

	_, found, err := GetWAFConfig()
	if err != nil {
		log.TLogln("WAF migrate: read config:", err)
		return
	}
	if found {
		return
	}

	wipPath := filepath.Join(Path, legacyWIPFile)
	bipPath := filepath.Join(Path, legacyBIPFile)
	hasWIP := fileExists(wipPath)
	hasBIP := fileExists(bipPath)
	if !hasWIP && !hasBIP {
		return
	}

	whitelist, err := readLegacyListFile(wipPath, hasWIP)
	if err != nil {
		log.TLogln("WAF migrate: read", legacyWIPFile+":", err)
		return
	}
	blacklist, err := readLegacyListFile(bipPath, hasBIP)
	if err != nil {
		log.TLogln("WAF migrate: read", legacyBIPFile+":", err)
		return
	}

	if err := SetWAFConfig(WAFConfig{
		Version:   WAFConfigVersion,
		Whitelist: whitelist,
		Blacklist: blacklist,
		Referers:  nil,
	}); err != nil {
		log.TLogln("WAF migrate: write settings.json:", err)
		return
	}

	var backups []string
	if hasWIP {
		if err := renameToBak(wipPath); err != nil {
			log.TLogln("WAF migrate: backup", legacyWIPFile+":", err)
		} else {
			backups = append(backups, legacyWIPFile+".bak")
		}
	}
	if hasBIP {
		if err := renameToBak(bipPath); err != nil {
			log.TLogln("WAF migrate: backup", legacyBIPFile+":", err)
		} else {
			backups = append(backups, legacyBIPFile+".bak")
		}
	}

	log.TLogln(fmt.Sprintf(
		"WAF: migrated legacy %s/%s (%d white, %d black) → settings.json; backups: %s",
		legacyWIPFile, legacyBIPFile, len(whitelist), len(blacklist), strings.Join(backups, ", "),
	))
}

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func readLegacyListFile(path string, exists bool) ([]string, error) {
	if !exists {
		return []string{}, nil
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return splitNonEmptyLines(string(buf)), nil
}

// splitNonEmptyLines mirrors waf.textToList: keep comments and desc:rule
// lines, drop blank lines, normalize CRLF.
func splitNonEmptyLines(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

func renameToBak(path string) error {
	bak := path + ".bak"
	if fileExists(bak) {
		if err := os.Remove(bak); err != nil {
			return err
		}
	}
	return os.Rename(path, bak)
}
