package torr

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
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

type contextResponseWriter struct {
	http.ResponseWriter
	ctx context.Context
}

func (w *contextResponseWriter) Write(p []byte) (n int, err error) {
	// Check context before each write
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
		return w.ResponseWriter.Write(p)
	}
}

func (t *Torrent) Stream(fileID int, req *http.Request, resp http.ResponseWriter) error {
	// Increment active streams counter
	streamID := atomic.AddInt32(&activeStreams, 1)
	defer atomic.AddInt32(&activeStreams, -1)
	// Stream disconnect timeout (same as torrent)
	streamTimeout := sets.BTsets.TorrentDisconnectTimeout

	// Check if torrent is closed at the very beginning
	t.muTorrent.Lock()
	if t.Stat == state.TorrentClosed {
		t.muTorrent.Unlock()
		http.NotFound(resp, req)
		return fmt.Errorf("torrent is closed (stream ID: %d)", streamID)
	}
	t.muTorrent.Unlock()

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

	// Helper function to check if torrent is closed
	isTorrentClosed := func() bool {
		t.muTorrent.Lock()
		defer t.muTorrent.Unlock()
		return t.Stat == state.TorrentClosed || t.Torrent == nil
	}

	// Ensure reader is always closed
	defer func() {
		// Check if torrent is still valid before closing
		if !isTorrentClosed() {
			t.CloseReader(reader)
		}
	}()

	if sets.BTsets.ResponsiveMode {
		reader.SetResponsive()
	}

	// Log connection
	host, port, clerr := net.SplitHostPort(req.RemoteAddr)

	if sets.BTsets.EnableDebug {
		if clerr != nil {
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
	if streamTimeout > 0 {
		resp.Header().Set("X-Stream-Timeout", fmt.Sprintf("%d", streamTimeout))
	}
	// Add ETag
	etag := hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", t.Hash().HexString(), file.Path())))
	resp.Header().Set("ETag", httptoo.EncodeQuotedString(etag))
	// DLNA headers
	resp.Header().Set("transferMode.dlna.org", "Streaming")
	// add MimeType
	mime, err := mt.MimeTypeByPath(file.Path())
	if err == nil && mime.IsMedia() {
		resp.Header().Set("content-type", mime.String())
	}
	// DLNA Seek
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
	if streamTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(streamTimeout)*time.Second)
		defer cancel()
	}

	// Update request with new context
	req = req.WithContext(ctx)

	// Handle client disconnections better
	wrappedResp := &contextResponseWriter{
		ResponseWriter: resp,
		ctx:            ctx,
	}

	// Add recovery for ServeContent panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Stream:%d] Recovered from panic in ServeContent: %v", streamID, r)
			http.Error(resp, "Internal server error", http.StatusInternalServerError)
		}
	}()

	// Check if torrent is still valid before starting to serve
	if isTorrentClosed() {
		log.Printf("[Stream:%d] Torrent closed before serving content", streamID)
		http.NotFound(resp, req)
		return fmt.Errorf("torrent closed before serving (stream ID: %d)", streamID)
	}

	// Create a wrapper context that cancels when torrent is closed
	streamCtx, streamCancel := context.WithCancel(ctx)
	defer streamCancel()

	// Monitor torrent status in background
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if isTorrentClosed() {
					streamCancel()
					return
				}
			case <-streamCtx.Done():
				return
			}
		}
	}()

	// Update request with torrent context
	req = req.WithContext(streamCtx)

	http.ServeContent(wrappedResp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)

	if sets.BTsets.EnableDebug {
		if clerr != nil {
			log.Printf("[Stream:%d] Disconnect client", streamID)
		} else {
			log.Printf("[Stream:%d] Disconnect client %s:%s", streamID, host, port)
		}
	}
	return nil
}

// GetActiveStreams returns number of currently active streams
func GetActiveStreams() int32 {
	return atomic.LoadInt32(&activeStreams)
}
