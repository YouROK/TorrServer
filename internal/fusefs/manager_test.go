package fusefs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPreparePathRejectsEmpty(t *testing.T) {
	if _, err := preparePath("   "); !errors.Is(err, ErrEmptyPath) {
		t.Errorf("err = %v, want ErrEmptyPath", err)
	}
}

func TestPreparePathRejectsRelative(t *testing.T) {
	if _, err := preparePath("mnt/silo"); !errors.Is(err, ErrNotAbsolute) {
		t.Errorf("err = %v, want ErrNotAbsolute", err)
	}
}

func TestPreparePathRejectsFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatalf("файл не создан: %v", err)
	}

	if _, err := preparePath(f); !errors.Is(err, ErrNotDir) {
		t.Errorf("err = %v, want ErrNotDir", err)
	}
}

func TestPreparePathRejectsMissing(t *testing.T) {
	_, err := preparePath(filepath.Join(t.TempDir(), "nope"))
	if err == nil || errors.Is(err, ErrNotDir) {
		t.Errorf("err = %v, want ошибку доступа к каталогу", err)
	}
}

func TestPreparePathCleans(t *testing.T) {
	dir := t.TempDir()

	got, err := preparePath(filepath.Join(dir, "sub", "..") + string(filepath.Separator))
	if err != nil {
		t.Fatalf("preparePath: %v", err)
	}
	if got != dir {
		t.Errorf("path = %q, want %q", got, dir)
	}
}

// TestMountValidatesPathBeforeSource проверяет порядок проверок: ошибка
// в пути должна возвращаться раньше, чем отсутствие источника.
func TestMountValidatesPathBeforeSource(t *testing.T) {
	m := NewManager()

	_, err := m.Mount("", Options{})
	if !errors.Is(err, ErrEmptyPath) && !errors.Is(err, ErrUnsupported) {
		t.Errorf("err = %v, want ErrEmptyPath или ErrUnsupported", err)
	}
}

func TestUnmountWithoutMount(t *testing.T) {
	m := NewManager()

	if _, err := m.Unmount(); !errors.Is(err, ErrNotMounted) {
		t.Errorf("err = %v, want ErrNotMounted", err)
	}
}

func TestStatusWithoutMount(t *testing.T) {
	m := NewManager()

	if _, mounted := m.Status(); mounted {
		t.Error("Status сообщил о монтировании без него")
	}
	if m.Mounted() {
		t.Error("Mounted = true без монтирования")
	}
}

// TestCloseWithoutMount идемпотентен: выгрузка плагина не должна падать.
func TestCloseWithoutMount(t *testing.T) {
	m := NewManager()

	if err := m.Close(); err != nil {
		t.Errorf("Close = %v, want nil", err)
	}
	if err := m.Close(); err != nil {
		t.Errorf("повторный Close = %v, want nil", err)
	}
}

func TestSupportedMatchesPlatform(t *testing.T) {
	if Supported() != platformSupported() {
		t.Error("Supported расходится с platformSupported")
	}
}
