package server

import (
	"fmt"
	"net/http"
	"sort"

	"server/utils"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/bytes"
)

func initInfo(e *echo.Echo) {
	server.GET("/cache", cachePage)
	server.GET("/stat", statePage)
	server.GET("/btstat", btStatePage)
}

func btStatePage(c echo.Context) error {
	bts.WriteState(c.Response())
	return c.NoContent(http.StatusOK)
}

func cachePage(c echo.Context) error {
	return c.Render(http.StatusOK, "cachePage", nil)
}

func statePage(c echo.Context) error {
	state := bts.BTState()

	msg := ""

	msg += fmt.Sprintf("Listen port: %d<br>\n", state.LocalPort)
	msg += fmt.Sprintf("Peer ID: %+q<br>\n", state.PeerID)
	msg += fmt.Sprintf("Banned IPs: %d<br>\n", state.BannedIPs)

	for _, dht := range state.DHTs {
		msg += fmt.Sprintf("%s DHT server at %s:<br>\n", dht.Addr().Network(), dht.Addr().String())
		dhtStats := dht.Stats()
		msg += fmt.Sprintf("\t&emsp;# Nodes: %d (%d good, %d banned)<br>\n", dhtStats.Nodes, dhtStats.GoodNodes, dhtStats.BadNodes)
		msg += fmt.Sprintf("\t&emsp;Server ID: %x<br>\n", dht.ID())
		msg += fmt.Sprintf("\t&emsp;Announces: %d<br>\n", dhtStats.SuccessfulOutboundAnnouncePeerQueries)
		msg += fmt.Sprintf("\t&emsp;Outstanding transactions: %d<br>\n", dhtStats.OutstandingTransactions)
	}

	sort.Slice(state.Torrents, func(i, j int) bool {
		return state.Torrents[i].Hash().HexString() < state.Torrents[j].Hash().HexString()
	})
	msg += "Torrents:<br>\n"
	for _, t := range state.Torrents {
		st := t.Stats()
		msg += fmt.Sprintf("Name: %v<br>\n", st.Name)
		msg += fmt.Sprintf("Hash: %v<br>\n", st.Hash)
		msg += fmt.Sprintf("Status: %v<br>\n", st.TorrentStatus)
		msg += fmt.Sprintf("Loaded Size: %v<br>\n", bytes.Format(st.LoadedSize))
		msg += fmt.Sprintf("Torrent Size: %v<br>\n<br>\n", bytes.Format(st.TorrentSize))

		msg += fmt.Sprintf("Preloaded Bytes: %v<br>\n", bytes.Format(st.PreloadedBytes))
		msg += fmt.Sprintf("Preload Size: %v<br>\n<br>\n", bytes.Format(st.PreloadSize))

		msg += fmt.Sprintf("Download Speed: %v/Sec<br>\n", utils.Format(st.DownloadSpeed))
		msg += fmt.Sprintf("Upload Speed: %v/Sec<br>\n<br>\n", utils.Format(st.UploadSpeed))

		msg += fmt.Sprintf("\t&emsp;TotalPeers: %v<br>\n", st.TotalPeers)
		msg += fmt.Sprintf("\t&emsp;PendingPeers: %v<br>\n", st.PendingPeers)
		msg += fmt.Sprintf("\t&emsp;ActivePeers: %v<br>\n", st.ActivePeers)
		msg += fmt.Sprintf("\t&emsp;ConnectedSeeders: %v<br>\n", st.ConnectedSeeders)
		msg += fmt.Sprintf("\t&emsp;HalfOpenPeers: %v<br>\n", st.HalfOpenPeers)

		msg += fmt.Sprintf("\t&emsp;BytesWritten: %v (%v)<br>\n", st.BytesWritten, bytes.Format(st.BytesWritten))
		msg += fmt.Sprintf("\t&emsp;BytesWrittenData: %v (%v)<br>\n", st.BytesWrittenData, bytes.Format(st.BytesWrittenData))
		msg += fmt.Sprintf("\t&emsp;BytesRead: %v (%v)<br>\n", st.BytesRead, bytes.Format(st.BytesRead))
		msg += fmt.Sprintf("\t&emsp;BytesReadData: %v (%v)<br>\n", st.BytesReadData, bytes.Format(st.BytesReadData))
		msg += fmt.Sprintf("\t&emsp;BytesReadUsefulData: %v (%v)<br>\n", st.BytesReadUsefulData, bytes.Format(st.BytesReadUsefulData))
		msg += fmt.Sprintf("\t&emsp;ChunksWritten: %v<br>\n", st.ChunksWritten)
		msg += fmt.Sprintf("\t&emsp;ChunksRead: %v<br>\n", st.ChunksRead)
		msg += fmt.Sprintf("\t&emsp;ChunksReadUseful: %v<br>\n", st.ChunksReadUseful)
		msg += fmt.Sprintf("\t&emsp;ChunksReadWasted: %v<br>\n", st.ChunksReadWasted)
		msg += fmt.Sprintf("\t&emsp;PiecesDirtiedGood: %v<br>\n", st.PiecesDirtiedGood)
		msg += fmt.Sprintf("\t&emsp;PiecesDirtiedBad: %v<br>\n<br>\n", st.PiecesDirtiedBad)
		if len(st.FileStats) > 0 {
			msg += fmt.Sprintf("\t&emsp;Files:<br>\n")
			for _, f := range st.FileStats {
				msg += fmt.Sprintf("\t&emsp;\t&emsp;%v Size:%v<br>\n", f.Path, bytes.Format(f.Length))
			}
		}

		hash := metainfo.NewHashFromHex(st.Hash)
		cState := bts.CacheState(hash)
		if cState != nil {
			msg += fmt.Sprintf("CacheType:<br>\n")
			msg += fmt.Sprintf("Capacity: %v<br>\n", bytes.Format(cState.Capacity))
			msg += fmt.Sprintf("Filled: %v<br>\n", bytes.Format(cState.Filled))
			msg += fmt.Sprintf("PiecesLength: %v<br>\n", bytes.Format(cState.PiecesLength))
			msg += fmt.Sprintf("PiecesCount: %v<br>\n", cState.PiecesCount)
			for _, p := range cState.Pieces {
				msg += fmt.Sprintf("\t&emsp;Piece: %v\t&emsp; Access: %s\t&emsp; Buffer size: %d(%s)\t&emsp; Complete: %v\t&emsp; Hash: %s\n<br>", p.Id, p.Accessed.Format("15:04:05.000"), p.BufferSize, bytes.Format(int64(p.BufferSize)), p.Completed, p.Hash)
			}
		}
		msg += "<hr><br><br>\n\n"
	}
	//msg += `
	//<script>
	//document.addEventListener("DOMContentLoaded", function(event) {
	//	setTimeout(function(){
	//		location.reload();
	//	}, 1000);
	//});
	//</script>
	//
	//`
	return c.HTML(http.StatusOK, msg)
}
