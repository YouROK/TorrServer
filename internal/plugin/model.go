package plugin

import (
	"io/fs"
	"time"
)

// PluginState сохраняется в базе данных bbolt (BucketPlugins)
// Этим управляет исключительно Owner через админку
type PluginState struct {
	ID        string    `json:"id"`      // Системный ID (имя папки плагина)
	Order     int       `json:"order"`   // Порядковый номер в очереди (1, 2, 3...)
	Enabled   bool      `json:"enabled"` // Включен ли плагин
	CreatedAt time.Time `json:"created_at"`
}

// Plugin представляет загруженный в память плагин
type Plugin struct {
	ID       string
	Manifest *Manifest
	VFS      fs.FS // Виртуальная файловая система (песочница)
	State    PluginState
}
