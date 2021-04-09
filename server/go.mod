module server

replace (
	github.com/anacrolix/dht v1.0.1 => github.com/YouROK/dht v0.0.0-20210323193211-11c03221cb67
	github.com/anacrolix/torrent v1.2.6 => github.com/yourok/torrent v0.0.0-20210406082438-ad488c2037fc
)

go 1.16

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/anacrolix/missinggo v1.1.0
	github.com/anacrolix/torrent v1.2.6
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.3
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.8.1
	go.etcd.io/bbolt v1.3.5
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
)
