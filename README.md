<p align="center" style="text-align: center">
  <img src="https://github.com/YouROK/TorrServer/assets/144587546/53f7175a-cda4-4a06-86b6-2ac07582dcf1" width="33%"><br/>
</p>

<p align="center">
  Simple and powerful tool for streaming torrents.
  <br/>
  <br/>
  <a href="https://github.com/YouROK/TorrServer/blob/master/LICENSE">
    <img alt="GitHub" src="https://img.shields.io/github/license/YouROK/TorrServer"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/YouROK/TorrServer">
    <img src="https://goreportcard.com/badge/github.com/YouROK/TorrServer" />
  </a>
  <a href="https://pkg.go.dev/github.com/YouROK/TorrServer">
    <img src="https://pkg.go.dev/badge/github.com/YouROK/TorrServer.svg" alt="Go Reference"/>
  </a>
  <a href="https://github.com/YouROK/TorrServer/issues">
    <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="CodeFactor" />
  </a>
  <a href="https://github.com/YouROK/TorrServer/actions/workflows/docker_image.yml" rel="nofollow">
    <img src="https://img.shields.io/github/actions/workflow/status/YouROK/TorrServer/docker_image.yml?logo=Github" alt="Build" />
  </a>
  <a href="https://github.com/YouROK/TorrServer/releases" rel="nofollow">
    <img alt="GitHub release (latest SemVer)" src="https://img.shields.io/github/v/release/YouROK/TorrServer?label=version"/>
  </a>
  <a href="https://github.com/YouROK/TorrServer/tags" rel="nofollow">
    <img alt="GitHub tag (latest SemVer pre-release)" src="https://img.shields.io/github/v/tag/YouROK/TorrServer?include_prereleases&label=pre-release"/>
  </a>
</p>

## Introduction

TorrServer is a program that allows users to view torrents online without the need for preliminary file downloading.
The core functionality of TorrServer includes caching torrents and subsequent data transfer via the HTTP protocol,
allowing the cache size to be adjusted according to the system parameters and the user's internet connection speed.

## AI Documentation

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/YouROK/TorrServer)

## Features

- Caching
- Streaming
- Local and Remote Server
- Viewing torrents on various devices
- Integration with other apps through API
- Torznab search (Jackett, Prowlarr, and similar indexer managers)
- Cross-browser modern web interface
- Optional DLNA server
- Optional GStreamer HLS transcoding (`-gst` builds)

## Getting Started

### Installation

Download the application for the required platform in the [releases](https://github.com/YouROK/TorrServer/releases) page. After installation, open the link <http://127.0.0.1:8090> in the browser.

#### Windows

Run `TorrServer-windows-amd64.exe`.

#### Linux

Run in console

```bash
curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh | sudo bash
```

The script supports interactive and non-interactive installation, configuration, updates, and removal. When running the script interactively, you can:

- **Install/Update**: Choose to install or update TorrServer
- **Reconfigure**: If TorrServer is already installed, you'll be prompted to reconfigure settings (port, auth, read-only mode, logging, BBR)
- **Uninstall**: Type `Delete` (or `Удалить` in Russian) to uninstall TorrServer

**Download first and set execute permissions:**

```bash
curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh -o installTorrServerLinux.sh && chmod 755 installTorrServerLinux.sh
```

**Command-line examples:**

- Install a specific version:

  ```bash
  sudo bash ./installTorrServerLinux.sh --install 135 --silent
  ```

- Update to latest version:

  ```bash
  sudo bash ./installTorrServerLinux.sh --update --silent
  ```

- Reconfigure settings interactively:

  ```bash
  sudo bash ./installTorrServerLinux.sh --reconfigure
  ```

- Check for updates:

  ```bash
  sudo bash ./installTorrServerLinux.sh --check
  ```

- Downgrade to a specific version:

  ```bash
  sudo bash ./installTorrServerLinux.sh --down 135
  ```

- Remove/uninstall:

  ```bash
  sudo bash ./installTorrServerLinux.sh --remove --silent
  ```

- Change the systemd service user:

  ```bash
  sudo bash ./installTorrServerLinux.sh --change-user root --silent
  ```

**All available commands:**

- `--install [VERSION]` - Install latest or specific version
- `--update` - Update to latest version
- `--reconfigure` - Reconfigure TorrServer settings (port, auth, read-only mode, logging, BBR)
- `--check` - Check for updates (version info only)
- `--down VERSION` - Downgrade to specific version
- `--remove` - Uninstall TorrServer
- `--change-user USER` - Change service user (root|torrserver)
- `--root` - Run service as root user
- `--silent` - Non-interactive mode with defaults
- `--help` - Show help message

#### macOS

Run in Terminal.app

```bash
curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerMac.sh -o installTorrserverMac.sh && chmod 755 installTorrServerMac.sh && bash ./installTorrServerMac.sh
```

Alternative install script for Intel Macs: <https://github.com/dancheskus/TorrServerMacInstaller>

#### IOCage Plugin (Unofficial)

On FreeBSD (TrueNAS/FreeNAS) you can use this plugin: <https://github.com/filka96/iocage-plugin-TorrServer>

#### NAS Systems (Unofficial)

- Several releases are available through this link: <https://github.com/vladlenas>
- **Synology NAS** packages repo source: <https://grigi.lt>

### Server args

- `--port PORT`, `-p PORT` - web server port (default 8090)
- `--ip IP`, `-i IP` - web server addr (default empty, binds to all interfaces)
- `--ssl` - enables https for web server
- `--sslport PORT` -  web server https port (default 8091). If not set, will be taken from db (if stored previously) or the default will be used.
- `--sslcert PATH` -  path to ssl cert file. If not set, will be taken from db (if stored previously) or default self-signed certificate/key will be generated.
- `--sslkey PATH` - path to ssl key file. If not set, will be taken from db (if stored previously) or default self-signed certificate/key will be generated.
- `--force-https` - with `--ssl`, the HTTP listener (`--port`) answers only with **307 Temporary Redirect** to the same path on HTTPS (`--sslport`). The web UI and API are served on HTTPS only; nothing is served on HTTP except redirects. Requires `--ssl` (startup fails if `--force-https` is set without `--ssl`). Default is off so plain HTTP still works when SSL is disabled.
- `--path PATH`, `-d PATH` - database and config dir path
- `--logpath LOGPATH`, `-l LOGPATH` - server log file path
- `--weblogpath WEBLOGPATH`, `-w WEBLOGPATH` - web access log file path
- `--rdb`, `-r` - start in read-only DB mode
- `--httpauth`, `-a` - enable http auth on all requests
- `--dontkill`, `-k` - don't kill server on signal
- `--ui`, `-u` - open torrserver page in browser
- `--torrentsdir TORRENTSDIR`, `-t TORRENTSDIR` - autoload torrents from dir
- `--torrentaddr TORRENTADDR` - Torrent client address (format [IP]:PORT, ex. :32000, 127.0.0.1:32768 etc)
- `--pubipv4 PUBIPV4`, `-4 PUBIPV4` - set public IPv4 addr
- `--pubipv6 PUBIPV6`, `-6 PUBIPV6` - set public IPv6 addr
- `--searchwa`, `-s` - allow search without authentication
- `--maxsize MAXSIZE`, `-m MAXSIZE` - max allowed stream size (in Bytes)
- `--tg TGTOKEN`, `-T TGTOKEN` - [Telegram bot](server/tgbot/README.md) token
- `--fuse FUSEPATH`, `-f FUSEPATH` - fuse mount path
- `--webdav` - enable web dav
- `--proxyurl PROXYURL` - set proxy URL for BitTorrent traffic (http, socks4, socks5, socks5h), example: socks5h://user:password@example.com:2080
- `--proxymode PROXYMODE` - set proxy mode: "tracker" (only HTTP trackers, default), "peers" (only peer connections), or "full" (all traffic)
- `--help`, `-h` - display this help and exit
- `--version` - display version and exit

Example:

```bash
TorrServer-darwin-arm64 [--port PORT] [--ip IP] [--path PATH] [--logpath LOGPATH] [--weblogpath WEBLOGPATH] [--rdb] [--httpauth] [--dontkill] [--ui] [--torrentsdir TORRENTSDIR] [--torrentaddr TORRENTADDR] [--pubipv4 PUBIPV4] [--pubipv6 PUBIPV6] [--searchwa] [--maxsize MAXSIZE] [--tg TGTOKEN] [--fuse FUSEPATH] [--webdav] [--ssl] [--sslport PORT] [--sslcert PATH] [--sslkey PATH] [--force-https]
```

### Running in Docker & Docker Compose

Run in console

```bash
docker run --rm -d --name torrserver -p 8090:8090 ghcr.io/yourok/torrserver:latest
```

For running in persistence mode, just mount volume to container by adding `-v ~/ts:/opt/ts`, where `~/ts` folder path is just example, but you could use it anyway... Result example command:

```bash
docker run --rm -d --name torrserver -v ~/ts:/opt/ts -p 8090:8090 ghcr.io/yourok/torrserver:latest
```

#### Environments

- `TS_HTTPAUTH` - 1, and place auth file into `~/ts/config` folder for enabling basic auth
- `TS_RDB` - if 1, then the enabling `--rdb` flag
- `TS_DONTKILL` - if 1, then the enabling `--dontkill` flag
- `TS_PORT` - for changind default port to **5555** (example), also u need to change `-p 8090:8090` to `-p 5555:5555` (example)
- `TS_CONF_PATH` - for overriding torrserver config path inside container. Example `/opt/tsss`
- `TS_TORR_DIR` - for overriding torrents directory. Example `/opt/torr_files`
- `TS_LOG_PATH` - for overriding log path. Example `/opt/torrserver.log`
- `TS_PROXYURL` - set proxy URL for BitTorrent traffic (http, socks4, socks5, socks5h), example: socks5h://user:password@example.com:2080
- `TS_PROXYMODE` - set proxy mode: "tracker" (only HTTP trackers, default), "peers" (only peer connections), or "full" (all traffic)

Example with full overrided command (on default values):

```bash
docker run --rm -d -e TS_PORT=5665 -e TS_DONTKILL=1 -e TS_HTTPAUTH=1 -e TS_RDB=1 -e TS_CONF_PATH=/opt/ts/config -e TS_LOG_PATH=/opt/ts/log -e TS_TORR_DIR=/opt/ts/torrents -e TS_PROXYURL=socks5h://user:password@example.com:2080 -e TS_PROXYMODE=tracker --name torrserver -v ~/ts:/opt/ts -p 5665:5665 ghcr.io/yourok/torrserver:latest
```

#### Docker Compose

```yml
# docker-compose.yml

version: '3.3'
services:
    torrserver:
        image: ghcr.io/yourok/torrserver
        container_name: torrserver
        network_mode: host    # to allow DLNA feature
        environment:
            - TS_PORT=5665
            - TS_DONTKILL=1
            - TS_HTTPAUTH=0
            - TS_CONF_PATH=/opt/ts/config
            - TS_TORR_DIR=/opt/ts/torrents
        volumes:
            - './CACHE:/opt/ts/torrents'
            - './CONFIG:/opt/ts/config'
        ports:
            - '5665:5665'
        restart: unless-stopped
        

```

### Smart TV (using Media Station X)

1. Install **Media Station X** on your Smart TV (see [platform support](https://msx.benzac.de/info/?tab=PlatformSupport))

2. Open it and go to: **Settings -> Start Parameter -> Setup**

3. Enter current ip and port of the TorrServe(r), e.g. `127.0.0.1:8090`

## Development

### Go server

To run the Go server locally, just run

```bash
cd server
go run ./cmd
```

### Web development

To run the web server locally, just run

```bash
yarn start
```

More info at <https://github.com/YouROK/TorrServer/tree/master/web#readme>

### Build

#### Server

- Install [Golang](https://golang.org/doc/install) 1.20+
- Go to the TorrServer source directory
- Run build script under linux or macOS `build-all.sh`

#### Web

- Install **npm** and **yarn**
- Go to the web directory
- Run `NODE_OPTIONS=--openssl-legacy-provider yarn build`

#### Android

To build an Android server you will need the Android Toolchain.

#### Swagger

`swag` must be installed on the system to [re]build Swagger documentation.

```bash
go install github.com/swaggo/swag/cmd/swag@latest
cd server; swag init -g web/server.go

# Documentation can be linted using
swag fmt
```

## API

### API Docs

API documentation is hosted as Swagger format available at path `/swagger/index.html`.

## Authentication

The users data file should be located near to the settings. Basic auth, read more in wiki <https://en.wikipedia.org/wiki/Basic_access_authentication>.

`accs.db` in JSON format:

```json
{
    "User1": "Pass1",
    "User2": "Pass2"
}
```

Note: You should enable authentication with -a (--httpauth) TorrServer startup option.

## Whitelist/Blacklist IP

The lists file should be located in the same directory with config.db.

- Whitelist file name: `wip.txt`
- Blacklist file name: `bip.txt`

Whitelist has priority over everything else.

Example:

```text
local:127.0.0.0-127.0.0.255
127.0.0.0-127.0.0.255
local:127.0.0.1
127.0.0.1
# at the beginning of the line, comment
```

## Torznab

TorrServer can talk to **Torznab** indexers so you can search for torrents from tools like **Jackett** and **Prowlarr**, including searching several configured indexers at once.

Configure it in the web UI: **Settings → Torznab**.

### Indexer parameters

Each Torznab indexer needs:

- **Host URL**: full URL to the Torznab API endpoint.
  - Jackett example:

  ```shell
  http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/
  ```

  - Prowlarr example:
  
  ```shell
  http://localhost:9696/1
  ```
  
  - Make sure to include the correct trailing slash (`/`) in your indexer's URL,
  as required by your Torznab provider. TorrServer will try to properly format the path,
  but matching your indexer's expected format is best to avoid connection issues.
  
- **API Key**: the key from your Torznab indexer manager.

### Enabling Torznab search

1. Open **Settings**.
2. Open the **Torznab** tab.
3. Turn on **Enable Torznab Search**.
4. Enter **Host URL** and **API Key**, then **Add Server** for each indexer.
5. **Save** settings.

## GStreamer

GStreamer enables **HLS transcoding** for Matroska/WebM torrents when the client cannot play the original video or audio codec directly.

### The `-gst` binary (main requirement)

The **only** way to enable GStreamer in TorrServer is to run a build compiled with the `gst` tag. There is no on/off switch in `settings.json`.

| Binary | GStreamer |
| --- | --- |
| `TorrServer-gst-<os>-<arch>` | **Yes** — `/gst/*` routes and transcoding |
| `TorrServer-<os>-<arch>` (standard) | **No** — GStreamer settings can be saved but pipelines will not run |

Download `TorrServer-gst-*` from [releases](https://github.com/YouROK/TorrServer/releases) (Windows/Linux amd64+arm64/macOS amd64+arm64), or build:

```bash
cd server
go build -tags="nosqlite gst" -trimpath -o TorrServer-gst ./cmd
```

Windows builds with embedded libs additionally use the `embed_gstlib` tag (see below).

Per-codec behavior is configured in the `gstreamer` settings block (`TranscodeH264`, `TranscodeH265`, `TranscodeAV1`, `TranscodeVP9`).

Linux/macOS [install scripts](installTorrServerLinux.sh) install the **standard** binary only — for GStreamer, use the `-gst` release manually.

### Bundled vs dynamically loaded GStreamer

TorrServer itself stays a single Go binary. GStreamer is **not** linked into it at compile time (Linux/macOS/Windows use runtime loading via `dlopen` / `LoadLibrary`). What changes is **where** the GStreamer libraries and plugins come from:

| Mode | How it works | Typical use |
| --- | --- | --- |
| **Bundled (embedded)** | GStreamer runtime is packed inside the `-gst` binary and extracted to cache on first run | Windows `TorrServer-gst-windows-amd64.exe` built with `embed_gstlib` |
| **Bundled (portable)** | Place a `gst-lib/` directory next to the TorrServer executable (same layout as an extracted GStreamer runtime) | Portable installs, custom deployments |
| **Dynamic (system)** | TorrServer loads `libgstreamer` / DLLs from OS packages or a system install path (`GSTPath`, `/opt/gstreamer`, framework path, etc.) | Linux, macOS, Windows with [official GStreamer installer](https://gstreamer.freedesktop.org/download/) |

Auto-detection order includes: `GSTPath` from settings → `gst-lib/` beside the binary → common system paths → `LD_LIBRARY_PATH` / `PATH`.

Verify whichever mode you use with `GET /gst/echo` or the status lines on the **GStreamer** settings tab.

### Installing GStreamer for dynamic loading

Only needed when GStreamer is **not** already bundled (most Linux/macOS `-gst` builds, Windows without `embed_gstlib`).

**Debian / Ubuntu**

```bash
sudo apt update
sudo apt install -y \
  gstreamer1.0-tools \
  gstreamer1.0-plugins-base \
  gstreamer1.0-plugins-good \
  gstreamer1.0-plugins-bad \
  gstreamer1.0-plugins-ugly \
  gstreamer1.0-libav
```

**Fedora / RHEL / Rocky / AlmaLinux**

```bash
sudo dnf install -y \
  gstreamer1-tools \
  gstreamer1-plugins-base \
  gstreamer1-plugins-good \
  gstreamer1-plugins-bad-free \
  gstreamer1-plugins-ugly-free \
  gstreamer1-libav
```

`x264enc` (used when transcoding video to H.264) may require [RPM Fusion](https://rpmfusion.org/) on Fedora.

**Arch Linux**

```bash
sudo pacman -S --needed \
  gst-plugins-base \
  gst-plugins-good \
  gst-plugins-bad \
  gst-plugins-ugly \
  gst-libav
```

**macOS**

[Official framework](https://gstreamer.freedesktop.org/download/) or Homebrew:

```bash
brew install gstreamer gst-plugins-base gst-plugins-good gst-plugins-bad gst-plugins-ugly gst-libav
```

Set `GSTPath` if needed (e.g. `/Library/Frameworks/GStreamer.framework/Versions/1.0`).

**Windows (dynamic)**

Install MSVC 64-bit runtime from [gstreamer.freedesktop.org](https://gstreamer.freedesktop.org/download/) (default: `C:\Program Files\gstreamer\1.0\mingw_x86_64`), or use a `gst-lib/` folder next to the exe. Embedded `-gst` Windows builds may not need a separate install.

### Web UI configuration

1. Open **Settings**.
2. Enable **PRO mode**.
3. Open the **GStreamer** tab.
4. Adjust options and click **Save GStreamer Settings**.

Settings are stored separately from main BitTorrent settings and take effect immediately for new streams.

### `settings.json` block

GStreamer options are stored in `settings.json` under the `gstreamer` key (inside the `Settings` section). Legacy keys `gst` and `GStreamer` are still read on load; saving from the web UI or API writes the `gstreamer` key.

Example (`settings.json`):

```json
{
  "gstreamer": {
    "GSTVersion": 1.26,
    "GSTPath": "",
    "Source": "stream",
    "InactiveMinutes": 5,
    "AACBitrateKbps": 256,
    "SegmentSeconds": 6,
    "appsinkBuffers": 1000,
    "TranscodeH264": false,
    "TranscodeH265": false,
    "TranscodeAV1": false,
    "TranscodeVP9": false,
    "VideoBitrate": 10000,
    "tempfs": false,
    "tempfs_ring": 0
  }
}
```

| Field | Description |
| --- | --- |
| `GSTVersion` | Installed GStreamer version (minimum `1.26`, e.g. `1.26`, `1.28`). Used for pipeline feature selection. |
| `GSTPath` | Path to GStreamer installation. Empty = auto-detection. |
| `Source` | Input URL mode: `stream` (`/stream/...`) or `play` (`/play/...`). |
| `InactiveMinutes` | Freeze pipeline after this many minutes without playback. |
| `AACBitrateKbps` | Audio transcoding bitrate in kbps. |
| `SegmentSeconds` | HLS segment length in seconds. |
| `appsinkBuffers` | Number of buffers in the appsink queue. |
| `TranscodeH264` | Transcode H.264 video when needed. |
| `TranscodeH265` | Transcode H.265/HEVC video when needed. |
| `TranscodeAV1` | Transcode AV1 video when needed. |
| `TranscodeVP9` | Transcode VP9 video when needed. |
| `VideoBitrate` | Target video bitrate in kbps when transcoding. |
| `tempfs` | Use memory-backed tempfs for segments (Linux). |
| `tempfs_ring` | Extra tempfs ring blocks (`0` = default). |

### API

**Settings** (requires authentication when `--httpauth` is enabled):

- `GET /gst/settings` — current config and platform defaults
- `POST /gst/settings` — update config

  ```json
  { "action": "set", "config": { "GSTVersion": 1.26, "Source": "stream" } }
  ```

  Reset to defaults:

  ```json
  { "action": "def" }
  ```

**Streaming** (available in `-gst` builds):

| Endpoint | Description |
| --- | --- |
| `GET /gst/echo` | GStreamer / gst-discoverer health check |
| `GET /gst/:hash/probe` | Probe torrent file codecs (`index`, `id`, or `fileID` query) |
| `GET /gst/:hash/master.m3u8` | HLS master playlist |
| `GET /gst/:hash/init.mp4` | Initialization segment |
| `GET /gst/:hash/seg/*segment` | Media segment |
| `GET /gst/:hash/heartbeat` | Keep-alive for active transcode task |
| `GET /gst/remove` | Stop transcode task (`hash` or `id` query) |

## Donate

- [YooMoney](https://yoomoney.ru/to/410013733697114/200)
- [Boosty](https://boosty.to/yourok)
- [TBank](https://www.tbank.ru/cf/742qEMhKhKn)

## Thanks to everyone who tested and helped

- [anacrolix](https://github.com/anacrolix) Matt Joiner
- [tsynik](https://github.com/tsynik) Nikk Gitanes
- [dancheskus](https://github.com/dancheskus) for react web GUI and PWA code
- [kolsys](https://github.com/kolsys) for initial Media Station X support
- [damiva](https://github.com/damiva) for Media Station X code updates
- [vladlenas](https://github.com/vladlenas) for NAS builds
- [pavelpikta](https://github.com/pavelpikta) Pavel Pikta for linux install script and more
- [Nemiroff](https://github.com/Nemiroff) Tw1cker
- [spawnlmg](https://github.com/spawnlmg) SpAwN_LMG for testing
- [TopperBG](https://github.com/TopperBG) Dimitar Maznekov for Bulgarian web translation
- [FaintGhost](https://github.com/FaintGhost) Zhang Yaowei for Simplified Chinese web translation
- [Anton111111](https://github.com/Anton111111) Anton Potekhin for sleep on Windows fixes
- [lieranderl](https://github.com/lieranderl) Evgeni for adding SSL support code
- [cocool97](https://github.com/cocool97) for openapi API documentation and torrent categories
- [shadeov](https://github.com/shadeov) for README improvements
- [butaford](https://github.com/butaford) Pavel for make docker file and scripts
- [filimonic](https://github.com/filimonic) Alexey D. Filimonov
- [leporel](https://github.com/leporel) Viacheslav Evseev
- and others
