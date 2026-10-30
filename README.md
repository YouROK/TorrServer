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
  <a href="https://github.com/YouROK/TorrServer/actions/workflows/ci.yml" rel="nofollow">
    <img src="https://img.shields.io/github/actions/workflow/status/YouROK/TorrServer/ci.yml?logo=Github" alt="CI" />
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
- Integration with other apps through [API](docs/api.md)
- [Torznab](docs/torznab.md) search (Jackett, Prowlarr, and similar indexer managers)
- Cross-browser modern web interface
- Optional DLNA server
- Optional [GStreamer](docs/gstreamer.md) HLS remuxing and transcoding (`-gst` builds from release 141.10)
- Native [MCP](server/mcp/README.md) server for AI agents (OpenClaw, Hermes, and other MCP clients)
- [External subtitles](docs/subtitles.md) (SRT/ASS/SSA/VTT to WebVTT) for web players and external players

## Quick start

Linux (interactive installer, systemd service):

```bash
curl -fsSL https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh -o installTorrServerLinux.sh && chmod 755 installTorrServerLinux.sh && sudo bash ./installTorrServerLinux.sh
```

Docker:

```bash
docker run --rm -d --name torrserver -p 8090:8090 ghcr.io/yourok/torrserver:latest
```

Windows, macOS, NAS and others: download the binary for your platform from the [releases](https://github.com/YouROK/TorrServer/releases) page and run it.

Then open <http://127.0.0.1:8090> in a browser. [Installation](docs/installation.md) covers every platform and the installer's options.

## Documentation

| Topic | What is in it |
|---|---|
| [Installation](docs/installation.md) | Windows, Linux installer script, macOS, IOCage, NAS |
| [Docker](docs/docker.md) | `docker run`, environment variables, Compose |
| [Server arguments](docs/cli.md) | Command-line flags |
| [HTTPS](docs/https.md) | Modes, certificates, Let's Encrypt on a LAN, Android TV box |
| [Reverse proxy](docs/reverse-proxy.md) | Caddy and nginx in front of TorrServer |
| [Authentication](docs/auth.md) | HTTP Basic auth and `accs.db` |
| [Web Application Firewall](docs/waf.md) | IP and Referer/Origin rules |
| [Retrackers](docs/retrackers.md) | Adding trackers to torrents |
| [Memory use](docs/memory.md) | Reading memory numbers on Linux, `--pprof` |
| [Torznab](docs/torznab.md) | Jackett and Prowlarr search |
| [GStreamer](docs/gstreamer.md) | HLS remux and transcoding builds |
| [API](docs/api.md) | HTTP API and Swagger |
| [MCP](server/mcp/README.md) | Tools for AI agents |
| [Telegram bot](server/tgbot/README.md) | Bot setup and commands |
| [External subtitles](docs/subtitles.md) | Supported formats and layouts |
| [Smart TV](docs/smart-tv.md) | Media Station X |
| [Development](docs/development.md) | Running from source, web UI, building, [BUILD.md](docs/BUILD.md) |
| [Credits and support](docs/credits.md) | Donate, thanks |

The same list with more detail is in [docs/README.md](docs/README.md).
