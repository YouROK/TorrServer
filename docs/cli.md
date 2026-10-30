# Server args

- `--port PORT`, `-p PORT` - web server port (default 8090)
- `--ip IP`, `-i IP` - web server bind addr (repeatable; default empty binds to all interfaces)
- `--ssl` - enables HTTPS for web server
- `--sslport PORT` -  web server https port (default 8091). If not set, will be taken from db (if stored previously) or the default will be used.
- `--sslcert PATH` -  path to SSL cert file. If not set, will be taken from db (if stored previously) or default self-signed certificate/key will be generated. Replacing the cert/key files (e.g. a certbot renewal) takes effect without a restart. An invalid user-supplied cert stops startup instead of being replaced.
- `--sslkey PATH` - path to SSL key file. If not set, will be taken from db (if stored previously) or default self-signed certificate/key will be generated.
- `--force-https` - with `--ssl`, the HTTP listener (`--port`) answers every request with **307 Temporary Redirect** to the same path on HTTPS (`--sslport`). Requires `--ssl` (startup fails if `--force-https` is set without `--ssl`). Default is off. Media players and TVs usually can't use a self-signed certificate, see [HTTPS](https.md).
- `--http-media` - with `--force-https`, keep serving media URLs (`/stream`, `/play`, `/playlist`, `/playlistall`, which DLNA also uses, and GStreamer HLS under `/gst/<hash>/`) over plain HTTP for players that can't use HTTPS; everything else still redirects. Stream URLs and Basic auth credentials then travel unencrypted, so only use it on a trusted network and never expose the HTTP port to the internet. Requires `--force-https`.
- `--https-only` - with `--ssl`, don't open the plain HTTP port (`--port`) at all: the UI, API, media, DLNA links and Bonjour all use HTTPS on `--sslport`. Unlike `--force-https`, clients using the usual `http://host:<port>` address can't send credentials or requests over plain HTTP before being redirected. Use it with a trusted certificate, see [HTTPS](https.md); with the self-signed one, players and TVs won't play. Requires `--ssl`; can't be combined with `--http-media`.

- `--path PATH`, `-d PATH` - database and config dir path
- `--logpath LOGPATH`, `-l LOGPATH` - server log file path
- `--weblogpath WEBLOGPATH`, `-w WEBLOGPATH` - web access log file path
- `--rdb`, `-r` - start in read-only DB mode
- `--httpauth`, `-a` - enable HTTP Auth on all requests
- `--dontkill`, `-k` - don't kill server on signal
- `--ui`, `-u` - open torrserver page in browser
- `--torrentsdir TORRENTSDIR`, `-t TORRENTSDIR` - autoload torrents from dir
- `--torrentaddr TORRENTADDR` - Torrent client address (format [IP]:PORT, ex. :32000, 127.0.0.1:32768 etc.)
- `--pubipv4 PUBIPV4`, `-4 PUBIPV4` - set public IPv4 addr
- `--pubipv6 PUBIPV6`, `-6 PUBIPV6` - set public IPv6 addr
- `--searchwa`, `-s` - allow search without authentication
- `--maxsize MAXSIZE`, `-m MAXSIZE` - max allowed stream size (in Bytes)
- `--tgtoken TGTOKEN`, `-T TGTOKEN` - [Telegram bot](../server/tgbot/README.md) token
- `--fusepath FUSEPATH`, `-f FUSEPATH` - FUSE mount path
- `--webdav` - enable WebDAV
- `--proxyurl PROXYURL` - set proxy URL for BitTorrent traffic (HTTP, SOCKS4, SOCKS5, SOCKS5H), example: socks5h://user:<password@example.com>:2080
- `--proxymode PROXYMODE` - set proxy mode: "tracker" (only HTTP trackers, default), "peers" (only peer connections), or "full" (all traffic)
- `--pprof` - enable `/debug/pprof/` (heap, goroutine and CPU profiles). Use it with `--httpauth`: a heap profile exposes memory contents. Memory and goroutine counters are also at `/runtime/status`.
- `--help`, `-h` - display this help and exit
- `--version` - display version and exit

Example:

```bash
TorrServer-darwin-arm64 [--port PORT] [--ip IP ...] [--path PATH] [--logpath LOGPATH] [--weblogpath WEBLOGPATH] [--rdb] [--httpauth] [--dontkill] [--ui] [--torrentsdir TORRENTSDIR] [--torrentaddr TORRENTADDR] [--pubipv4 PUBIPV4] [--pubipv6 PUBIPV6] [--searchwa] [--maxsize MAXSIZE] [--tgtoken TGTOKEN] [--fusepath FUSEPATH] [--webdav] [--ssl] [--sslport PORT] [--sslcert PATH] [--sslkey PATH] [--force-https] [--http-media] [--https-only]
```
