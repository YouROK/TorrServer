module server

go 1.16

replace (
	github.com/anacrolix/dht/v2 v2.9.1 => github.com/anacrolix/dht/v2 v2.10.0
	github.com/anacrolix/dms v1.2.2 => github.com/yourok/dms v0.0.0-20210726184814-0838f1936b67
)

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/anacrolix/dht/v2 v2.10.0 // indirect
	github.com/anacrolix/dms v1.2.2
	github.com/anacrolix/missinggo v1.3.0
	github.com/anacrolix/torrent v1.29.1
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/location v0.0.2
	github.com/gin-gonic/gin v1.7.1
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1
	go.etcd.io/bbolt v1.3.5
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
)
