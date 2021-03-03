package settings

var (
	tdb      *TDB
	Path     string
	ReadOnly bool
	HttpAuth bool
)

func InitSets(readOnly bool) {
	ReadOnly = readOnly
	tdb = NewTDB()
	loadBTSets()
	Migrate()
}

func CloseDB() {
	tdb.CloseDB()
}
