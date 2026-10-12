# TorrServer Telegram Bot

[![GitHub License](https://img.shields.io/github/license/YouROK/TorrServer)](https://github.com/YouROK/TorrServer/blob/master/LICENSE)
[![TorrServer Integrated](https://img.shields.io/badge/TorrServer-integrated-blue)](https://github.com/YouROK/TorrServer)

## Introduction

Telegram bot for managing [TorrServer](https://github.com/YouROK/TorrServer) — add torrents, stream, search, and control the server directly from Telegram.

## Features

- Torrent management — add, remove, drop, list via magnet, hash, or `torrs://` (title/category/poster preserved)
- Categories — movie / tv / music / other on add, `/setcat`, `/list [category]`
- Streaming — short `/play` and long `/stream` URLs, M3U (optional category filter). File list copies **Play** (short URL). `/link` copies Hash / Play / Magnet, then **Stream** on a second row when the long URL is ≤256 characters; `/next` copies Play and Stream when it fits. Telegram URL buttons would open the in-app browser, which cannot play a stream.
- Next unwatched — `/next` picks the next TV episode from filenames and viewed marks
- Export & import — magnets list; import multiple from text
- Search — RuTor and Torznab with one-click add
- Inline mode — `@botname` in any chat: library results open the torrent card (`t.me/<bot>?start=t_<hash8>`); search adds or pastes a magnet
- Status & snake — real-time status, cache visualization
- File operations — browse files, download to Telegram
- FFprobe — media metadata via `/ffp`
- Localization — Russian and English
- Deep links — `t.me/<bot>?start=list`, `list_tv`, `next`, `t_<hash8>`, `add`, `search_<query>`
- Copy buttons — Hash / Play / Magnet on torrent cards; `/link` adds Stream on its own row when the long URL fits (otherwise the message shows `copy_too_long`); file rows copy Play only; `/next` copies Play and Stream when it fits; Copy M3U on playlist messages
- Players — paste the copied `/play` URL into VLC, Infuse, or another player. iOS Infuse/VLC/SenPlayer buttons live in the Mini App (HTTPS `HostWeb`), not as Telegram URL buttons (custom schemes are rejected)
- Posters — torrent cards send the poster photo when the torrent has an HTTP(S) poster
- Admin — shutdown, settings, presets (whitelist users only)

## Getting Started

### Enable the Bot

Start TorrServer with a Telegram bot token:

```bash
TorrServer --tgtoken YOUR_BOT_TOKEN
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
| `HostTG`   | Telegram API URL (default: `https://api.telegram.org`). Point this at a local [telegram-bot-api](https://github.com/tdlib/telegram-bot-api) to upload files up to 2 GB; the official API allows 50 MB. |
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

### Deep links

`https://t.me/<YourBot>?start=<payload>` (payload max 64 characters: `A–Z a–z 0–9 _ -`):

| Payload | Action |
| ------- | ------ |
| `list` | Library hub |
| `list_tv` / `list_movie` / `list_music` / `list_other` | Filtered library |
| `next` | Next unwatched TV episode |
| `t_<8hex>` | Open torrent whose info hash starts with those hex digits |
| `add` | Prompt to paste a magnet |
| `search_<query>` | Search (`_` becomes a space) |

With HTTPS `HostWeb`, the chat menu button opens the Mini App (`/?tg=1`).

## Commands

Slash menu (`/`) shows **all ordinary commands** (library, search, playback, tools, …). Admin commands (`/settings`, `/preset`, `/shutdown`) are registered only for `WhiteIds` (Telegram `chat` scope in the private chat). Everyone else does not see them in `/`. Typed admin commands still hit the handler and return `admin_only` when the user is not on the whitelist.

Group chats get a minimal list (`/start`, `/help`). Everything else still works if typed; use **⋯ More** / `/more` for the hub.

On first `/start` with no saved language, the bot uses Telegram `language_code` (`en*` → English, otherwise Russian) and stores it in `tg_langs.json`. `/lang` refreshes the reply keyboard **and** that user’s slash list.

### Chat menu button

- HTTPS `HostWeb`: the button next to the text field is the **Mini App** (same web UI). Slash commands remain available by typing `/`.
- Otherwise the button is the **commands** list (`MenuButtonCommands`).

### Reply keyboard

Persistent (`is_persistent`) with an input placeholder. Same five buttons; Mini App is not duplicated on the keyboard.

| Button | Action |
| -------- | -------- |
| Library | `/list` hub (one message) |
| Search | Ask for query (next message), or `/search <query>` |
| Status | `/stat` |
| Add | Hint to paste magnet/hash |
| More | Inline hub: Library / Tools / Help / Language / Open Web (Admin if you are on `WhiteIds`) |

Tools that need a torrent (snake, preload, cache, ffprobe) set a short-lived pick: the library opens and choosing an item runs that tool. `/cancel` clears it.

### Core

| Command | Description |
| --------- | ------------- |
| `/help`, `/start`, `/id` | Help and user ID |
| `/more` | Extra actions hub |
| `/list [category]` | Library hub (filter: movie, tv, music, other) |
| `/add <link>` | Add torrent (magnet, hash, torrs://); asks for category if unset |
| `/clear` | Remove all (with confirmation) |
| `/hash [N]` | Show info hashes |
| `/cancel` | Cancel pending settings, preset, search, or tool pick |
| `/lang [RU\|EN]` | Language |

### Management

| Command | Description |
| --------- | ------------- |
| `/remove <hash\|N>` | Remove torrent |
| `/drop <hash\|N>` | Disconnect (keep in DB) |
| `/set <hash\|N> <title>` | Set title (keeps category and poster) |
| `/setcat <hash\|N> <movie\|tv\|music\|other\|->` | Set or clear category |
| `/status [hash\|N]` | Status with refresh/stop |
| `/cache <hash\|N>` | Cache stats |
| `/preload <hash\|N> <index>` | Preload file |

### Links & Playback

| Command | Description |
| --------- | ------------- |
| `/link`, `/play` | Short `/play/{hash}/{id}` and long `/stream` URLs (copy buttons) |
| `/m3u`, `/m3uall [category]` | M3U playlist (copy button) |
| `/next [query\|hash] [category]` | Next unwatched TV episode (default category `tv`; copy short play URL) |

### Search

| Command | Description |
| --------- | ------------- |
| `/search <query>` | RuTor + Torznab (all sources) |
| `/rutor <query>` | RuTor only |
| `/torznab <query> [index]` | Torznab indexers |

### Other

| Command | Description |
| --------- | ------------- |
| `/export`, `/import` | Export/import magnets |
| `/categories` | Category counts + filter buttons |
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
| --------- | ------------- |
| `/shutdown` | Shut down server |
| `/settings` | Interactive settings: one value per line, two buttons per row, localized toggle labels (Search, Network, Other, Cache, Paths, Storage) |
| `/preset <name>` | Apply named preset: `performance`, `storage`, `streaming`, `low`, `default` |
| `/preset <key> <value> ...` | Apply key-value pairs: `cache 256`, `preload 50`, `conn 100`, etc. |

**Preset examples:**

- `/preset performance` — max cache, high preload, no limits
- `/preset cache 256 preload 50` — set cache 256 MB and preload 50%
- `/preset cache 512 conn 100 down 0 up 0` — multiple values

**Preset keys:** `cache`, `preload`, `readahead`, `conn`, `timeout`, `port`, `down`, `up`, `retr`, `responsive`, `cachedrop`

## Inline Mode

In **other chats**, type `@YourBotName` to open the result panel. In a **private chat with the bot** (especially iOS), use the **Inline search** button (`switch_inline_query_current_chat`) instead of typing `@bot` as a normal message. If that text is sent anyway, the bot strips `@username` and searches the rest (`аватар`, not `@bot аватар`).

- **Empty, "list", or "play"** — torrents with a deep link to the torrent card (`t.me/<bot>?start=t_<hash8>`, paginated; JPEG/PNG/WebP posters as thumbnails when set)
- **1+ characters** — search RuTor + Torznab (paginated)
- In a **private chat with the bot**, choosing a search result adds the torrent to the library. In other chats it pastes the magnet.

The library hub and the Search prompt include the **Inline search** button.

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
- **Settings** — sensitive values masked in `/settings`; home uses a 2-column nav and Back to More

## Dependencies

- [telebot v4](https://gopkg.in/telebot.v4) — Telegram Bot API
- [go-humanize](https://github.com/dustin/go-humanize)
- [go-ffprobe](https://gopkg.in/vansante/go-ffprobe.v2)
