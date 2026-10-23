//go:build linux && !android

package fusefs

import (
	"bufio"
	"os"
	"strings"
)

// fuseConfPath - файл настроек FUSE, где разрешается allow_other.
const fuseConfPath = "/etc/fuse.conf"

// preflightMount проверяет условия, о которых fusermount сообщает только
// в stderr. go-fuse этот текст не читает и отдаёт лишь код возврата,
// поэтому причину отказа определяем заранее.
func preflightMount(opts Options) error {
	if !opts.AllowOther {
		return nil
	}
	// Суперпользователю allow_other разрешён без настройки.
	if os.Geteuid() == 0 {
		return nil
	}
	if fuseConfAllowsOther() {
		return nil
	}
	return ErrAllowOther
}

// fuseConfAllowsOther ищет в fuse.conf незакомментированный параметр
// user_allow_other.
func fuseConfAllowsOther() bool {
	return parseFuseConfAllowsOther(fuseConfPath)
}

// parseFuseConfAllowsOther читает указанный файл настроек FUSE.
func parseFuseConfAllowsOther(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "user_allow_other") {
			return true
		}
	}
	return false
}
