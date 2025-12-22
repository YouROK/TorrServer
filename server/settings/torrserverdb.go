package settings

type TorrServerDB interface {
	CloseDB()
	Get(xPath, name string) []byte
	Set(xPath, name string, value []byte)
	List(xPath string) []string
	Rem(xPath, name string)
	Clear(xPath string)
}
