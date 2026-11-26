# TorrServer MCP

[![GitHub License](https://img.shields.io/github/license/YouROK/TorrServer)](https://github.com/YouROK/TorrServer/blob/master/LICENSE)
[![TorrServer Integrated](https://img.shields.io/badge/TorrServer-integrated-blue)](https://github.com/YouROK/TorrServer)

## Introduction

Native [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server built into TorrServer. AI agents such as [OpenClaw](https://docs.openclaw.ai/tools/mcp) and [Hermes](https://hermes-agent.nousresearch.com/docs/user-guide/features/mcp) can discover, add, and manage torrents, and return play links for the next unwatched TV episode.

The MCP interface calls the same torrent helpers as the REST API and Telegram bot. It does not replace them.

## Endpoint

Streamable HTTP on the existing web port (default **8090**):

```
http://<host>:8090/mcp
```

With `--ssl`, use `https://<host>:<sslport>/mcp`. Startup logs:

```
MCP Streamable HTTP endpoint /mcp
```

When HTTP Basic auth is enabled (`-a` / `--httpauth`, or Docker `TS_HTTPAUTH=1`), MCP uses the same `accs.db` credentials as `POST /torrents`. Send `Authorization: Basic …`.

Play URLs returned by tools are ordinary HTTP links for VLC, mpv, or a browser.

## Connect an agent

### OpenClaw

```json
{
  "mcp": {
    "servers": {
      "torrserver": {
        "url": "http://127.0.0.1:8090/mcp",
        "transport": "streamable-http"
      }
    }
  }
}
```

With auth:

```json
{
  "mcp": {
    "servers": {
      "torrserver": {
        "url": "http://127.0.0.1:8090/mcp",
        "transport": "streamable-http",
        "headers": {
          "Authorization": "Basic <base64-user-pass>"
        }
      }
    }
  }
}
```

CLI:

```bash
openclaw mcp add torrserver --url http://127.0.0.1:8090/mcp --transport streamable-http
openclaw mcp doctor torrserver --probe
```

### Hermes

```yaml
mcp_servers:
  torrserver:
    url: "http://127.0.0.1:8090/mcp"
    headers:
      Authorization: "Basic <base64-user-pass>"
```

Reload MCP in the active session after editing config (`/reload-mcp`).

## Tools

| Tool | Purpose |
|------|---------|
| `get_server_info` | Version, base URL, categories, search/auth flags |
| `list_torrents` | Library list; optional `category` and `search` |
| `get_torrent` | One torrent: files, viewed flags, season/episode, play URLs |
| `add_torrent` | Magnet, infohash, http(s) `.torrent` URL, or `torrs://`; waits for metadata; saves to DB by default |
| `update_torrent` | Title, category, poster |
| `remove_torrent` | Permanent remove (memory, DB, disk cache) |
| `drop_torrent` | Unload from the client; keep in DB |
| `get_play_url` | Stream URL for a file (`/stream/...?play` and `/play/{hash}/{id}`) |
| `get_playlist_url` | M3U for one torrent or the whole library |
| `list_viewed` / `mark_viewed` / `unmark_viewed` | Per-file watch marks (optional timecode) |
| `get_next_unwatched` | Next unwatched TV episode from filenames + viewed marks |
| `search_torrents` | RuTor and/or Torznab, if enabled in settings |

Not exposed: wipe-all, shutdown, settings mutation, `.torrent` multipart upload.

Library categories: `movie`, `tv`, `music`, `other` (empty = uncategorized).

## Next unwatched episode

TorrServer has no series database. `get_next_unwatched` parses video filenames (`S01E05`, `1x05`, `Season 1/Episode 05`, Russian `сезон` / `серия`) and skips files already in viewed marks.

- Default `category` is `tv`. Optional `query` matches show title; optional `hash` limits to one torrent.
- Returns a play URL, `SxxEyy` code, and remaining unwatched count.
- If every matching episode is viewed, returns `caught_up: true` and the last watched file.

After the user watches something, call `mark_viewed` so the next request skips that file.

## Security

- Same HTTP Basic auth as the REST API when `-a` is set.
- Mutating tools honor read-only DB mode (`-r`).
- MCP does not expose shutdown or wipe.
