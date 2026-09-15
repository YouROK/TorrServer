package torrent

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"golang.org/x/time/rate"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// GeneratePeerID создает стандартизированный ID клиента qBittorrent
func GeneratePeerID(prefix string) string {
	randomBytes := make([]byte, 20)
	_, _ = rand.Read(randomBytes)
	encoded := base32.StdEncoding.EncodeToString(randomBytes)
	return prefix + encoded[:20-len(prefix)]
}

// NewRateLimiter создает ограничитель скорости (КБ/сек в rate.Limiter)
func NewRateLimiter(rateKB int) *rate.Limiter {
	if rateKB <= 0 {
		return rate.NewLimiter(rate.Inf, 0)
	}
	bytesPerSec := rateKB * 1024
	burst := bytesPerSec
	if burst < 16*1024 {
		burst = 16 * 1024
	}
	return rate.NewLimiter(rate.Limit(bytesPerSec), burst)
}

// OpenTorrentFile читает .torrent файл с диска и создает спецификацию
func OpenTorrentFile(filePath string) (*torrent.TorrentSpec, error) {
	minfo, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}

	mag := minfo.Magnet(nil, &info)
	return &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: info.Name,
		InfoHash:    minfo.HashInfoBytes(),
	}, nil
}

// ParseTorrentSpec преобразует magnet-ссылку или 40-символьный hex-хэш в TorrentSpec
func ParseTorrentSpec(input string) (*torrent.TorrentSpec, error) {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "magnet:?") {
		mag, err := metainfo.ParseMagnetUri(input)
		if err != nil {
			return nil, fmt.Errorf("invalid magnet uri: %w", err)
		}
		var tiers [][]string
		for _, tr := range mag.Trackers {
			tiers = append(tiers, []string{tr})
		}
		return &torrent.TorrentSpec{
			InfoHash:    mag.InfoHash,
			Trackers:    tiers,
			DisplayName: mag.DisplayName,
		}, nil
	}
	if len(input) == 40 {
		hash := metainfo.NewHashFromHex(input)
		return &torrent.TorrentSpec{
			InfoHash: hash,
		}, nil
	}
	return nil, fmt.Errorf("invalid torrent input: must be magnet link or 40-char infohash")
}
