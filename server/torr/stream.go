package torr

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anacrolix/missinggo/httptoo"
	"github.com/anacrolix/torrent"

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

	reader := t.NewReader(file)

	log.Println("Connect client")

	sets.SetViewed(&sets.Viewed{t.Hash().HexString(), fileID})

	resp.Header().Set("Connection", "close")
	resp.Header().Set("ETag", httptoo.EncodeQuotedString(fmt.Sprintf("%s/%s", t.Hash().HexString(), file.Path())))

	http.ServeContent(resp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)

	t.CloseReader(reader)
	log.Println("Disconnect client")
	return nil
}
