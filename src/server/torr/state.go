package torr

import (
	"fmt"
	"sort"

	"github.com/anacrolix/torrent"
)

type BTState struct {
	LocalPort int
	PeerID    string
	BannedIPs int
	DHTs      []torrent.DhtServer

	Torrents []*Torrent
}

func (bt *BTServer) BTState() *BTState {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	btState := new(BTState)
	btState.LocalPort = bt.client.LocalPort()
	btState.PeerID = fmt.Sprintf("%x", bt.client.PeerID())
	btState.BannedIPs = len(bt.client.BadPeerIPs())
	btState.DHTs = bt.client.DhtServers()

	for _, t := range bt.torrents {
		btState.Torrents = append(btState.Torrents, t)
	}
	return btState
}

type TorrentStats struct {
	Name string
	Hash string

	TorrentStatus       TorrentStatus
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
	HalfOpenPeers    int

	BytesWritten        int64
	BytesWrittenData    int64
	BytesRead           int64
	BytesReadData       int64
	BytesReadUsefulData int64
	ChunksWritten       int64
	ChunksRead          int64
	ChunksReadUseful    int64
	ChunksReadWasted    int64
	PiecesDirtiedGood   int64
	PiecesDirtiedBad    int64

	FileStats []TorrentFileStat
}

type TorrentFileStat struct {
	Id     int
	Path   string
	Length int64
}

func (t *Torrent) Stats() TorrentStats {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()

	st := TorrentStats{}

	st.Name = t.Name()
	st.Hash = t.hash.HexString()
	st.TorrentStatus = t.status
	st.TorrentStatusString = t.status.String()

	if t.Torrent != nil {
		st.LoadedSize = t.Torrent.BytesCompleted()
		st.TorrentSize = t.Length()
		st.PreloadedBytes = t.PreloadedBytes
		st.PreloadSize = t.PreloadSize
		st.DownloadSpeed = t.DownloadSpeed
		st.UploadSpeed = t.UploadSpeed

		tst := t.Torrent.Stats()
		st.BytesWritten = tst.BytesWritten.Int64()
		st.BytesWrittenData = tst.BytesWrittenData.Int64()
		st.BytesRead = tst.BytesRead.Int64()
		st.BytesReadData = tst.BytesReadData.Int64()
		st.BytesReadUsefulData = tst.BytesReadUsefulData.Int64()
		st.ChunksWritten = tst.ChunksWritten.Int64()
		st.ChunksRead = tst.ChunksRead.Int64()
		st.ChunksReadUseful = tst.ChunksReadUseful.Int64()
		st.ChunksReadWasted = tst.ChunksReadWasted.Int64()
		st.PiecesDirtiedGood = tst.PiecesDirtiedGood.Int64()
		st.PiecesDirtiedBad = tst.PiecesDirtiedBad.Int64()
		st.TotalPeers = tst.TotalPeers
		st.PendingPeers = tst.PendingPeers
		st.ActivePeers = tst.ActivePeers
		st.ConnectedSeeders = tst.ConnectedSeeders
		st.HalfOpenPeers = tst.HalfOpenPeers

		files := t.Files()

		sort.Slice(files, func(i, j int) bool {
			return files[i].Path() < files[j].Path()
		})

		for i, f := range files {
			st.FileStats = append(st.FileStats, TorrentFileStat{
				Id:     i,
				Path:   f.Path(),
				Length: f.Length(),
			})
		}
	}
	return st
}
