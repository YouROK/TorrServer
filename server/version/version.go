package version

import (
	"log"
	"runtime/debug"
)

// Version is set at build time via -ldflags "-X server/version.Version=<tag>"
var Version = "MatriX.141.2"

func GetTorrentVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Printf("Failed to read build info")
		return ""
	}
	for _, dep := range bi.Deps {
		if dep.Path == "github.com/anacrolix/torrent" {
			if dep.Replace != nil {
				return dep.Replace.Version
			}

			return dep.Version
		}
	}
	return ""
}
