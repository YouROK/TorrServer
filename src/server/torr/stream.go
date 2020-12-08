package torr

import (
	"errors"
	"log"
	"net/http"
	"time"

	sets "server/settings"
)

func (t *Torrent) Stream(fileIndex int, req *http.Request, resp http.ResponseWriter) error {
	t.WaitInfo()
	files := t.Files()
	if fileIndex < 1 || fileIndex > len(files) {
		return errors.New("file index out of range")
	}
	file := files[fileIndex-1]
	reader := t.NewReader(file)

	log.Println("Connect client")

	sets.SetViewed(&sets.Viewed{t.Hash().HexString(), fileIndex})

	//TODO проверить почему плеер постоянно переподключается
	resp.Header().Set("Connection", "close")
	//resp.Header().Set("ETag", httptoo.EncodeQuotedString(fmt.Sprintf("%s/%s", t.Hash().HexString(), file.Path())))

	http.ServeContent(resp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)

	t.CloseReader(reader)
	log.Println("Disconnect client")
	return nil
}
