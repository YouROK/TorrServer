module server

go 1.17

replace (
	github.com/anacrolix/dms v1.2.2 => github.com/tsynik/dms v0.0.0-20210909053938-38af4173d4ac
	github.com/anacrolix/torrent v1.31.0 => github.com/tsynik/torrent v1.2.7-0.20210907192509-2141ede9aa09
)

exclude github.com/willf/bitset v1.2.0

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/anacrolix/dms v1.2.2
	github.com/anacrolix/missinggo v1.3.0
	github.com/anacrolix/torrent v1.31.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/location v0.0.2
	github.com/gin-gonic/gin v1.7.4
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1
	go.etcd.io/bbolt v1.3.6
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
)

require (
	github.com/RoaringBitmap/roaring v0.9.4 // indirect
	github.com/alexflint/go-scalar v1.0.0 // indirect
	github.com/anacrolix/chansync v0.1.0 // indirect
	github.com/anacrolix/confluence v1.8.0 // indirect
	github.com/anacrolix/dht/v2 v2.10.4 // indirect
	github.com/anacrolix/ffprobe v1.0.0 // indirect
	github.com/anacrolix/log v0.9.0 // indirect
	github.com/anacrolix/missinggo/perf v1.0.0 // indirect
	github.com/anacrolix/missinggo/v2 v2.5.2 // indirect
	github.com/anacrolix/multiless v0.1.1-0.20210529082330-de2f6cf29619 // indirect
	github.com/anacrolix/stm v0.3.0 // indirect
	github.com/anacrolix/sync v0.4.0 // indirect
	github.com/anacrolix/upnp v0.1.2-0.20200416075019-5e9378ed1425 // indirect
	github.com/anacrolix/utp v0.1.0 // indirect
	github.com/benbjohnson/immutable v0.3.0 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
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
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/rs/dnscache v0.0.0-20210201191234-295bba877686 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/ugorji/go/codec v1.2.6 // indirect
	github.com/willf/bitset v1.1.11 // indirect
	github.com/willf/bloom v2.0.3+incompatible // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
