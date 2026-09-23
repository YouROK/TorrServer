package torrfs

import (
	"sort"
	"strings"
	"time"

	"silo/internal/torrent"
	"silo/internal/user"
)

type treeBuilder struct {
	dirs  map[string]*dirEntry
	files []*torrent.TorrentFileStat
}

type dirEntry struct {
	name     string
	children *treeBuilder
}

func newTreeBuilder() *treeBuilder {
	return &treeBuilder{dirs: map[string]*dirEntry{}}
}

func (b *treeBuilder) insert(path string, file *torrent.TorrentFileStat) {
	path = strings.Trim(path, "/")
	if path == "" {
		return
	}
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		b.files = append(b.files, file)
		return
	}
	name := parts[0]
	sub, ok := b.dirs[name]
	if !ok {
		sub = &dirEntry{name: name, children: newTreeBuilder()}
		b.dirs[name] = sub
	}
	sub.children.insert(parts[1], file)
}

// buildChildrenFromTree собирает список узлов из дерева
func buildChildrenFromTree(
	fs *TorrFS,
	u *user.User,
	hash string,
	mtime time.Time,
	b *treeBuilder,
	parentPath string,
) []Node {
	if b == nil {
		return nil
	}

	dirNames := make([]string, 0, len(b.dirs))
	for name := range b.dirs {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	var nodes []Node

	for _, rawName := range dirNames {
		dir := b.dirs[rawName]
		safeName := sanitizeName(rawName)
		nodes = append(nodes, &DirNode{
			fs:    fs,
			user:  u,
			hash:  hash,
			name:  safeName,
			path:  joinPath(parentPath, safeName),
			mtime: mtime,
			sub:   dir.children,
		})
	}

	files := append([]*torrent.TorrentFileStat(nil), b.files...)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	for _, f := range files {
		safeName := sanitizeName(f.Name)
		nodes = append(nodes, &FileNode{
			fs:      fs,
			user:    u,
			hash:    hash,
			fileIdx: f.Id,
			name:    safeName,
			path:    joinPath(parentPath, safeName),
			size:    f.Length,
			mtime:   mtime,
			mime:    mimeByExtension(f.Name),
		})
	}

	return nodes
}

func joinPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

func shortHash(hash string) string {
	if len(hash) <= 8 {
		return hash
	}
	return hash[:8]
}
