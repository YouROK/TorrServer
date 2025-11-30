package torrstor

import (
	"errors"

	"github.com/anacrolix/torrent/storage"
)

type PieceFake struct{}

func (PieceFake) ReadAt(p []byte, off int64) (n int, err error) {
	err = errors.New("can't read fake piece")
	return
}

func (PieceFake) WriteAt(p []byte, off int64) (n int, err error) {
	err = errors.New("can't write fake piece")
	return
}

func (PieceFake) MarkComplete() error {
	return errors.New("can't mark complete fake piece")
}

func (PieceFake) MarkNotComplete() error {
	return errors.New("can't mark not complete fake piece")
}

func (PieceFake) Completion() storage.Completion {
	return storage.Completion{
		Complete: false,
		Ok:       true,
	}
}
