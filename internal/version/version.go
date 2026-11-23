package version

import (
	"log"
	"runtime/debug"
)

// Version - версия релиза, при сборке берётся из тега:
// -ldflags "-X silo/internal/version.Version=<tag>". Без ldflags - "dev".
var Version = "dev"

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
