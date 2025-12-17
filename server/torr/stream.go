package torr

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/anacrolix/dms/dlna"
	"github.com/anacrolix/missinggo/v2/httptoo"
	"github.com/anacrolix/torrent"

	mt "server/mimetype"
	sets "server/settings"
	"server/torr/state"
)

// Add atomic counter for concurrent streams
var activeStreams int32

func (t *Torrent) Stream(fileID int, req *http.Request, resp http.ResponseWriter) error {
	// Increment active streams counter
	streamID := atomic.AddInt32(&activeStreams, 1)
	defer atomic.AddInt32(&activeStreams, -1)

	// Stream disconnect timeout
	StreamTimeout := sets.BTsets.TorrentDisconnectTimeout

	if !t.GotInfo() {
		http.NotFound(resp, req)
		return errors.New("torrent doesn't have info yet")
	}

	// Get file information
	st := t.Status()
	var stFile *state.TorrentFileStat
	for _, fileStat := range st.FileStats {
		if fileStat.Id == fileID {
			stFile = fileStat
			break
		}
	}
	if stFile == nil {
		return fmt.Errorf("file with id %v not found", fileID)
	}

	// Find the actual torrent file
	files := t.Files()
	var file *torrent.File
	for _, tfile := range files {
		if tfile.Path() == stFile.Path {
			file = tfile
			break
		}
	}
	if file == nil {
		return fmt.Errorf("file with id %v not found", fileID)
	}

	// Check file size limit
	if int64(sets.MaxSize) > 0 && file.Length() > int64(sets.MaxSize) {
		err := fmt.Errorf("file size exceeded max allowed %d bytes", sets.MaxSize)
		log.Printf("File %s size (%d) exceeded max allowed %d bytes", file.DisplayPath(), file.Length(), sets.MaxSize)
		http.Error(resp, err.Error(), http.StatusForbidden)
		return err
	}

	// Create reader with context for timeout
	reader := t.NewReader(file)
	if reader == nil {
		return errors.New("cannot create torrent reader")
	}

	// Ensure reader is always closed
	defer t.CloseReader(reader)

	if sets.BTsets.ResponsiveMode {
		reader.SetResponsive()
	}

	host, port, hperr := net.SplitHostPort(req.RemoteAddr)

	// Log connection
	if sets.BTsets.EnableDebug {
		if hperr != nil {
			log.Printf("[Stream:%d] Connect client (Active streams: %d)", streamID, atomic.LoadInt32(&activeStreams))
		} else {
			log.Printf("[Stream:%d] Connect client %s:%s (Active streams: %d)",
				streamID, host, port, atomic.LoadInt32(&activeStreams))
		}
	}

	// Mark as viewed
	sets.SetViewed(&sets.Viewed{
		Hash:      t.Hash().HexString(),
		FileIndex: fileID,
	})

	// Set response headers
	resp.Header().Set("Connection", "close")

	// Add timeout header if configured
	if StreamTimeout > 0 {
		resp.Header().Set("X-Stream-Timeout", fmt.Sprintf("%d", StreamTimeout))
	}

	etag := hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", t.Hash().HexString(), file.Path())))
	resp.Header().Set("ETag", httptoo.EncodeQuotedString(etag))

	// DLNA headers
	resp.Header().Set("transferMode.dlna.org", "Streaming")
	mime, err := mt.MimeTypeByPath(file.Path())
	if err == nil && mime.IsMedia() {
		resp.Header().Set("content-type", mime.String())
	}

	if req.Header.Get("getContentFeatures.dlna.org") != "" {
		resp.Header().Set("contentFeatures.dlna.org", dlna.ContentFeatures{
			SupportRange:    true,
			SupportTimeSeek: true,
		}.String())
	}

	// Add support for range requests
	if req.Header.Get("Range") != "" {
		resp.Header().Set("Accept-Ranges", "bytes")
	}

	// Create a context with timeout if configured
	ctx := req.Context()
	if StreamTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(StreamTimeout)*time.Second)
		defer cancel()

		// Update request with new context
		req = req.WithContext(ctx)
	}

	// Create a trackingReadSeeker that implements both io.Reader and io.Seeker
	tracker := &trackingReadSeeker{
		ReadSeeker: reader,
		StreamID:   streamID,
		FileName:   file.DisplayPath(),
	}

	// Use a custom ResponseWriter to handle disconnections
	wrappedWriter := &responseWriter{
		ResponseWriter: resp,
		OnWrite: func(n int) {
			// Track bytes sent if needed
			atomic.AddInt64(&tracker.BytesSent, int64(n))
		},
	}

	// Serve content with our tracking read seeker
	http.ServeContent(wrappedWriter, req, file.Path(), time.Unix(t.Timestamp, 0), tracker)

	// Log disconnection
	if sets.BTsets.EnableDebug {
		if hperr != nil {
			log.Printf("[Stream:%d] Disconnect client (Read: %d, Sent: %d bytes, Duration: %v)",
				streamID, tracker.BytesRead, tracker.BytesSent, time.Since(tracker.StartTime))
		} else {
			log.Printf("[Stream:%d] Disconnect client %s:%s (Read: %d, Sent: %d bytes, Duration: %v)",
				streamID, host, port, tracker.BytesRead, tracker.BytesSent, time.Since(tracker.StartTime))
		}
	}

	return nil
}

// GetActiveStreams returns number of currently active streams
func GetActiveStreams() int32 {
	return atomic.LoadInt32(&activeStreams)
}

// trackingReadSeeker wraps an io.ReadSeeker to track bytes read
type trackingReadSeeker struct {
	io.ReadSeeker
	StreamID  int32
	FileName  string
	BytesRead int64
	BytesSent int64
	StartTime time.Time
}

func (trs *trackingReadSeeker) Read(p []byte) (n int, err error) {
	if trs.StartTime.IsZero() {
		trs.StartTime = time.Now()
	}

	n, err = trs.ReadSeeker.Read(p)
	if n > 0 {
		atomic.AddInt64(&trs.BytesRead, int64(n))

		// Log progress for large reads
		if sets.BTsets.EnableDebug && trs.BytesRead%(10*1024*1024) == 0 {
			log.Printf("[Stream:%d] %s: Read %d MB",
				trs.StreamID, trs.FileName, trs.BytesRead/(1024*1024))
		}
	}
	return
}

func (trs *trackingReadSeeker) Seek(offset int64, whence int) (int64, error) {
	newPos, err := trs.ReadSeeker.Seek(offset, whence)
	if sets.BTsets.EnableDebug && err == nil {
		log.Printf("[Stream:%d] %s: Seek to %d (whence: %d)",
			trs.StreamID, trs.FileName, newPos, whence)
	}
	return newPos, err
}

// Helper struct to handle response writing with callbacks
type responseWriter struct {
	http.ResponseWriter
	OnWrite func(int)
}

func (rw *responseWriter) Write(p []byte) (n int, err error) {
	n, err = rw.ResponseWriter.Write(p)
	if rw.OnWrite != nil && n > 0 {
		rw.OnWrite(n)
	}
	return
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	// You can track status codes here if needed
	rw.ResponseWriter.WriteHeader(statusCode)
}
