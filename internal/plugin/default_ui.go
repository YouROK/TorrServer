package plugin

import (
	"embed"
	"io/fs"
)

//go:embed all:default_ui
var defaultUIFiles embed.FS

// GetDefaultUIFS возвращает виртуальную ФС встроенного дефолтного плагина
func GetDefaultUIFS() (fs.FS, error) {
	return fs.Sub(defaultUIFiles, "default_ui")
}
