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
- Optional GStreamer HLS remuxing and transcoding (`-gst` builds from release 141.10)

## Getting Started

### Installation

Download the application for the required platform in the [releases](https://github.com/YouROK/TorrServer/releases) page. After installation, open the link <http://127.0.0.1:8090> in the browser.

Standard binaries are named `TorrServer-<platform>-<arch>` (for example, `TorrServer-linux-amd64`). From release **141.10**, optional **GStreamer (gst)** builds are named `TorrServer-gst-<platform>-<arch>`. Supported gst targets are Windows amd64, Linux amd64/arm64, and macOS amd64/arm64. The install scripts below can install either variant when the matching release asset is available.

#### Windows

Run `TorrServer-windows-amd64.exe`.

#### Linux

Run in console

```bash
curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh | sudo bash
```

The script supports interactive and non-interactive installation, configuration, updates, and removal. When running the script interactively, you can:

- **Install/Update**: Choose to install or update TorrServer
- **GStreamer build**: For releases 141.10+ on amd64/arm64, choose the gst build with transcoding support (or pass `--gst`)
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

- Install GStreamer build (141.10+):

  ```bash
  sudo bash ./installTorrServerLinux.sh --install --gst
  sudo bash ./installTorrServerLinux.sh --update --gst --silent
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
- `--gst` - Install GStreamer build with transcoding support (141.10+, amd64/arm64 only)
- `--root` - Run service as root user
- `--silent` - Non-interactive mode with defaults
- `--help` - Show help message

#### macOS

Run in Terminal.app

```bash
curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerMac.sh -o installTorrServerMac.sh && chmod 755 installTorrServerMac.sh && bash ./installTorrServerMac.sh
```

The macOS install script supports the same commands as the Linux script, including `--install`, `--update`, `--remove`, `--reconfigure`, and `--gst` for the GStreamer build (141.10+).

**Command-line examples:**

- Install latest version:

  ```bash
  bash ./installTorrServerMac.sh --install
  ```

- Install GStreamer build:

  ```bash
  bash ./installTorrServerMac.sh --install --gst
  ```

- Update silently:

  ```bash
  bash ./installTorrServerMac.sh --update --silent
  ```

Alternative install script for Intel Macs: <https://github.com/dancheskus/TorrServerMacInstaller>

#### IOCage Plugin (Unofficial)

On FreeBSD (TrueNAS/FreeNAS) you can use this plugin: <https://github.com/filka96/iocage-plugin-TorrServer>

#### NAS Systems (Unofficial)

- Several releases are available through this link: <https://github.com/vladlenas>
- **Synology NAS** packages repo source: <https://grigi.lt>

### Server args

- `--port PORT`, `-p PORT` - web server port (default 8090)
- `--ip IP`, `-i IP` - web server bind addr (repeatable; default empty binds to all interfaces)
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
TorrServer-darwin-arm64 [--port PORT] [--ip IP ...] [--path PATH] [--logpath LOGPATH] [--weblogpath WEBLOGPATH] [--rdb] [--httpauth] [--dontkill] [--ui] [--torrentsdir TORRENTSDIR] [--torrentaddr TORRENTADDR] [--pubipv4 PUBIPV4] [--pubipv6 PUBIPV6] [--searchwa] [--maxsize MAXSIZE] [--tg TGTOKEN] [--fuse FUSEPATH] [--webdav] [--ssl] [--sslport PORT] [--sslcert PATH] [--sslkey PATH] [--force-https]
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
cd server
swag init -g web/server.go --parseDependency --parseInternal --parseDepth 5

# Documentation can be linted using
swag fmt
```

Standard binaries serve a filtered Swagger spec at runtime (only `/gst/settings`); `-gst` builds include all `/gst/*` endpoints.

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

GStreamer adds HLS output for Matroska/WebM torrents and, when enabled, AVI files. Compatible video and AAC audio can be remuxed without quality loss; unsupported streams can be transcoded to H.264/AAC.

### The `-gst` binary

GStreamer support is available only in a binary compiled with the `gst` build tag. There is no runtime switch that can add `/gst/*` routes to a standard binary.

| Binary | GStreamer |
| --- | --- |
| `TorrServer-gst-<os>-<arch>` | Yes - `/gst/*` routes, remuxing, and transcoding |
| `TorrServer-<os>-<arch>` | No - `GET /gst/settings` reports `built_in: false` |

Use the `--gst` option with the Linux/macOS install scripts, download a matching asset from [releases](https://github.com/YouROK/TorrServer/releases), or build from the repository root:

```bash
cd server
CGO_ENABLED=0 go build -tags="nosqlite gst" -trimpath -o TorrServer-gst ./cmd
```

Supported gst targets are Windows amd64, Linux amd64/arm64, and macOS amd64/arm64. A Windows single-file build with an embedded runtime additionally uses the `embed_gstlib` tag and `server/gstreamer/gst-libs/win-x86_64`; see [BUILD_WINDOWS_GSTREAMER.md](BUILD_WINDOWS_GSTREAMER.md).

### Runtime loading

The gst build uses `purego` to load GStreamer at runtime. It is built with `CGO_ENABLED=0` and does not link GStreamer into the executable at build time.

| Runtime mode | Platforms | How it works |
| --- | --- | --- |
| System or custom path | Windows, Linux, macOS | Loads libraries and plugins from `GSTPath` or a platform installation |
| Portable `gst-lib/` beside the executable | Windows amd64 | Uses the normal GStreamer directory layout without embedding it |
| Embedded `gst-lib` | Windows amd64 | Extracts the runtime from the executable into a versioned TorrServer cache directory on first use |

Runtime roots are tried in this order:

1. `GSTPath` from settings.
2. Platform defaults such as `/usr`, `/usr/local`, the macOS framework/Homebrew prefixes, or the Windows MinGW install directory.
3. Windows only: `gst-lib/` beside the executable.
4. Windows only: the embedded runtime, when compiled with `embed_gstlib`.
5. The operating-system loader search path (`PATH`, `LD_LIBRARY_PATH`, or `DYLD_LIBRARY_PATH`).

Linux and macOS do not use a portable `gst-lib` directory. Install GStreamer system-wide or set `GSTPath`. When GStreamer loads successfully, its real version from `gst_version()` takes precedence over the configured fallback version.

### Runtime requirements

TorrServer requires **GStreamer 1.22 or newer** and `gst-discoverer-1.0`. The full package set is recommended because the pipeline may need HTTP/TLS support, container demuxers, codec parsers, audio decoders, `avenc_aac`, and `x264enc` depending on the source and configuration.

`HDRToSDR` additionally requires the custom `hdrtonemap` GStreamer element. It is included in the embedded Windows runtime. A system/custom runtime must provide a compatible plugin through its normal plugin directory or `GST_PLUGIN_PATH`.

### Installing GStreamer

Installation is required for Linux/macOS gst builds and Windows gst builds that do not embed or ship `gst-lib`.

**Debian / Ubuntu**

```bash
sudo apt-get update

sudo apt-get install -y --no-install-recommends \
  libgstreamer1.0-0 \
  libgstreamer-plugins-base1.0-0 \
  gstreamer1.0-plugins-base \
  gstreamer1.0-plugins-good \
  gstreamer1.0-plugins-bad \
  gstreamer1.0-plugins-base-apps \
  gstreamer1.0-plugins-ugly \
  gstreamer1.0-libav \
  gstreamer1.0-tools \
  ocl-icd-libopencl1 \
  ca-certificates
```

Package roles:

| Package group | Purpose |
| --- | --- |
| `libgstreamer1.0-0`, `libgstreamer-plugins-base1.0-0` | Core and GstApp libraries loaded by TorrServer |
| `gstreamer1.0-plugins-base`, `gstreamer1.0-plugins-base-apps` | Base elements and `gst-discoverer-1.0` |
| `gstreamer1.0-plugins-good` | HTTP source, Matroska/WebM demuxing, and common media elements |
| `gstreamer1.0-plugins-bad` | Modern codec parsers, timestamp helpers, and additional format support |
| `gstreamer1.0-plugins-ugly` | `x264enc` CPU fallback for H.264 transcoding |
| `gstreamer1.0-libav` | `avenc_aac` and FFmpeg-based codec support |
| `gstreamer1.0-tools` | `gst-inspect-1.0` diagnostics |
| `ocl-icd-libopencl1` | OpenCL loader for optional GPU HDR tone mapping; CPU fallback is used when unavailable |
| `ca-certificates` | TLS certificate validation for HTTPS sources |

The selected repository must provide GStreamer 1.22 or newer.

Hardware encoders are optional. They also require a compatible vendor driver and GStreamer encoder plugin; TorrServer tests the available candidates and falls back to `x264enc` when none can start.

**Fedora / RHEL / Rocky / AlmaLinux**

```bash
sudo dnf install -y \
  gstreamer1 \
  gstreamer1-tools \
  gstreamer1-plugins-base \
  gstreamer1-plugins-base-tools \
  gstreamer1-plugins-good \
  gstreamer1-plugins-bad-free \
  gstreamer1-plugins-ugly-free \
  gstreamer1-libav \
  ocl-icd \
  ca-certificates
```

`gstreamer1-plugins-base-tools` provides `gst-discoverer-1.0`. Full `x264enc` and libav support may require [RPM Fusion](https://rpmfusion.org/) or the equivalent repository for the distribution.

**Arch Linux**

```bash
sudo pacman -S --needed \
  gstreamer \
  gst-plugins-base \
  gst-plugins-good \
  gst-plugins-bad \
  gst-plugins-ugly \
  gst-libav \
  ocl-icd \
  ca-certificates
```

**macOS**

Install the [official GStreamer Runtime](https://gstreamer.freedesktop.org/download/#macos), or use the current Homebrew formula, which includes the GStreamer plugin sets:

```bash
brew install gstreamer
```

Set `GSTPath` when auto-detection cannot find the installation. Common roots are `/Library/Frameworks/GStreamer.framework/Versions/1.0`, `/opt/homebrew`, and `/usr/local`.

**Windows**

The embedded `TorrServer-gst-windows-amd64.exe` needs no separate GStreamer installation. For a dynamic build, install the **MinGW x86_64 Runtime** from [gstreamer.freedesktop.org](https://gstreamer.freedesktop.org/download/#windows). The runtime installer is sufficient; development files are not required to run TorrServer.

The default root is:

```text
C:\Program Files\gstreamer\1.0\mingw_x86_64
```

Alternatively, place the same runtime layout in `gst-lib/` beside the executable. Do not mix an MSVC installation path with the MinGW libraries or the bundled MinGW `hdrtonemap` plugin.

### Verifying the installation

Check the runtime and discoverer first:

```bash
gst-inspect-1.0 --version
gst-discoverer-1.0 --version
gst-inspect-1.0 souphttpsrc
gst-inspect-1.0 matroskademux
gst-inspect-1.0 mp4mux
gst-inspect-1.0 appsink
gst-inspect-1.0 avenc_aac
gst-inspect-1.0 x264enc
```

`avenc_aac` is needed when the selected audio is not already AAC. `x264enc` is the CPU fallback when video transcoding is enabled. Check `gst-inspect-1.0 hdrtonemap` only when `HDRToSDR` is required.

The shell commands require the GStreamer `bin` directory on `PATH`. For an embedded Windows build, use `GET /gst/echo` or the **GStreamer** settings tab instead. The health response reports `found`, `available`, and `works` for the native runtime, `gst-discoverer`, HDR tone mapping, and the embedded runtime where applicable.

### Web UI configuration

1. Open **Settings**.
2. Enable **PRO mode**.
3. Open the **GStreamer** tab.
4. Adjust options and click **Save GStreamer Settings**.

Settings are stored separately from the main BitTorrent settings and take effect for new tasks. Existing tasks keep the configuration with which they were created.

### `settings.json` block

GStreamer options are stored under the top-level `gstreamer` key in `settings.json`. Legacy keys `gst` and `GStreamer` are still read on load; saving from the web UI or API writes the `gstreamer` key.

Example (`settings.json`):

```json
{
  "gstreamer": {
    "GSTVersion": 1.22,
    "GSTPath": "",
    "Source": "stream",
    "MaxTasks": 0,
    "InactiveMinutes": 5,
    "AACBitrateKbps": 256,
    "AACChannels": 0,
    "AACSamplerate": 0,
    "SegmentSeconds": 6,
    "SegmentDiff": 20,
    "Subtitles": true,
    "TranscodeH264": false,
    "TranscodeH265": false,
    "TranscodeAV1": false,
    "TranscodeVP9": false,
    "TranscodeVP8": false,
    "TranscodeAVI": false,
    "HDRToSDR": false,
    "HardwareAcceleration": true,
    "UseGPU": true,
    "X264Ultrafast": false,
    "VideoBitrate": 10000
  }
}
```

On Windows, platform defaults use `GSTVersion: 1.28` and `GSTPath: "C:\\Program Files\\gstreamer\\1.0\\mingw_x86_64"`. Other platforms default to version `1.22` and an empty path.

| Field | Description |
| --- | --- |
| `GSTVersion` | Pipeline compatibility fallback, minimum `1.22`. A successfully detected runtime version takes precedence. |
| `GSTPath` | GStreamer installation root. Empty uses platform auto-detection. |
| `Source` | `stream` accepts any info hash; `play` requires a torrent already listed in TorrServer. |
| `MaxTasks` | Maximum concurrent GStreamer tasks. `0` is unlimited; the least recently active task is removed when the limit is exceeded. |
| `InactiveMinutes` | Freeze and release an idle pipeline after this timeout. The task is removed 20 minutes later if it stays inactive. |
| `AACBitrateKbps` | Bitrate for non-AAC audio transcoding. It is doubled for more than two channels. |
| `AACChannels` | Output channels for non-AAC audio. `0` uses the source value, clamped to 1-8; fallback is 2. |
| `AACSamplerate` | Output sample rate for non-AAC audio. `0` selects the nearest supported source rate; fallback is 48000 Hz. |
| `SegmentSeconds` | Target HLS duration. Copy mode uses Matroska Cue boundaries when available. |
| `SegmentDiff` | Keyframe alignment tolerance in copy mode. `0` disables the limit. |
| `Subtitles` | Expose supported embedded text subtitles as segmented WebVTT. Bitmap subtitles are not converted. |
| `TranscodeH264` | Convert H.264 video to H.264 instead of copying it. |
| `TranscodeH265` | Convert H.265/HEVC video to H.264 instead of copying it. |
| `TranscodeAV1` | Convert AV1 video to H.264 instead of copying it. |
| `TranscodeVP9` | Convert VP9 video to H.264 instead of copying it. |
| `TranscodeVP8` | Allow VP8 input and convert it to H.264. VP8 is rejected when disabled. |
| `TranscodeAVI` | Allow AVI input and convert its video to H.264. AVI is rejected when disabled. |
| `HDRToSDR` | Tone-map detected PQ/HLG HDR to SDR and convert the video to H.264. Requires `hdrtonemap`. |
| `HardwareAcceleration` | Use a tested hardware H.264 encoder when available; otherwise use `x264enc`. |
| `UseGPU` | Allow GPU video encoding and HDR processing. CPU fallbacks are used when unavailable. |
| `X264Ultrafast` | Use the x264 `ultrafast` preset instead of `veryfast`, reducing CPU load at the cost of compression efficiency. |
| `VideoBitrate` | Target H.264 video bitrate in kbps when video transcoding is active. |

With all `Transcode*` options disabled, H.264, H.265/HEVC, AV1, and VP9 video is copied. AAC audio is also copied; other audio codecs are decoded and encoded to AAC.

### API

**Settings** (requires authentication when `--httpauth` is enabled; read/write only in `-gst` builds):

- `GET /gst/settings` — on `-gst` builds: `built_in`, current config, and platform defaults; on standard builds: `{ "built_in": false }` only
- `POST /gst/settings` — update or reset config (`404` on standard builds)

  ```json
  { "action": "set", "config": { "GSTVersion": 1.22, "Source": "stream" } }
  ```

  Reset to defaults:

  ```json
  { "action": "def" }
  ```

**Streaming** (available in `-gst` builds):

| Endpoint | Description |
| --- | --- |
| `GET /gst/echo` | GStreamer / gst-discoverer health check |
| `GET /gst/:hash/probe` | Probe torrent file metadata (`index`, `id`, or `fileID` query); successful probes are cached for one hour |
| `GET /gst/:hash/master.m3u8` | Create/reuse a task and return the HLS master playlist (`index`, `audio`, and optional `seconds` query) |
| `GET /gst/:hash/video.m3u8` | HLS media playlist referenced by the master playlist |
| `GET /gst/:hash/init.mp4` | Initialization segment |
| `GET /gst/:hash/seg/*segment` | Media segment |
| `GET /gst/:hash/subs/:track.m3u8` | Segmented WebVTT subtitle playlist |
| `GET /gst/:hash/subs/:track/:segment.vtt` | WebVTT subtitle segment |
| `GET /gst/:hash/heartbeat` | Keep the task and its torrent active; returns torrent state details |
| `GET /gst/remove` | Dispose the task and drop the torrent cache (`hash` or `id` query) |

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
