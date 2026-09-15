package torrent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
)

// TorrentStat отражает текущее состояние сессии в памяти
type TorrentStat int

const (
	TorrentAdded TorrentStat = iota
	TorrentGettingInfo
	TorrentPreload
	TorrentWorking
	TorrentClosed
	TorrentInDB
)

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

// TorrentFileStat описывает файл внутри раздачи
type TorrentFileStat struct {
	Id       int    `json:"id"`
	Path     string `json:"path"`      // Полный путь внутри торрента
	Name     string `json:"name"`      // Чистое имя файла (например, "S01E01.mkv")
	Length   int64  `json:"length"`    // Размер в байтах
	FileHash string `json:"file_hash"` // Уникальный хэш по имени и размеру
}

// GenerateFileHash создает хэш файла по его базовому имени и длине в байтах
func GenerateFileHash(path string, length int64) string {
	filename := filepath.Base(path)
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d", filename, length)))
	return hex.EncodeToString(h.Sum(nil))
}

// TorrentStatus — оперативная статистика раздачи из памяти
type TorrentStatus struct {
	Title               string             `json:"title"`
	Name                string             `json:"name,omitempty"`
	Hash                string             `json:"hash"`
	Torrs               string             `json:"torrs,omitempty"`
	Category            string             `json:"category,omitempty"`
	Poster              string             `json:"poster,omitempty"`
	Data                string             `json:"data,omitempty"`
	Timestamp           int64              `json:"timestamp"`
	Stat                TorrentStat        `json:"stat"`
	StatString          string             `json:"stat_string"`
	LoadedSize          int64              `json:"loaded_size"`
	TorrentSize         int64              `json:"torrent_size"`
	PreloadedBytes      int64              `json:"preloaded_bytes"`
	PreloadSize         int64              `json:"preload_size"`
	DownloadSpeed       float64            `json:"download_speed"`
	UploadSpeed         float64            `json:"upload_speed"`
	TotalPeers          int                `json:"total_peers"`
	PendingPeers        int                `json:"pending_peers"`
	ActivePeers         int                `json:"active_peers"`
	ConnectedSeeders    int                `json:"connected_seeders"`
	HalfOpenPeers       int                `json:"half_open_peers"`
	BytesWritten        int64              `json:"bytes_written"`
	BytesRead           int64              `json:"bytes_read"`
	BytesReadUsefulData int64              `json:"bytes_read_useful_data"`
	FileStats           []*TorrentFileStat `json:"file_stats,omitempty"`
}

// TorrentRecord — глобальная физическая карточка раздачи в базе (без личных данных!)
type TorrentRecord struct {
	Hash      string             `json:"hash"`
	Size      int64              `json:"size"`
	Timestamp int64              `json:"timestamp"`
	MagnetUri string             `json:"magnet_uri,omitempty"`
	InfoBytes []byte             `json:"info_bytes,omitempty"`
	Trackers  []string           `json:"trackers,omitempty"`
	Files     []*TorrentFileStat `json:"files,omitempty"`
}

// EphemeralMeta — личные метаданные пользователя для временных торрентов в RAM
type EphemeralMeta struct {
	Title    string
	Poster   string
	Category string
}
