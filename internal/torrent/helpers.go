package torrent

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"os"
	"strings"

	"golang.org/x/time/rate"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
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

// InfoName достает имя раздачи из bencode info-словаря
func InfoName(infoBytes []byte) string {
	if len(infoBytes) == 0 {
		return ""
	}
	var info metainfo.Info
	if err := bencode.Unmarshal(infoBytes, &info); err != nil {
		return ""
	}
	return info.BestName()
}

// ParseTorrentFile разбирает .torrent файл и создает спецификацию
func ParseTorrentFile(data []byte) (*torrent.TorrentSpec, error) {
	minfo, err := metainfo.Load(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid torrent file: %w", err)
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, fmt.Errorf("invalid torrent info: %w", err)
	}

	return &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    minfo.UpvertedAnnounceList(),
		DisplayName: info.BestName(),
		InfoHash:    minfo.HashInfoBytes(),
	}, nil
}

// OpenTorrentFile читает .torrent файл с диска и создает спецификацию
func OpenTorrentFile(filePath string) (*torrent.TorrentSpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return ParseTorrentFile(data)
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
