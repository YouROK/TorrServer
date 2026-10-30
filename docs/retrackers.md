# Retrackers

When adding a torrent, TorrServer can modify announce trackers according to **Settings → Additional → Retrackers**:

| Mode | Behavior |
| ------ | ---------- |
| Don't add | Leave magnet/file trackers unchanged |
| Add (default) | Append the configured default/remote list |
| Remove | Clear trackers from the torrent |
| Replace | Replace with the configured default/remote list |

Related settings (same Web UI section, also via `POST /settings`):

- **`TrackersListURL`** — optional custom remote list URL. Leave **empty** to use the built-in ngosang `trackers_best_ip.txt` mirrors (tried in order):
  1. `https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best_ip.txt`
  2. `https://ngosang.github.io/trackerslist/trackers_best_ip.txt`
  3. `https://cdn.jsdelivr.net/gh/ngosang/trackerslist@master/trackers_best_ip.txt`
  4. `https://raw.githack.com/ngosang/trackerslist/master/trackers_best_ip.txt`
  
  If set, the custom URL is tried **first**, then the mirrors (duplicates are skipped). Failed/timed-out fetches fall back to the next URL, then to `DefaultTrackers` if all fail (5s timeout per URL).
- **`DefaultTrackers`** — local announce URLs, one per line (`udp`/`http`/`https`/`wss`; `#` comments allowed). Used alone when all remote fetches fail, or merged after a successful remote fetch.

Optional file overlay (always appended when present): put `trackers.txt` in the config directory (`--path` / `-d`), next to `config.db`. Only lines starting with `udp` or `http` are read from that file.
