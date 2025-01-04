package settings

type TorrServerDB interface {
	CloseDB()
	Get(xPath string, name string) []byte
	Set(xPath string, name string, value []byte)
	List(xPath string) []string
	Rem(xPath string, name string)
}
