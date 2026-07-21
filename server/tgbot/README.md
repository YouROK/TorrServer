# TorrServer Telegram Bot

[![GitHub License](https://img.shields.io/github/license/YouROK/TorrServer)](https://github.com/YouROK/TorrServer/blob/master/LICENSE)
[![TorrServer Integrated](https://img.shields.io/badge/TorrServer-integrated-blue)](https://github.com/YouROK/TorrServer)

## Introduction

Telegram bot for managing [TorrServer](https://github.com/YouROK/TorrServer) — add torrents, stream, search, and control the server directly from Telegram.

## Features

- Torrent management — add, remove, drop, list via magnet, hash, or `torrs://`
- Export & import — magnets list; import multiple from text
- Streaming — playback links, M3U playlists, preload
- Search — RuTor and Torznab with one-click add
- Inline mode — `@botname` in any chat: list torrents or search
- Status & snake — real-time status, cache visualization
- File operations — browse files, download to Telegram
- FFprobe — media metadata via `/ffp`
- Localization — Russian and English
- Admin — shutdown, settings, presets (whitelist users only)

## Getting Started

### Enable the Bot

Start TorrServer with a Telegram bot token:

```bash
TorrServer --tg YOUR_BOT_TOKEN
```

Or use `-T`:

```bash
TorrServer -T YOUR_BOT_TOKEN
```

Create a bot via [@BotFather](https://t.me/BotFather) to get the token.

### Configuration

Config file `tg.cfg` (JSON) in the TorrServer data directory:

| Field      | Description |
|------------|-------------|
| `HostTG`   | Telegram API URL (default: `https://api.telegram.org`) |
| `HostWeb`  | Base URL for stream links (auto-detected if empty) |
| `Socks5`   | Optional SOCKS5 for reaching Telegram (e.g. `127.0.0.1:1080`, `socks5://user:pass@host:port`) if direct access to `api.telegram.org` is blocked or times out |
| `WhiteIds` | Allowed Telegram user IDs. Empty = allow everyone for normal commands, but **no admins** (`/settings`, `/shutdown`, `/preset` require an ID in this list). |
| `BlackIds` | Blocked user IDs |

**Admin note:** put your Telegram user ID (from `/start` / `/help`) into `WhiteIds`, otherwise admin commands never work.

### Mini App (optional)

Set `HostWeb` to a public **HTTPS** URL of TorrServer (reverse proxy / tunnel). Then:

- Bot menu **Open Web** launches the existing web UI inside Telegram
- Chat menu button (next to the text field) opens the same Mini App when HTTPS is configured
- Valid Telegram `initData` can satisfy HTTP Basic auth for whitelisted users (or any user if `WhiteIds` is empty)

HTTP `HostWeb` (LAN IP) still works for stream links, but Telegram will not open it as a Mini App.

Example:

```json
{
  "HostTG": "https://api.telegram.org",
  "HostWeb": "http://192.168.1.100:8090",
  "Socks5": "127.0.0.1:1080",
  "WhiteIds": [123456789],
  "BlackIds": []
}
```

If your network cannot connect to Telegram’s API directly, run a local SOCKS5 proxy (for example [sing-box](https://github.com/SagerNet/sing-box), v2ray, or `ssh -D`) and set `Socks5` to its address.

## Commands

Slash menu (`/`) shows **primary** commands only: `/start`, `/help`, `/list`, `/add`, `/search`, `/more`, `/cancel`, `/lang`, plus admin `/settings`, `/preset`, `/shutdown`. Everything else still works if typed; use **⋯ More** / `/more` for the hub.

### Reply keyboard

| Button | Action |
|--------|--------|
| Library | `/list` hub (one message) |
| Search | Ask for query (next message), or `/search <query>` |
| Status | `/stat` |
| Add | Hint to paste magnet/hash |
| More | Inline hub: Library / Tools / Links / Admin / Help / Open Web |

### Core

| Command | Description |
|---------|-------------|
| `/help`, `/start`, `/id` | Help and user ID |
| `/more` | Extra actions hub |
| `/list` | Library hub |
| `/add <link>` | Add torrent (magnet, hash, torrs://) |
| `/clear` | Remove all (with confirmation) |
| `/hash [N]` | Show info hashes |
| `/cancel` | Cancel pending settings/preset/search input |
| `/lang [RU\|EN]` | Language |

### Management

| Command | Description |
|---------|-------------|
| `/remove <hash\|N>` | Remove torrent |
| `/drop <hash\|N>` | Disconnect (keep in DB) |
| `/set <hash\|N> <title>` | Set title |
| `/status [hash\|N]` | Status with refresh/stop |
| `/cache <hash\|N>` | Cache stats |
| `/preload <hash\|N> <index>` | Preload file |

### Links & Playback

| Command | Description |
|---------|-------------|
| `/link`, `/play` | Stream URL |
| `/m3u`, `/m3uall` | M3U playlist |

### Search

| Command | Description |
|---------|-------------|
| `/search <query>` | RuTor + Torznab (all sources) |
| `/rutor <query>` | RuTor only |
| `/torznab <query> [index]` | Torznab indexers |

### Other

| Command | Description |
|---------|-------------|
| `/export`, `/import` | Export/import magnets |
| `/categories` | List categories |
| `/server`, `/stats`, `/stat` | Server info |
| `/viewed` | Viewed files |
| `/ffp <hash\|N> <id> [json]` | FFprobe metadata |
| `/speedtest [size]` | Download test (1–100 MB) |
| `/snake [hash\|N] [cols] [rows]` | Cache visualization |
| `/queue` | Telegram upload queue |
| `/echo` | Version |
| `/db` | DB dump |

### Admin Only

| Command | Description |
|---------|-------------|
| `/shutdown` | Shut down server |
| `/settings` | Interactive settings menu (sub-pages: Search, Network, Other, Cache, Paths, Storage) |
| `/preset <name>` | Apply named preset: `performance`, `storage`, `streaming`, `low`, `default` |
| `/preset <key> <value> ...` | Apply key-value pairs: `cache 256`, `preload 50`, `conn 100`, etc. |

**Preset examples:**
- `/preset performance` — max cache, high preload, no limits
- `/preset cache 256 preload 50` — set cache 256 MB and preload 50%
- `/preset cache 512 conn 100 down 0 up 0` — multiple values

**Preset keys:** `cache`, `preload`, `readahead`, `conn`, `timeout`, `port`, `down`, `up`, `retr`, `responsive`, `cachedrop`

## Inline Mode

Type `@YourBotName` in any chat:

- **Empty, "list", or "play"** — torrents with play links
- **2+ characters** — search RuTor + Torznab

## Text Input

Paste as plain message to add torrent:

- `magnet:?xt=urn:btih:...`
- `torrs://...`
- 40-char info hash

Reply to file list with `2-12` to download files 2–12 to Telegram.

## Security

- **Whitelist** — restrict to specific user IDs
- **Blacklist** — block user IDs
- **Admin** — when whitelist is used, admin = whitelisted users
- **Settings** — sensitive values masked in `/settings`

## Dependencies

- [telebot v4](https://gopkg.in/telebot.v4) — Telegram Bot API
- [go-humanize](https://github.com/dustin/go-humanize)
- [go-ffprobe](https://gopkg.in/vansante/go-ffprobe.v2)
