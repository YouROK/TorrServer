package plugin

import (
	"errors"
	"os"

	"silo/internal/fusefs"
	"silo/internal/torrfs"
	"silo/internal/user"
)

// torrFSAdapter отдаёт дерево торрентов пользователя как источник для монтирования.
type torrFSAdapter struct {
	fs   *torrfs.TorrFS
	user *user.User
}

func (a *torrFSAdapter) Stat(path string) (fusefs.Node, error) {
	n, err := a.fs.Stat(a.user, path)
	if err != nil {
		return nil, wrapFSError(err)
	}
	return n, nil
}

func (a *torrFSAdapter) List(path string) ([]fusefs.Node, error) {
	nodes, err := a.fs.List(a.user, path)
	if err != nil {
		return nil, wrapFSError(err)
	}

	result := make([]fusefs.Node, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, n)
	}
	return result, nil
}

func (a *torrFSAdapter) Open(path string) (fusefs.File, error) {
	h, err := a.fs.OpenFile(a.user, path)
	if err != nil {
		return nil, wrapFSError(err)
	}
	return h, nil
}

// wrapFSError помечает отсутствующие пути, чтобы файловая система
// отвечала ENOENT, а не ошибкой ввода-вывода.
func wrapFSError(err error) error {
	if errors.Is(err, torrfs.ErrNotFound) || errors.Is(err, torrfs.ErrNotDir) {
		return os.ErrNotExist
	}
	return err
}
