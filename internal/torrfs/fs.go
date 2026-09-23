package torrfs

import (
	"io"
	"path"
	"strings"

	"silo/internal/torrent"
	"silo/internal/user"
)

type TorrFS struct {
	users *user.Service
	store *torrent.Store
	mgr   *torrent.Manager
}

func New(users *user.Service, store *torrent.Store, mgr *torrent.Manager) *TorrFS {
	return &TorrFS{
		users: users,
		store: store,
		mgr:   mgr,
	}
}

func (t *TorrFS) Root(u *user.User) *RootNode {
	return &RootNode{fs: t, user: u}
}

// Stat возвращает узел по пути.
func (t *TorrFS) Stat(u *user.User, p string) (Node, error) {
	return t.resolve(u, p)
}

// List возвращает содержимое директории.
func (t *TorrFS) List(u *user.User, p string) ([]Node, error) {
	n, err := t.resolve(u, p)
	if err != nil {
		return nil, err
	}
	if !n.IsDir() {
		return nil, ErrNotDir
	}
	return n.Children()
}

// Open открывает файл на чтение (io.ReadSeekCloser).
func (t *TorrFS) Open(u *user.User, p string) (io.ReadSeekCloser, error) {
	n, err := t.resolve(u, p)
	if err != nil {
		return nil, err
	}
	if n.IsDir() {
		return nil, ErrIsDir
	}
	return n.Open()
}

func (t *TorrFS) resolve(u *user.User, p string) (Node, error) {
	clean := cleanPath(p)
	var cur Node = t.Root(u)
	if clean == "" {
		return cur, nil
	}

	for _, part := range strings.Split(clean, "/") {
		if !cur.IsDir() {
			return nil, ErrNotDir
		}
		children, err := cur.Children()
		if err != nil {
			return nil, err
		}
		var next Node
		for _, c := range children {
			if c.Name() == part {
				next = c
				break
			}
		}
		if next == nil {
			return nil, ErrNotFound
		}
		cur = next
	}
	return cur, nil
}

func cleanPath(p string) string {
	p = path.Clean(strings.TrimSpace(p))
	if p == "." || p == "/" {
		return ""
	}
	return strings.TrimPrefix(p, "/")
}

// OpenFile открывает файл по пути и возвращает Handle с метаданными.
func (t *TorrFS) OpenFile(u *user.User, p string) (*Handle, error) {
	n, err := t.resolve(u, p)
	if err != nil {
		return nil, err
	}
	if n.IsDir() {
		return nil, ErrIsDir
	}
	r, err := n.Open()
	if err != nil {
		return nil, err
	}
	return NewHandle(r, n.Name(), n.ModTime()), nil
}
