# Running in Docker & Docker Compose

Run in console

```bash
docker run --rm -d --name torrserver -p 8090:8090 ghcr.io/yourok/torrserver:latest
```

For persistent data, mount a local directory and set paths explicitly:

```bash
docker run --rm -d --name torrserver \
  -v ./ts:/opt/ts \
  -e TS_CONF_PATH=/opt/ts \
  -e TS_LOG_PATH=/opt/ts/torrserver.log \
  -e TS_TORR_DIR=/opt/ts/torrents \
  -p 8090:8090 \
  ghcr.io/yourok/torrserver:latest
```

## Environment Variables

- `TS_HTTPAUTH` – Set to `1` to enable basic authentication. The authentication file must be placed in the `~/ts/config` directory. This also protects the MCP endpoint at `/mcp`.
- `TS_RDB` – If set to `1`, enables the `--rdb` command-line flag.
- `TS_DONTKILL` – If set to `1`, enables the `--dontkill` command-line flag.
- `TS_IP` – Specifies the bind address for the web server using the `--ip` flag.
- `TS_PORT` – Overrides the default port (e.g., to `5555`). The corresponding port mapping must also be updated from `-p 8090:8090` to `-p 5555:5555` to reflect the change.
- `TS_CONF_PATH` – Overrides the internal configuration path for TorrServer inside the container (e.g., `/opt/tsss`).
- `TS_TORR_DIR` – Overrides the directory used for storing torrent files (e.g., `/opt/torr_files`).
- `TS_TORR_ADDR` – Sets the torrent client address via the `--torrentaddr` flag.
- `TS_LOG_PATH` – Overrides the log file path (e.g., `/opt/torrserver.log`).
- `TS_PROXYURL` – Defines a proxy URL for BitTorrent traffic. Supports HTTP, SOCKS4, SOCKS5, and SOCKS5H protocols (e.g., `socks5h://user:password@example.com:2080`).
- `TS_PROXYMODE` – Determines the proxy usage mode. Acceptable values are:
  - `tracker` – Proxy only HTTP trackers (default).
  - `peers` – Proxy only peer connections.
  - `full` – Proxy all BitTorrent traffic.
- `TS_SEARCH_WA_ENABLE` – If set to `1`, enables the `--searchwa` flag.
- `TS_SSL_ENABLE` – If set to `1`, enables the `--ssl` flag.
- `TS_SSL_PORT` – Specifies the HTTPS port via the `--sslport` flag.
- `TS_SSL_CERT_PATH` – Specifies the path to the SSL certificate file via the `--sslcert` flag.
- `TS_SSL_KEY_PATH` – Specifies the path to the SSL private key file via the `--sslkey` flag.
- `TS_FORCE_HTTPS_ENABLE` – If set to `1`, enables the `--force-https` flag.
- `TS_HTTP_MEDIA_ENABLE` – If set to `1`, enables the `--http-media` flag.
- `TS_HTTPS_ONLY_ENABLE` – If set to `1`, enables the `--https-only` flag.
- `TS_WEB_LOG_PATH` – Overrides the web server log path using the `--weblogpath` flag.
- `TS_PUBLIC_IPV4_ADDR` – Sets the public IPv4 address using the `--pubipv4` flag.
- `TS_PUBLIC_IPV6_ADDR` – Sets the public IPv6 address using the `--pubipv6` flag.
- `TS_MAX_SIZE` – Defines the maximum allowed stream size (in bytes) via the `--maxsize` flag.
- `TS_TELEGRAM_TOKEN` – Sets the Telegram bot token using the `--tgtoken` flag.
- `TS_FUSE_PATH` – Sets the FUSE mount point path using the `--fusepath` flag.
- `TS_WEBDAV_ENABLE` – If set to `1`, enables WebDAV support via the `--webdav` flag.

Go runtime variables (read by the Go runtime, not by the entrypoint):

- `GODEBUG` – The image sets `madvdontneed=1`, so piece cache RAM released after a torrent closes goes back to the OS right away.
- `GOMEMLIMIT` – Unset by default. A soft heap limit, for example `GOMEMLIMIT=200MiB`. While a torrent plays, memory sits well above the cache size: with the default 64 MiB cache the process grew from about 29 MiB idle to about 127 MiB. That is the cache, the 8 MiB windows at the start and end of each open file that are never evicted, BitTorrent client state, and garbage between collections (Go lets the heap grow to about twice the live data before it collects). With a limit set, the GC runs more often as the heap approaches it. Use it on routers and NAS boxes with little RAM. Keep it well above the cache size plus about 64 MiB: a limit below what the cache holds makes the GC run constantly and burns CPU without freeing anything.

Example with full override command (on default values):

```bash
docker run --rm -d --name torrserver \
  -v ./ts:/opt/ts \
  -e TS_PORT=5665 \
  -e TS_DONTKILL=1 \
  -e TS_HTTPAUTH=1 \
  -e TS_RDB=1 \
  -e TS_CONF_PATH=/opt/ts \
  -e TS_LOG_PATH=/opt/ts/torrserver.log \
  -e TS_TORR_DIR=/opt/ts/torrents \
  -e TS_PROXYURL=socks5h://user:password@example.com:2080 \
  -e TS_PROXYMODE=tracker \
  -p 5665:5665 \
  ghcr.io/yourok/torrserver:latest
```

## Docker Compose

With `network_mode: host` (needed for DLNA), do not publish `ports:` — the container uses the host network directly. Use bridge networking and a `ports` mapping only when DLNA is not required.

```yml
# docker-compose.yml

version: '3.3'
services:
    torrserver:
        image: ghcr.io/yourok/torrserver
        container_name: torrserver
        network_mode: host    # to allow DLNA feature; omit ports: in this mode
        environment:
            - TS_PORT=5665
            - TS_DONTKILL=1
            - TS_HTTPAUTH=0
            - TS_CONF_PATH=/opt/ts/config
            - TS_TORR_DIR=/opt/ts/torrents
        volumes:
            - './CACHE:/opt/ts/torrents'
            - './CONFIG:/opt/ts/config'
        restart: unless-stopped
        

```
