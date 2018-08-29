package server

import (
	"server/torr"
	"server/web/helpers"
)

type TorrentStat struct {
	Name string
	Hash string

	TorrentStatus       int
	TorrentStatusString string

	LoadedSize  int64
	TorrentSize int64

	PreloadedBytes int64
	PreloadSize    int64

	DownloadSpeed float64
	UploadSpeed   float64

	TotalPeers       int
	PendingPeers     int
	ActivePeers      int
	ConnectedSeeders int

	FileStats []FileStat
}

type FileStat struct {
	Id     int
	Path   string
	Length int64
}

func getTorPlayState(tor *torr.Torrent) TorrentStat {
	tst := tor.Stats()
	ts := TorrentStat{}
	ts.Name = tst.Name
	ts.Hash = tst.Hash
	ts.TorrentStatus = int(tst.TorrentStatus)
	ts.TorrentStatusString = tst.TorrentStatusString
	ts.LoadedSize = tst.LoadedSize
	ts.TorrentSize = tst.TorrentSize
	ts.PreloadedBytes = tst.PreloadedBytes
	ts.PreloadSize = tst.PreloadSize
	ts.DownloadSpeed = tst.DownloadSpeed
	ts.UploadSpeed = tst.UploadSpeed
	ts.TotalPeers = tst.TotalPeers
	ts.PendingPeers = tst.PendingPeers
	ts.ActivePeers = tst.ActivePeers
	ts.ConnectedSeeders = tst.ConnectedSeeders

	files := helpers.GetPlayableFiles(tst)
	ts.FileStats = make([]FileStat, len(files))
	for i, f := range files {
		ts.FileStats[i] = FileStat{
			Id:     f.Id,
			Path:   f.Path,
			Length: f.Length,
		}
	}

	return ts
}
