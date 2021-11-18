module server

go 1.17

replace (
	github.com/anacrolix/dms v1.3.0 => github.com/tsynik/dms v0.0.0-20211115041208-76e0609a5d1c
	github.com/anacrolix/torrent v1.38.0 => github.com/tsynik/torrent v1.2.7-0.20211118235503-1cf1470494b3
)

exclude (
	github.com/willf/bitset v1.2.0
	github.com/willf/bitset v1.2.1
)

require (
	github.com/alexflint/go-arg v1.4.2
	github.com/anacrolix/dms v1.3.0
	github.com/anacrolix/missinggo v1.3.0
	github.com/anacrolix/torrent v1.38.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/location v0.0.2
	github.com/gin-gonic/gin v1.7.4
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1
	go.etcd.io/bbolt v1.3.6
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11
)

require (
	github.com/RoaringBitmap/roaring v0.9.4 // indirect
	github.com/alexflint/go-scalar v1.1.0 // indirect
	github.com/anacrolix/chansync v0.3.0 // indirect
	github.com/anacrolix/confluence v1.10.0 // indirect
	github.com/anacrolix/dht/v2 v2.13.0 // indirect
	github.com/anacrolix/ffprobe v1.0.0 // indirect
	github.com/anacrolix/log v0.10.0 // indirect
	github.com/anacrolix/missinggo/perf v1.0.0 // indirect
	github.com/anacrolix/missinggo/v2 v2.5.2 // indirect
	github.com/anacrolix/multiless v0.2.0 // indirect
	github.com/anacrolix/stm v0.3.0 // indirect
	github.com/anacrolix/sync v0.4.0 // indirect
	github.com/anacrolix/upnp v0.1.2-0.20200416075019-5e9378ed1425 // indirect
	github.com/anacrolix/utp v0.1.0 // indirect
	github.com/benbjohnson/immutable v0.3.0 // indirect
	github.com/bits-and-blooms/bitset v1.2.1 // indirect
	github.com/bradfitz/iter v0.0.0-20191230175014-e8f45d346db8 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.9.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/rs/dnscache v0.0.0-20211102005908-e0241e321417 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/ugorji/go/codec v1.2.6 // indirect
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871 // indirect
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20211117180635-dee7805ff2e1 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
