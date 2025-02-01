package settings

import (
	"os"
	"path/filepath"

	"server/log"
)

var (
	tdb      TorrServerDB
	Path     string
	IP       string
	Port     string
	Ssl      bool
	SslPort  string
	ReadOnly bool
	HttpAuth bool
	SearchWA bool
	PubIPv4  string
	PubIPv6  string
	TorAddr  string
	MaxSize  int64
)

func InitSets(readOnly, searchWA bool) {
	ReadOnly = readOnly
	SearchWA = searchWA

	bboltDB := NewTDB()
	if bboltDB == nil {
		log.TLogln("Error open bboltDB:", filepath.Join(Path, "config.db"))
		os.Exit(1)
	}

	jsonDB := NewJsonDB()
	if jsonDB == nil {
		log.TLogln("Error open jsonDB")
		os.Exit(1)
	}

	dbRouter := NewXPathDBRouter()
	// First registered DB becomes default route
	dbRouter.RegisterRoute(jsonDB, "Settings")
	dbRouter.RegisterRoute(jsonDB, "Viewed")
	dbRouter.RegisterRoute(bboltDB, "Torrents")

	tdb = NewDBReadCache(dbRouter)

	// We migrate settings here, it must be done before loadBTSets()
	if err := MigrateToJson(bboltDB, jsonDB); err != nil {
		log.TLogln("MigrateToJson failed")
		os.Exit(1)
	}
	loadBTSets()
	MigrateTorrents()
}

func CloseDB() {
	tdb.CloseDB()
}
