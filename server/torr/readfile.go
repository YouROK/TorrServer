package torr

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/anacrolix/torrent"

	"server/torr/state"
)

// ReadFile reads a small torrent file in full (e.g. subtitles).
// Unlike Stream() it does not mark the file as viewed and does not write HTTP headers.
// Files larger than maxSize are rejected. Reading is aborted when ctx is done.
func (t *Torrent) ReadFile(ctx context.Context, fileID int, maxSize int64) ([]byte, error) {
	if !t.GotInfo() {
		return nil, errors.New("torrent doesn't have info yet")
	}

	// Find the file as Stream() does: first by Id in the status, then by Path among the files.
	var stFile *state.TorrentFileStat
	for _, fileStat := range t.Status().FileStats {
		if fileStat.Id == fileID {
			stFile = fileStat
			break
		}
	}
	if stFile == nil {
		return nil, fmt.Errorf("file with id %v not found", fileID)
	}

	var file *torrent.File
	for _, tfile := range t.Files() {
		if tfile.Path() == stFile.Path {
			file = tfile
			break
		}
	}
	if file == nil {
		return nil, fmt.Errorf("file with id %v not found", fileID)
	}

	if file.Length() > maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", file.Length(), maxSize)
	}

	reader := t.NewReader(file)
	if reader == nil {
		return nil, errors.New("cannot create torrent reader")
	}
	// Closing the reader also unblocks the goroutine below if the context expires.
	defer t.CloseReader(reader)
	reader.SetResponsive()

	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1) // buffer of 1: the goroutine never blocks if we already left on ctx
	size := file.Length()        // already checked: size <= maxSize
	go func() {
		// Read exactly size bytes: the reader does not stop at the end of the file
		// by itself and would return data of the next file in the torrent.
		// (Stream() is protected because ServeContent copies exactly size bytes.)
		buf := make([]byte, size)
		_, err := io.ReadFull(reader, buf)
		done <- result{buf, err}
	}()

	select {
	case r := <-done:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
