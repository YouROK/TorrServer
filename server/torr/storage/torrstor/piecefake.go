package torrstor

import (
	"errors"

	"github.com/anacrolix/torrent/storage"
)

type PieceFake struct{}

func (PieceFake) ReadAt(p []byte, off int64) (n int, err error) {
	// err = errors.New("can't read fake piece")
	return
}

func (PieceFake) WriteAt(p []byte, off int64) (n int, err error) {
	// err = errors.New("can't write fake piece")
	return
}

func (PieceFake) MarkComplete() error {
	return errors.New("can't mark complete fake piece")
}

func (PieceFake) MarkNotComplete() error {
	return errors.New("can't mark not complete fake piece")
}

//	type Completion struct {
//		Err error
//		// The state is known or cached.
//		Ok bool
//		// If Ok, whether the data is correct. TODO: Check all callsites test Ok first.
//		Complete bool
//	}
func (PieceFake) Completion() storage.Completion {
	return storage.Completion{
		Complete: false,
		Err:      errors.New("fake piece"),
		Ok:       false,
	}
}
