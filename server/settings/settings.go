package settings

import (
	"os"
	"path/filepath"

	"server/log"
)

var (
	tdb      *TDB
	Path     string
	Port     string
	ReadOnly bool
	HttpAuth bool
	PubIPv4  string
	PubIPv6  string
)

func InitSets(readOnly bool) {
	ReadOnly = readOnly
	tdb = NewTDB()
	if tdb == nil {
		log.TLogln("Error open db:", filepath.Join(Path, "config.db"))
		os.Exit(1)
	}
	loadBTSets()
	Migrate()
}

func CloseDB() {
	tdb.CloseDB()
}
