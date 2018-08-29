package torr

import (
	"github.com/anacrolix/dht"
)

type BTState struct {
	LocalPort int
	PeerID    string
	BannedIPs int
	DHTs      []*dht.Server

	Torrents []*Torrent
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
