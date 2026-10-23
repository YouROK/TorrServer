//go:build linux && !android

package fusefs

import (
	"errors"
	"os"
	"testing"
)

// TestParseFuseConfAllowsOther проверяет разбор fuse.conf: параметр
// учитывается только вне комментария.
func TestParseFuseConfAllowsOther(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/fuse.conf"

	cases := []struct {
		body string
		want bool
	}{
		{"#user_allow_other\n", false},
		{"user_allow_other\n", true},
		{"  user_allow_other  \n", true},
		{"# comment\nuser_allow_other\n", true},
		{"", false},
		{"other_option\n", false},
	}

	for _, c := range cases {
		if err := os.WriteFile(path, []byte(c.body), 0644); err != nil {
			t.Fatalf("файл не записан: %v", err)
		}
		if got := parseFuseConfAllowsOther(path); got != c.want {
			t.Errorf("тело %q -> %v, want %v", c.body, got, c.want)
		}
	}
}

// TestParseFuseConfMissingFile проверяет, что отсутствующий файл не
// разрешает allow_other.
func TestParseFuseConfMissingFile(t *testing.T) {
	if parseFuseConfAllowsOther(t.TempDir() + "/absent.conf") {
		t.Error("отсутствующий файл не должен разрешать allow_other")
	}
}

// TestPreflightWithoutAllowOther проверяет, что проверка не мешает
// обычному монтированию.
func TestPreflightWithoutAllowOther(t *testing.T) {
	if err := preflightMount(Options{}); err != nil {
		t.Errorf("preflightMount без allow_other = %v, want nil", err)
	}
}

// TestPreflightAllowOtherRejected проверяет отказ, когда параметр
// запрошен, но не разрешён в fuse.conf.
func TestPreflightAllowOtherRejected(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("под root allow_other разрешён без настройки")
	}
	if parseFuseConfAllowsOther(fuseConfPath) {
		t.Skip("в системе уже разрешён user_allow_other")
	}

	err := preflightMount(Options{AllowOther: true})
	if !errors.Is(err, ErrAllowOther) {
		t.Errorf("err = %v, want ErrAllowOther", err)
	}
}
