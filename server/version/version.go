package version

import (
	"log"
	"runtime/debug"
	// "github.com/anacrolix/torrent"
)

const Version = "MatriX.121.OE"

func GetTorrentVersion() string {
	// _ = torrent.NewDefaultClientConfig()
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Printf("Failed to read build info")
		return ""
	}
	for _, dep := range bi.Deps {
		if dep.Path == "github.com/anacrolix/torrent" || dep.Path == "github.com/tsynik/torrent" {
			return dep.Version
		}
	}
	return ""
}
