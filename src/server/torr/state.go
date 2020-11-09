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
	Name string `json:"name,omitempty"`
	Hash string `json:"hash,omitempty"`

	TorrentStatus       TorrentStatus `json:"torrent_status,omitempty"`
	TorrentStatusString string        `json:"torrent_status_string,omitempty"`

	LoadedSize  int64 `json:"loaded_size,omitempty"`
	TorrentSize int64 `json:"torrent_size,omitempty"`

	PreloadedBytes int64 `json:"preloaded_bytes,omitempty"`
	PreloadSize    int64 `json:"preload_size,omitempty"`

	DownloadSpeed float64 `json:"download_speed,omitempty"`
	UploadSpeed   float64 `json:"upload_speed,omitempty"`

	TotalPeers       int `json:"total_peers,omitempty"`
	PendingPeers     int `json:"pending_peers,omitempty"`
	ActivePeers      int `json:"active_peers,omitempty"`
	ConnectedSeeders int `json:"connected_seeders,omitempty"`
	HalfOpenPeers    int `json:"half_open_peers,omitempty"`

	BytesWritten        int64 `json:"bytes_written,omitempty"`
	BytesWrittenData    int64 `json:"bytes_written_data,omitempty"`
	BytesRead           int64 `json:"bytes_read,omitempty"`
	BytesReadData       int64 `json:"bytes_read_data,omitempty"`
	BytesReadUsefulData int64 `json:"bytes_read_useful_data,omitempty"`
	ChunksWritten       int64 `json:"chunks_written,omitempty"`
	ChunksRead          int64 `json:"chunks_read,omitempty"`
	ChunksReadUseful    int64 `json:"chunks_read_useful,omitempty"`
	ChunksReadWasted    int64 `json:"chunks_read_wasted,omitempty"`
	PiecesDirtiedGood   int64 `json:"pieces_dirtied_good,omitempty"`
	PiecesDirtiedBad    int64 `json:"pieces_dirtied_bad,omitempty"`

	FileStats []TorrentFileStat `json:"file_stats,omitempty"`
}

type TorrentFileStat struct {
	Id     int    `json:"id,omitempty"`
	Path   string `json:"path,omitempty"`
	Length int64  `json:"length,omitempty"`
}

func (t *Torrent) Stats() TorrentStats {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()

	st := TorrentStats{}

	st.Name = t.Name()
	st.Hash = t.hash.HexString()
	st.TorrentStatus = t.Status
	st.TorrentStatusString = t.Status.String()

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
