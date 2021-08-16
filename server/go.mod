module server

go 1.16

replace github.com/anacrolix/dht/v2 v2.9.1 => github.com/anacrolix/dht/v2 v2.10.0

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/anacrolix/dht/v2 v2.10.3 // indirect
	github.com/anacrolix/missinggo v1.3.0
	github.com/anacrolix/torrent v1.30.3-0.20210816011131-16176b762e4a
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/location v0.0.2
	github.com/gin-gonic/gin v1.7.1
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1
	go.etcd.io/bbolt v1.3.6
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
)
