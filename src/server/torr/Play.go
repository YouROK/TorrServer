package torr

import (
	"fmt"
	"net/http"
	"time"

	"server/settings"
	"server/utils"

	"github.com/anacrolix/missinggo/httptoo"
	"github.com/anacrolix/torrent"
	"github.com/labstack/echo"
)

func (bt *BTServer) View(torr *Torrent, file *torrent.File, c echo.Context) error {
	go settings.SetViewed(torr.Hash().HexString(), file.Path())
	reader := torr.NewReader(file, 0)

	fmt.Println("Connect reader:", len(torr.readers))
	c.Response().Header().Set("Connection", "close")
	c.Response().Header().Set("ETag", httptoo.EncodeQuotedString(fmt.Sprintf("%s/%s", torr.Hash().HexString(), file.Path())))

	http.ServeContent(c.Response(), c.Request(), file.Path(), time.Time{}, reader)

	fmt.Println("Disconnect reader:", len(torr.readers))
	torr.CloseReader(reader)
	return c.NoContent(http.StatusOK)
}

func (bt *BTServer) Play(torr *Torrent, file *torrent.File, preload int64, c echo.Context) error {
	if torr.status == TorrentAdded {
		if !torr.GotInfo() {
			return echo.NewHTTPError(http.StatusBadRequest, "torrent closed befor get info")
		}
	}
	if torr.status == TorrentGettingInfo {
		if !torr.WaitInfo() {
			return echo.NewHTTPError(http.StatusBadRequest, "torrent closed befor get info")
		}
	}

	if torr.PreloadedBytes == 0 {
		torr.Preload(file, preload)
	}

	redirectUrl := c.Scheme() + "://" + c.Request().Host + "/torrent/view/" + torr.Hash().HexString() + "/" + utils.CleanFName(file.Path())
	return c.Redirect(http.StatusFound, redirectUrl)

	//return bt.View(torr, file, c)
}
