package state

type TorrentStat int

func (t TorrentStat) String() string {
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
	TorrentAdded = TorrentStat(iota)
	TorrentGettingInfo
	TorrentPreload
	TorrentWorking
	TorrentClosed
	TorrentInDB
)

type TorrentStatus struct {
	Title               string      `json:"title"`
	Category            string      `json:"category"`
	Poster              string      `json:"poster"`
	Data                string      `json:"data,omitempty"`
	Timestamp           int64       `json:"timestamp"`
	Name                string      `json:"name,omitempty"`
	Hash                string      `json:"hash,omitempty"`
	TorrsHash           string      `json:"torrs_hash,omitempty"`
	Stat                TorrentStat `json:"stat"`
	StatString          string      `json:"stat_string"`
	LoadedSize          int64       `json:"loaded_size,omitempty"`
	TorrentSize         int64       `json:"torrent_size,omitempty"`
	PreloadedBytes      int64       `json:"preloaded_bytes,omitempty"`
	PreloadSize         int64       `json:"preload_size,omitempty"`
	DownloadSpeed       float64     `json:"download_speed,omitempty"`
	UploadSpeed         float64     `json:"upload_speed,omitempty"`
	TotalPeers          int         `json:"total_peers,omitempty"`
	PendingPeers        int         `json:"pending_peers,omitempty"`
	ActivePeers         int         `json:"active_peers,omitempty"`
	ConnectedSeeders    int         `json:"connected_seeders,omitempty"`
	HalfOpenPeers       int         `json:"half_open_peers,omitempty"`
	BytesWritten        int64       `json:"bytes_written,omitempty"`
	BytesWrittenData    int64       `json:"bytes_written_data,omitempty"`
	BytesRead           int64       `json:"bytes_read,omitempty"`
	BytesReadData       int64       `json:"bytes_read_data,omitempty"`
	BytesReadUsefulData int64       `json:"bytes_read_useful_data,omitempty"`
	ChunksWritten       int64       `json:"chunks_written,omitempty"`
	ChunksRead          int64       `json:"chunks_read,omitempty"`
	ChunksReadUseful    int64       `json:"chunks_read_useful,omitempty"`
	ChunksReadWasted    int64       `json:"chunks_read_wasted,omitempty"`
	PiecesDirtiedGood   int64       `json:"pieces_dirtied_good,omitempty"`
	PiecesDirtiedBad    int64       `json:"pieces_dirtied_bad,omitempty"`
	DurationSeconds     float64     `json:"duration_seconds,omitempty"`
	BitRate             string      `json:"bit_rate,omitempty"`

	FileStats []*TorrentFileStat `json:"file_stats,omitempty"`
	Playback  []*PlaybackStatus  `json:"playback,omitempty"`
}

// PlaybackStatus is what a client streaming right now is showing. The read head is ahead of
// the picture by whatever the client keeps buffered, so the two are reported side by side:
// Buffer is what was subtracted and BufferMeasured says whether the session produced that
// number itself or it came from the configured fallback.
type PlaybackStatus struct {
	FileIndex      int   `json:"file_index"`
	Head           int64 `json:"head"`
	Buffer         int64 `json:"buffer"`
	BufferMeasured bool  `json:"buffer_measured"`
	// BufferSeconds is the same buffer in film time, which is what it is measured in when the
	// container carries timestamps. Zero when it had to be inferred from a bitrate instead.
	BufferSeconds  float64 `json:"buffer_seconds,omitempty"`
	SessionSeconds float64 `json:"session_seconds,omitempty"`
	// Viewing says this connection has passed the same gates the saver uses: it has streamed
	// long enough and shown something, so it is a viewer rather than a probe or a preload.
	Viewing  bool    `json:"viewing"`
	Position int64   `json:"position"`
	TimeCode float64 `json:"timecode,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	// Source is the container the time was read out of, or "estimate" when the file carries
	// no timestamps and the average bitrate had to be used.
	Source string `json:"source,omitempty"`
}

type TorrentFileStat struct {
	Id     int    `json:"id,omitempty"`
	Path   string `json:"path,omitempty"`
	Length int64  `json:"length,omitempty"`
}
