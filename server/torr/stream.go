package torr

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/anacrolix/dms/dlna"
	"github.com/anacrolix/missinggo/v2/httptoo"
	"github.com/anacrolix/torrent"

	mt "server/mimetype"
	sets "server/settings"
	"server/torr/state"
)

func (t *Torrent) Stream(fileID int, req *http.Request, resp http.ResponseWriter) error {
	if !t.GotInfo() {
		http.NotFound(resp, req)
		return errors.New("torrent don't get info")
	}

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
	if int64(sets.MaxSize) > 0 && file.Length() > int64(sets.MaxSize) {
		log.Println("file", file.DisplayPath(), "size exceeded max allowed", sets.MaxSize, "bytes")
		return fmt.Errorf("file size exceeded max allowed %d bytes", sets.MaxSize)
	}

	torrDir := sets.DownloadDir
	localFilePath := filepath.Join(torrDir, file.Path())

	if _, err := os.Stat(localFilePath); err == nil {
		// File exists, serve it directly
		if sets.BTsets.EnableDebug {
			log.Printf("Serving local file: %s\n", localFilePath)
		}

		// Set appropriate headers
		resp.Header().Set("Connection", "close")
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

		// Serve the local file
		http.ServeFile(resp, req, localFilePath)
		return nil
	}

	// If the file doesn't exist locally, continue with the existing torrent streaming logic
	reader := t.NewReader(file)
	if sets.BTsets.ResponsiveMode {
		reader.SetResponsive()
	}

	host, port, err := net.SplitHostPort(req.RemoteAddr)
	if sets.BTsets.EnableDebug {
		if err != nil {
			log.Println("Connect client")
		} else {
			log.Println("Connect client", host, port)
		}
	}

	sets.SetViewed(&sets.Viewed{Hash: t.Hash().HexString(), FileIndex: fileID})

	resp.Header().Set("Connection", "close")
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

	http.ServeContent(resp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)

	t.CloseReader(reader)
	if sets.BTsets.EnableDebug {
		if err != nil {
			log.Println("Disconnect client")
		} else {
			log.Println("Disconnect client", host, port)
		}
	}
	return nil
}
