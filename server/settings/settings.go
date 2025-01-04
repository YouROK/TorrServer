package settings

import (
	"os"
	"path/filepath"

	"server/log"
)

var (
	tdb      TorrServerDB
	Path     string
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

func InitSets(readOnly, searchWA bool, useLegacyBackend bool) {
	ReadOnly = readOnly
	SearchWA = searchWA

	var SettingsStorage TorrServerDB
	var ViewedStorage TorrServerDB
	var TorrentsStorage TorrServerDB

	if useLegacyBackend {
		log.TLogln("Using legacy storage backends...")

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

		SettingsStorage, ViewedStorage, TorrentsStorage = jsonDB, jsonDB, bboltDB

		// We migrate settings here, it must be done before loadBTSets()
		if err := Migrate2(bboltDB, jsonDB); err != nil {
			log.TLogln("Migrate2 failed")
			os.Exit(1)
		}
	} else {
		log.TLogln("Using SQLite storage backends...")

		sqliteDB := NewSqliteDB(filepath.Join(Path, "torrserver.db"))
		if sqliteDB == nil {
			log.TLogln("Error creating sqlite database")
			os.Exit(1)
		}

		jsonDB := NewJsonDB()
		if jsonDB == nil {
			log.TLogln("Error open jsonDB")
			os.Exit(1)
		}

		SettingsStorage, ViewedStorage, TorrentsStorage = jsonDB, jsonDB, sqliteDB
	}

	dbRouter := NewXPathDBRouter()
	// First registered DB becomes default route
	dbRouter.RegisterRoute(SettingsStorage, "Settings")
	dbRouter.RegisterRoute(ViewedStorage, "Viewed")
	dbRouter.RegisterRoute(TorrentsStorage, "Torrents")

	tdb = NewDBReadCache(dbRouter)

	loadBTSets()
	Migrate1()
}

func CloseDB() {
	tdb.CloseDB()
}
