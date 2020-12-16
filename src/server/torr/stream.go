package torr

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anacrolix/missinggo/httptoo"
	sets "server/settings"
)

func (t *Torrent) Stream(fileIndex int, req *http.Request, resp http.ResponseWriter) error {
	if !t.GotInfo() {
		http.NotFound(resp, req)
		return errors.New("torrent don't get info")
	}
	files := t.Files()
	if fileIndex < 1 || fileIndex > len(files) {
		return errors.New("file index out of range")
	}
	file := files[fileIndex-1]
	reader := t.NewReader(file)

	log.Println("Connect client")

	sets.SetViewed(&sets.Viewed{t.Hash().HexString(), fileIndex})

	resp.Header().Set("Connection", "close")
	resp.Header().Set("ETag", httptoo.EncodeQuotedString(fmt.Sprintf("%s/%s", t.Hash().HexString(), file.Path())))

	http.ServeContent(resp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)

	t.CloseReader(reader)
	log.Println("Disconnect client")
	return nil
}
