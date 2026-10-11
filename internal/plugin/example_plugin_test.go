package plugin

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// stubRegistrar запоминает обращения плагина к веб-слою.
type stubRegistrar struct {
	handlers map[string]bool // "GET /api/info"
	static   map[string]bool
	dirs     map[string]string
}

func newStubRegistrar() *stubRegistrar {
	return &stubRegistrar{
		handlers: make(map[string]bool),
		static:   make(map[string]bool),
		dirs:     make(map[string]string),
	}
}

func (s *stubRegistrar) Reset()                {}
func (s *stubRegistrar) SetActiveTheme(string) {}
func (s *stubRegistrar) ActiveThemeID() string { return "" }
func (s *stubRegistrar) RemovePlugin(string)   {}

func (s *stubRegistrar) AddJSHandler(_, method, route string, _ gin.HandlerFunc) {
	s.handlers[method+" "+route] = true
}

func (s *stubRegistrar) AddStaticFile(_, route, vfsFile string, _ fs.FS) {
	s.static[route] = true
}

func (s *stubRegistrar) AddStaticDir(_, route, vfsDir string, _ fs.FS) {
	s.dirs[route] = vfsDir
}

// newExampleRuntime поднимает рантайм на реальных файлах плагина примера.
func newExampleRuntime(t *testing.T) (*JSRuntime, *stubRegistrar, *Manifest, *I18nRegistry, fs.FS) {
	t.Helper()

	root := filepath.Join("..", "..", "plugs", "example-plugin")
	if _, err := os.Stat(root); err != nil {
		t.Skipf("плагин примера недоступен: %v", err)
	}

	vfs := os.DirFS(root)
	manifest, err := LoadManifestFromFS(vfs)
	if err != nil {
		t.Fatalf("манифест примера не читается: %v", err)
	}

	db, err := database.Open(filepath.Join(t.TempDir(), "example.db"))
	if err != nil {
		t.Fatalf("база не открылась: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userSvc := user.NewService(user.NewStore(db), config.DefaultConfig())
	i18nReg := NewI18nRegistry()
	reg := newStubRegistrar()

	// Модули транскодинга и torrfs не линкуются: пример обязан работать без них
	rt := NewJSRuntime(manifest.ID, manifest, vfs, reg, db, nil, userSvc, i18nReg, nil, nil, nil, nil)

	code, err := fs.ReadFile(vfs, manifest.Entry)
	if err != nil {
		t.Fatalf("entry %q не читается: %v", manifest.Entry, err)
	}
	if err := rt.Execute(string(code)); err != nil {
		t.Fatalf("service.js примера не выполнился: %v", err)
	}
	return rt, reg, manifest, i18nReg, vfs
}

// TestExamplePluginServiceLoads проверяет, что service.js примера выполняется
// на настоящем рантайме без ошибок.
func TestExamplePluginServiceLoads(t *testing.T) {
	newExampleRuntime(t)
}

// TestExamplePluginRoutesRegistered проверяет, что все объявленные в примере
// API-маршруты действительно зарегистрированы и разрешены манифестом.
func TestExamplePluginRoutesRegistered(t *testing.T) {
	_, reg, manifest, _, _ := newExampleRuntime(t)

	if len(reg.handlers) == 0 {
		t.Fatal("ни один обработчик не зарегистрирован")
	}

	// Каждый зарегистрированный маршрут должен быть разрешён манифестом
	for key := range reg.handlers {
		_, route, ok := strings.Cut(key, " ")
		if !ok {
			t.Fatalf("не разобрать ключ маршрута %q", key)
		}
		if !routeAllowed(manifest, route) {
			t.Errorf("маршрут %q зарегистрирован, но не объявлен в info.yaml", route)
		}
	}

	// Ключевые маршруты новых модулей на месте
	for _, want := range []string{
		"POST /api/vfs/read", "POST /api/vfs/list", "POST /api/vfs/stat", "POST /api/vfs/exists",
		"GET /api/transcode/module-info", "GET /api/transcode/profiles", "POST /api/transcode/probe",
		"POST /api/torrent/set-viewed", "POST /api/torrent/tracker-policy", "POST /api/torrent/blocklist",
	} {
		if !reg.handlers[want] {
			t.Errorf("маршрут %q не зарегистрирован", want)
		}
	}

	// Статика примера смонтирована. staticDir дописывает /* к маршруту.
	for _, want := range []string{"/css/*", "/js/*", "/img/*"} {
		if _, ok := reg.dirs[want]; !ok {
			t.Errorf("статический каталог %q не зарегистрирован", want)
		}
	}
}

// TestExamplePluginI18nKeys проверяет, что словари примера читаются,
// не расходятся между языками и содержат ключи новых разделов.
func TestExamplePluginI18nKeys(t *testing.T) {
	_, _, _, _, vfs := newExampleRuntime(t)

	dicts := make(map[string]map[string]string, 2)
	for _, lang := range []string{"en", "ru"} {
		raw, err := fs.ReadFile(vfs, "i18n/"+lang+".json")
		if err != nil {
			t.Fatalf("нет словаря i18n/%s.json: %v", lang, err)
		}

		var dict map[string]string
		if err := json.Unmarshal(raw, &dict); err != nil {
			t.Fatalf("словарь i18n/%s.json не разбирается: %v", lang, err)
		}
		if len(dict) == 0 {
			t.Fatalf("словарь i18n/%s.json пуст", lang)
		}
		dicts[lang] = dict
	}

	// Наборы ключей должны совпадать: иначе часть строк останется без перевода
	for key := range dicts["en"] {
		if _, ok := dicts["ru"][key]; !ok {
			t.Errorf("ключ %q есть в en.json, но отсутствует в ru.json", key)
		}
	}
	for key := range dicts["ru"] {
		if _, ok := dicts["en"][key]; !ok {
			t.Errorf("ключ %q есть в ru.json, но отсутствует в en.json", key)
		}
		if strings.TrimSpace(dicts["ru"][key]) == "" {
			t.Errorf("ключ %q пуст в ru.json", key)
		}
	}

	// Ключи разделов, добавленных вместе с новыми модулями
	for _, key := range []string{
		"nav_vfs", "nav_transcode",
		"sec_vfs_title", "sec_vfs_desc",
		"sec_transcode_title", "sec_transcode_desc",
		"btn_transcode_info", "btn_transcode_probe",
		"btn_torrent_viewed", "btn_torrent_policy", "btn_torrent_blocklist",
	} {
		if _, ok := dicts["en"][key]; !ok {
			t.Errorf("ключ %q отсутствует в en.json", key)
		}
	}
}

// TestExamplePluginDumpRouteInventory печатает фактический набор маршрутов.
// Служит отладке: показывает, что именно зарегистрировал плагин.
func TestExamplePluginDumpRouteInventory(t *testing.T) {
	if testing.Short() {
		t.Skip("инвентаризация пропущена в коротком режиме")
	}
	_, reg, _, _, _ := newExampleRuntime(t)
	t.Logf("обработчиков: %d", len(reg.handlers))
	for k := range reg.handlers {
		t.Logf("  handler %s", k)
	}
	for k := range reg.static {
		t.Logf("  static  %s", k)
	}
	for k, v := range reg.dirs {
		t.Logf("  dir     %s -> %s", k, v)
	}
}
