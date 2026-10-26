package torrstor

import (
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	"server/settings"
)

const metainfoFileName = "info"

func isCacheMetaFile(name string) bool {
	return name == metainfoFileName || name == pinsFileName
}

func metainfoPath(hashHex string) string {
	return filepath.Join(settings.BTsets.TorrentsSavePath, hashHex, metainfoFileName)
}

func ApplyCachedMetainfo(spec *torrent.TorrentSpec) {
	if spec == nil || len(spec.InfoBytes) > 0 {
		return
	}
	if settings.BTsets == nil || !settings.BTsets.UseDisk {
		return
	}
	infoBytes := LoadMetainfo(spec.InfoHash.HexString())
	if len(infoBytes) == 0 {
		return
	}
	spec.InfoBytes = infoBytes
	if spec.DisplayName == "" {
		mi := metainfo.MetaInfo{InfoBytes: infoBytes}
		if info, err := mi.UnmarshalInfo(); err == nil {
			spec.DisplayName = info.Name
		}
	}
}

func SaveMetainfo(hashHex string, infoBytes []byte) {
	if settings.BTsets == nil || !settings.BTsets.UseDisk || len(infoBytes) == 0 {
		return
	}
	path := settings.BTsets.TorrentsSavePath
	if path == "" || path == "/" {
		return
	}
	dir := filepath.Join(path, hashHex)
	if err := os.MkdirAll(dir, 0o777); err != nil {
		log.TLogln("Error create cache dir for metainfo:", err)
		return
	}
	if err := os.WriteFile(metainfoPath(hashHex), infoBytes, 0o666); err != nil {
		log.TLogln("Error save metainfo cache:", err)
	}
}

// CachedTorrents lists the torrents whose metainfo is saved in the cache directory.
func CachedTorrents() []string {
	if settings.BTsets == nil || !settings.BTsets.UseDisk {
		return nil
	}
	path := settings.BTsets.TorrentsSavePath
	if path == "" || path == "/" {
		return nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	var hashes []string
	for _, entry := range entries {
		if !entry.IsDir() || !isInfoHashName(entry.Name()) {
			continue
		}
		if _, err := os.Stat(metainfoPath(entry.Name())); err == nil {
			hashes = append(hashes, entry.Name())
		}
	}
	return hashes
}

func LoadMetainfo(hashHex string) []byte {
	if settings.BTsets == nil || !settings.BTsets.UseDisk {
		return nil
	}
	path := settings.BTsets.TorrentsSavePath
	if path == "" || path == "/" {
		return nil
	}
	data, err := os.ReadFile(metainfoPath(hashHex))
	if err != nil || len(data) == 0 {
		return nil
	}
	return data
}
