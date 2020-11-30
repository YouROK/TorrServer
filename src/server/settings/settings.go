package settings

var (
	tdb  *TDB
	Path string
)

func InitSets(readOnly bool) {
	tdb = NewTDB(readOnly)
	loadBTSets()
}

func CloseDB() {
	tdb.CloseDB()
}

func IsReadOnly() bool {
	if tdb == nil || tdb.ReadOnly {
		return true
	}
	return false
}
