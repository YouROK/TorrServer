package torr

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anacrolix/missinggo/httptoo"
)

func (t *Torrent) Stream(fileIndex int, req *http.Request, resp http.ResponseWriter) error {
	files := t.Files()
	if fileIndex < 1 || fileIndex > len(files) {
		return errors.New("file index out of range")
	}
	file := files[fileIndex-1]
	reader := t.NewReader(file, 0)

	log.Println("Connect client")

	resp.Header().Set("Connection", "close")
	resp.Header().Set("ETag", httptoo.EncodeQuotedString(fmt.Sprintf("%s/%s", t.Hash().HexString(), file.Path())))

	http.ServeContent(resp, req, file.Path(), time.Time{}, reader)

	log.Println("Disconnect client")
	t.CloseReader(reader)
	return nil
}
