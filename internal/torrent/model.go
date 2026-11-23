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

// TorrentStatus - оперативная статистика раздачи из памяти
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
	ViewedFiles         []int              `json:"viewed_files,omitempty"`
}

// TorrentRecord - глобальная физическая карточка раздачи в базе (без личных данных!)
type TorrentRecord struct {
	Hash      string             `json:"hash"`
	Name      string             `json:"name,omitempty"` // Имя раздачи из метаданных
	Size      int64              `json:"size"`
	Timestamp int64              `json:"timestamp"`
	MagnetUri string             `json:"magnet_uri,omitempty"`
	InfoBytes []byte             `json:"info_bytes,omitempty"`
	Trackers  []string           `json:"trackers,omitempty"`
	Files     []*TorrentFileStat `json:"files,omitempty"`
}

// MergeTorrentRecords дополняет старую запись новой: пустые поля не затирают
// старое, трекеры объединяются без дублей.
func MergeTorrentRecords(old, in *TorrentRecord) *TorrentRecord {
	out := *old
	out.Hash = in.Hash

	if in.Name != "" {
		out.Name = in.Name
	}
	if in.Size != 0 {
		out.Size = in.Size
	}
	if in.Timestamp != 0 {
		out.Timestamp = in.Timestamp
	}
	if in.MagnetUri != "" {
		out.MagnetUri = in.MagnetUri
	}
	if len(in.InfoBytes) > 0 {
		out.InfoBytes = in.InfoBytes
	}
	if len(in.Files) > 0 {
		out.Files = in.Files
	}

	out.Trackers = mergeTrackerLists(old.Trackers, in.Trackers)
	return &out
}

// mergeTrackerLists объединяет списки трекеров без дублей, сохраняя порядок.
func mergeTrackerLists(old, in []string) []string {
	seen := make(map[string]bool, len(old)+len(in))
	merged := make([]string, 0, len(old)+len(in))

	for _, list := range [][]string{old, in} {
		for _, tr := range list {
			if tr == "" || seen[tr] {
				continue
			}
			seen[tr] = true
			merged = append(merged, tr)
		}
	}
	return merged
}

// ResolveTitle выбирает название раздачи: личное, затем имя из метаданных, затем техническое.
// Пустой hash отключает техническую подстановку: название остается пустым.
func ResolveTitle(title, metaName, hash string) string {
	if title != "" {
		return title
	}
	if metaName != "" {
		return metaName
	}
	if hash == "" {
		return ""
	}
	if len(hash) >= 8 {
		return "Torrent " + hash[:8]
	}
	return "Torrent"
}

// EphemeralMeta - личные метаданные пользователя для временных торрентов в RAM
type EphemeralMeta struct {
	Title    string
	Poster   string
	Category string
}
