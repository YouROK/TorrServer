package plugin

import (
	"fmt"
	"io/fs"
	"regexp"

	"gopkg.in/yaml.v3"
)

// pluginIDRe задаёт допустимые идентификаторы плагинов: буквы, цифры, точка,
// дефис, подчёркивание. ID попадает в пути на диске, поэтому разделители
// пути и ".." в него не пропускаем.
var pluginIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// ValidPluginID проверяет, что идентификатор пригоден для имени каталога.
func ValidPluginID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	return pluginIDRe.MatchString(id)
}

// MenuEntry - пункт меню, который плагин хочет добавить в веб-интерфейс
type MenuEntry struct {
	Title    string `yaml:"title" json:"title"`                   // Текст по умолчанию (fallback)
	TitleKey string `yaml:"title_key" json:"title_key,omitempty"` // Опциональный ключ i18n
	Route    string `yaml:"route" json:"route"`                   // Собственный роут плагина
	Rank     int    `yaml:"rank" json:"rank,omitempty"`           // 0 или отсутствие = видно всем
	Icon     string `yaml:"icon" json:"icon,omitempty"`           // Роут на иконку внутри статики плагина
}

type Manifest struct {
	ID          string      `yaml:"id" json:"id"`
	Name        string      `yaml:"name" json:"name"`
	Version     string      `yaml:"version" json:"version"`
	Author      string      `yaml:"author" json:"author"`
	Description string      `yaml:"description" json:"description"`
	Entry       string      `yaml:"entry" json:"entry"`
	ThemeUI     bool        `yaml:"theme_ui" json:"theme_ui"`
	Icon        string      `yaml:"icon" json:"icon,omitempty"` // Роут на иконку плагина (например /img/logo.png)
	Menu        []MenuEntry `yaml:"menu" json:"menu,omitempty"` // Пункты меню для веб-интерфейса
	Events      []string    `yaml:"events" json:"events"`       // Разрешенные внешние топики шины
	Routes      []string    `yaml:"routes" json:"routes"`       // Разрешенные веб-маршруты
}

func LoadManifestFromFS(vfs fs.FS) (*Manifest, error) {
	data, err := fs.ReadFile(vfs, "info.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read info.yaml: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse info.yaml: %w", err)
	}

	if m.ID == "" {
		return nil, fmt.Errorf("plugin manifest missing 'id' field")
	}
	if !ValidPluginID(m.ID) {
		return nil, fmt.Errorf("plugin manifest has invalid 'id' %q: allowed letters, digits, dot, dash and underscore", m.ID)
	}
	if m.Name == "" {
		return nil, fmt.Errorf("plugin manifest missing 'name' field")
	}

	return &m, nil
}
