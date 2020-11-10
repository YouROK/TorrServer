package state

type TorrentStatus int

func (t TorrentStatus) String() string {
	switch t {
	case TorrentAdded:
		return "Torrent added"
	case TorrentGettingInfo:
		return "Torrent getting info"
	case TorrentPreload:
		return "Torrent preload"
	case TorrentWorking:
		return "Torrent working"
	case TorrentClosed:
		return "Torrent closed"
	case TorrentInDB:
		return "Torrent in db"
	default:
		return "Torrent unknown status"
	}
}

const (
	TorrentAdded = TorrentStatus(iota)
	TorrentGettingInfo
	TorrentPreload
	TorrentWorking
	TorrentClosed
	TorrentInDB
)

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
