package plugin

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:default_ui
var defaultUIFiles embed.FS

//go:embed all:admin_ui
var adminUIFiles embed.FS

// BuiltinPlugin — один встроенный плагин: уже открытая VFS и разобранный манифест.
type BuiltinPlugin struct {
	VFS      fs.FS
	Manifest *Manifest
}

// LoadBuiltins собирает все встроенные плагины из embed-ресурсов.
func LoadBuiltins() (map[string]*BuiltinPlugin, []string, error) {
	sources := []struct {
		fsys embed.FS
		root string
	}{
		{defaultUIFiles, "default_ui"},
		{adminUIFiles, "admin_ui"},
	}

	plugins := make(map[string]*BuiltinPlugin, len(sources))
	order := make([]string, 0, len(sources))

	for _, src := range sources {
		sub, err := fs.Sub(src.fsys, src.root)
		if err != nil {
			return nil, nil, fmt.Errorf("builtin '%s': open embedded fs: %w", src.root, err)
		}

		man, err := LoadManifestFromFS(sub)
		if err != nil {
			return nil, nil, fmt.Errorf("builtin '%s': load manifest: %w", src.root, err)
		}

		if _, dup := plugins[man.ID]; dup {
			return nil, nil, fmt.Errorf("builtin '%s': duplicate id '%s'", src.root, man.ID)
		}

		plugins[man.ID] = &BuiltinPlugin{VFS: sub, Manifest: man}
		order = append(order, man.ID)
	}

	return plugins, order, nil
}
