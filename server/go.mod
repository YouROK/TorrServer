module server

go 1.21

toolchain go1.23.2

replace (
	github.com/anacrolix/torrent v1.54.1 => github.com/tsynik/torrent v1.2.22
	github.com/anacrolix/upnp v0.1.4 => github.com/tsynik/upnp v0.1.5
)

require (
	github.com/agnivade/levenshtein v1.1.1
	github.com/alexflint/go-arg v1.5.0
	github.com/anacrolix/dms v1.6.0
	github.com/anacrolix/log v0.15.2
	github.com/anacrolix/missinggo/v2 v2.7.3
	github.com/anacrolix/publicip v0.3.1
	github.com/anacrolix/torrent v1.54.1
	github.com/gin-contrib/cors v1.7.2
	github.com/gin-contrib/location v1.0.1
	github.com/gin-gonic/gin v1.10.0
	github.com/kljensen/snowball v0.9.0
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/pkg/errors v0.9.1
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.3
	github.com/wlynxg/anet v0.0.5
	go.etcd.io/bbolt v1.3.10
	golang.org/x/exp v0.0.0-20240531132922-fd00a4e0eefc
	golang.org/x/image v0.18.0
	golang.org/x/time v0.5.0
	gopkg.in/vansante/go-ffprobe.v2 v2.2.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/RoaringBitmap/roaring v1.9.4 // indirect
	github.com/alecthomas/atomic v0.1.0-alpha2 // indirect
	github.com/alexflint/go-scalar v1.2.0 // indirect
	github.com/anacrolix/chansync v0.4.0 // indirect
	github.com/anacrolix/dht/v2 v2.21.1 // indirect
	github.com/anacrolix/ffprobe v1.1.0 // indirect
	github.com/anacrolix/generics v0.0.1 // indirect
	github.com/anacrolix/missinggo v1.3.0 // indirect
	github.com/anacrolix/missinggo/perf v1.0.0 // indirect
	github.com/anacrolix/multiless v0.3.1-0.20221221005021-2d12701f83f7 // indirect
	github.com/anacrolix/stm v0.5.0 // indirect
	github.com/anacrolix/sync v0.5.1 // indirect
	github.com/anacrolix/upnp v0.1.4 // indirect
	github.com/anacrolix/utp v0.2.0 // indirect
	github.com/benbjohnson/immutable v0.4.3 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/bradfitz/iter v0.0.0-20191230175014-e8f45d346db8 // indirect
	github.com/bytedance/sonic v1.11.8 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/frankban/quicktest v1.14.6 // indirect
	github.com/gabriel-vasile/mimetype v1.4.7 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.21.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/rs/dnscache v0.0.0-20230804202142-fc85eb664529 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
