package plugin

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/fusefs"
	"silo/internal/user"

	"github.com/dop251/goja"
)

// newRuntimeForFuse собирает рантайм с объектом ts.fuse.
func newRuntimeForFuse(t *testing.T) *JSRuntime {
	t.Helper()

	db, err := database.Open(t.TempDir() + "/fuse.db")
	if err != nil {
		t.Fatalf("база не открылась: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userSvc := user.NewService(user.NewStore(db), config.DefaultConfig())

	vm := goja.New()
	rt := &JSRuntime{pluginID: "tester", vm: vm, userSvc: userSvc}
	vm.Set("fuse", rt.createFuseModule(fusefs.NewManager()))
	return rt
}

// fuseCall вызывает метод ts.fuse и разбирает ответ в объект.
func fuseCall(t *testing.T, rt *JSRuntime, expr string) map[string]any {
	t.Helper()

	v, err := rt.vm.RunString("JSON.stringify(" + expr + ")")
	if err != nil {
		t.Fatalf("%s: %v", expr, err)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(v.String()), &out); err != nil {
		t.Fatalf("%s: ответ не разобран: %v", expr, err)
	}
	return out
}

func TestFuseStatusWithoutMount(t *testing.T) {
	rt := newRuntimeForFuse(t)

	status := fuseCall(t, rt, "fuse.status()")
	if status["mounted"] != false {
		t.Errorf("mounted = %v, want false", status["mounted"])
	}
	if _, ok := status["supported"]; !ok {
		t.Error("status без поля supported")
	}
}

func TestFuseUnmountWithoutMount(t *testing.T) {
	rt := newRuntimeForFuse(t)

	result := fuseCall(t, rt, "fuse.unmount()")
	if result["ok"] != false {
		t.Errorf("ok = %v, want false", result["ok"])
	}
	if result["code"] != "not_mounted" {
		t.Errorf("code = %v, want not_mounted", result["code"])
	}
}

func TestFuseMountRejectsEmptyPath(t *testing.T) {
	rt := newRuntimeForFuse(t)

	result := fuseCall(t, rt, "fuse.mount('')")
	if result["ok"] != false {
		t.Errorf("ok = %v, want false", result["ok"])
	}
	if result["code"] != "empty_path" {
		t.Errorf("code = %v, want empty_path", result["code"])
	}
}

func TestFuseMountRejectsRelativePath(t *testing.T) {
	if !fusefs.Supported() {
		t.Skip("платформа без FUSE")
	}
	rt := newRuntimeForFuse(t)

	result := fuseCall(t, rt, "fuse.mount('mnt/silo')")
	if result["ok"] != false {
		t.Errorf("ok = %v, want false", result["ok"])
	}
	if result["code"] != "not_absolute" {
		t.Errorf("code = %v, want not_absolute", result["code"])
	}
}

// TestFuseErrorCodesHaveTranslations следит, чтобы каждый код отказа был
// переведён в плагине: иначе пользователь увидит ключ вместо текста.
func TestFuseErrorCodesHaveTranslations(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{
		{fusefs.ErrUnsupported, "unsupported"},
		{fusefs.ErrEmptyPath, "empty_path"},
		{fusefs.ErrNotAbsolute, "not_absolute"},
		{fusefs.ErrNotExist, "not_exist"},
		{fusefs.ErrPermission, "permission"},
		{fusefs.ErrNotDir, "not_a_directory"},
		{fusefs.ErrAlreadyMounted, "already_mounted"},
		{fusefs.ErrNotMounted, "not_mounted"},
		{fusefs.ErrNoSource, "no_source"},
		{fusefs.ErrAllowOther, "allow_other"},
		{fusefs.ErrNoDevice, "no_device"},
		{fusefs.ErrMountFailed, "mount_failed"},
		{errors.New("boom"), "mount_failed"},
	}

	root := filepath.Join("..", "..", "plugs", "fuse", "i18n")
	if _, err := os.Stat(root); err != nil {
		t.Skipf("плагин fuse недоступен: %v", err)
	}

	for _, lang := range []string{"en.json", "ru.json"} {
		data, err := os.ReadFile(filepath.Join(root, lang))
		if err != nil {
			t.Fatalf("%s не читается: %v", lang, err)
		}

		var dict map[string]string
		if err := json.Unmarshal(data, &dict); err != nil {
			t.Fatalf("%s не разобран: %v", lang, err)
		}

		for _, c := range cases {
			if got := fuseErrorCode(c.err); got != c.code {
				t.Errorf("fuseErrorCode(%v) = %q, want %q", c.err, got, c.code)
			}
			if _, ok := dict["plugin.fuse.error_"+c.code]; !ok {
				t.Errorf("%s: нет перевода plugin.fuse.error_%s", lang, c.code)
			}
		}
	}
}

// TestFuseErrorCarriesDetail проверяет, что категория ошибки отделена от
// подробности: категория переводится, подробность показывается как есть.
func TestFuseErrorCarriesDetail(t *testing.T) {
	if !fusefs.Supported() {
		t.Skip("платформа без FUSE")
	}
	rt := newRuntimeForFuse(t)

	result := fuseCall(t, rt, "fuse.mount('/nonexistent/silo/path')")
	if result["ok"] != false {
		t.Fatalf("ok = %v, want false", result["ok"])
	}
	if result["code"] != "not_exist" {
		t.Errorf("code = %v, want not_exist", result["code"])
	}
	if detail, _ := result["detail"].(string); !strings.Contains(detail, "/nonexistent/silo/path") {
		t.Errorf("detail = %v, want путь в подробности", result["detail"])
	}
}

// TestFuseErrorWrappedKeepsKind проверяет, что обёрнутая ошибка
// распознаётся по категории.
func TestFuseErrorWrappedKeepsKind(t *testing.T) {
	wrapped := &fusefs.Error{Kind: fusefs.ErrNotExist, Detail: "/tmp/nope"}

	if got := fuseErrorCode(wrapped); got != "not_exist" {
		t.Errorf("fuseErrorCode = %q, want not_exist", got)
	}
	if !errors.Is(wrapped, fusefs.ErrNotExist) {
		t.Error("обёрнутая ошибка не распознаётся через errors.Is")
	}
}
