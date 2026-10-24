# TorrServer agent instructions

Module path is `server`. Go commands run from `server/`. This checkout has no `internal/` tree and no root `go.mod`. Do not invent a second engine.

`server/go.mod` is `go 1.25.7`. CI installs `GO_VERSION` from `.github/versions.env` (currently 1.27.3) and keeps the module line aligned with `GO_ANDROID_VERSION`. `server/.golangci.yml` pins `run.go` to `1.25.7`; bump that line in the same change as the module.

## What this process is

TorrServer streams BitTorrent payloads over HTTP Range. The hot path is `GET/HEAD /stream` and `/play` → `torr.Torrent.Stream` → `torrstor.Reader` → piece cache → `anacrolix/torrent`. Optional surfaces (DLNA, Bonjour, FUSE, WebDAV, Telegram, MCP, GStreamer) must not add work to that path.

Out of scope unless the task names them: embedded JS, a plugin runtime, and an FFmpeg transcode pipeline. Remux/transcode already lives behind the `gst` build tag in `server/gstreamer`.

## Dependency forks

```
github.com/anacrolix/torrent v1.59.1 => github.com/lampac-nextgen/torrent v1.2.35
github.com/anacrolix/upnp    v0.1.4  => github.com/tsynik/upnp v0.1.5
github.com/anacrolix/utp     v0.2.0  => github.com/lampac-nextgen/utp v0.2.1
```

Import the upstream paths. The replace directives supply the forks.

Comments marked `NE` are upstream APIs this fork does not have (`DisableWebtorrent`, `HeaderObfuscationPolicy`, storage `Capacity`, directory file storage). Comments marked `OE` are the behavior this fork actually runs. Do not call `NE` methods. The seeder in tests cannot use a directory backend; `server/torr/offlinetest` serves bytes from `memStore`.

`configure()` in `server/torr/btserver.go` spoofs `qBittorrent/4.3.9` / peer id prefix `-qB4390-`. `UpnpID` stays `TorrServer/<version>`. `TotalHalfOpenConns` is hardcoded to 500. Do not "correct" the peer id to TorrServer's own name without a reason: swarms treat the qB prefix differently.

## UI boundary

The React UI is `web/` (Yarn, Vite). The binary serves a generated copy under `server/web/pages/template/pages`. Edit `web/src`, then rebuild (`gen_web.go`, `make webgen-clean`). Do not hand-edit the generated pages.

A new or changed string belongs in `web/src/locales`. `yarn i18n:check` requires the same flattened keys in `en`, `ru`, `ua`, `bg`, `fr`, `ro`, and `zh`, as nested objects.

When `web/src` changes, CI runs `yarn typecheck`, `yarn lint`, `yarn test`, `yarn i18n:check`, `yarn test:e2e`, and `yarn build`. From `web/`: `yarn typecheck`, `yarn lint`, `yarn test`, `yarn i18n:check`. Embed with `make webgen` (`go run gen_web.go`) or `make webgen-clean` (deletes `web/build` first). Do not pass `VITE_SERVER_HOST` pointing at localhost into a production `yarn build` — `web/vite.config.ts` refuses it.

Web client map (source of truth is `web/src`, not the embedded copy):

| Path | Role |
| --- | --- |
| `web/src/shared/hooks/useTorrentsQuery.ts` | Library poll. `queryFn` must take React Query's `signal`. `refetchIntervalInBackground` stays default false, so a hidden tab does not hit `/torrents`. The `document.hidden` branch only matters after a recompute. |
| `web/src/shared/hooks/useTorrentDetail.ts` | Details poll. Same signal rule. Mirrors rows into `TORRENTS_QUERY_KEY` via `upsertTorrentsInList`. |
| `web/src/shared/cache/useUpdateCache.ts` | `/cache` snake poll. `setTimeout` after the response (not `setInterval`). Aborts on unmount. Pauses while `document.hidden`. Active 100ms, idle 1s. Details polls on the Cache tab, the cache-map modal, and Stats (idle). The Files tab does not. |
| `web/src/shared/lib/posterPlay.ts` | Re-exports `streamPlayUrl`, `streamPreloadUrl`, `torrentPlaylistUrl` from `streamLinks.ts`. |
| `web/src/shared/lib/mediaBase.ts` | `externalLink` swaps scheme/host/port from `GET /mediabase`, then `withAuthMediaUrl`. |
| `web/src/shared/api/authCredentials.ts` | `withAuthMediaUrl` adds Basic userinfo only when `hostname` matches the page (port and scheme may differ). Do not put the password on another host. |
| `web/src/shared/api/settings.ts` | `sanitizeBTSetsForWire` before `POST /settings`. `CacheSize` is bytes on the wire. The dialog keeps megabytes in component state (`SettingsDialog`). |
| `web/src/features/settings/SettingsDialog.tsx` | Save stays disabled until `GET /settings` succeeds. `403` is read-only. |
| `web/src/features/details/FilesDataGrid.tsx` | File rows. `content-visibility: auto` so a 500-file pack does not lay out every row. |
| `web/src/index.tsx` | `refetchOnWindowFocus: false` globally. Do not add a second poller. |

API contracts the UI depends on:

- `POST /settings` with `action: set` and a nil `sets` is `400`. Read-only `set` and `def` are `403` `{"error":"Read-only mode"}` and must not call `restartServices`. `CacheSize` `0` is rewritten to 64MiB in `SetBTSets`. Integer fields reject JSON floats (`encoding/json` into `int` fails the whole body).
- `GET /stream` and `GET /stream/*fname` both exist. Preload is `/stream?link&index&preload` (no filename). Play is `/stream/<file>?link&index&play`. `link` is an infohash in the UI; a magnet in that query has to be encoded or `&dn` / `&tr` split it. The UI builds those URLs in `web/src/shared/lib/streamLinks.ts` (`encodePathSegment` for the filename, `URLSearchParams` for `link` and `index`).
- `GET /mediabase` is the external-player origin. In-page `<video>` stays on the page origin.
- `BTSets.StoreViewedInJson` is not a settings-form control. Storage backend is `GET/POST /storage/settings` (`json` or `bbolt`). Round-trip the field if it arrived on `GET /settings`.
- TMDB search URL is built in `web/src/shared/lib/torrentHelpers.ts` (`buildTmdbSearchUrl`). The server default `APIURL` is `https://api.themoviedb.org` without `/3`. The client appends `/3/search/multi`.
- Service worker: `web/vite.config.ts` `navigateFallbackDenylist` must include every API prefix (`/stream`, `/torrents`, `/cache`, `/settings`, `/echo`, `/viewed`, `/gst`, `/ffp`, `/playlistall`, `/runtime`, `/ssl`, `/mediabase`, …). `server/web/pages/template/route.go` sends `no-cache` for `index.html`, `sw.js`, and `workbox-*`, and `immutable` for `/static/`.
- MSX `/msx/proxy?url=` is an intentional open proxy (tests fetch `127.0.0.1`). Do not "close" it without an MSX replacement. `http.Client` only speaks http(s).

`POST /torrents` actions the UI sends: `list`, `get`, `add`, `set`, `drop`, `rem`, `wipe`. `drop` keeps the DB row (`IN_DB`). An empty handler return is HTTP 200.

## Package map

| Path | Role | Care |
| --- | --- | --- |
| `server/cmd` | `go-arg` CLI, `exitFunc`, CA roots | `main` calls `exitFunc(run())`. Tests replace `exitFunc`. `run` returns an int. |
| `server/server.go` | Process start/stop | `settings.EmbeddedStop = Stop`. `stopOnce` is reset at the start of `Start`. |
| `server/settings` | bbolt `config.db` + JSON files, `BTSets` | Globals. See storage section. |
| `server/torr` | `BTServer`, `Torrent`, preload, `Stream` | `bt.mu` is not reentrant. |
| `server/torr/utils` | Trackers, blocklist, rate limit, peer id | Dedicated HTTP client. Missing `blocklist` file is ignored. |
| `server/torr/storage/torrstor` | Piece cache, readers, ranges | Lock split below. |
| `server/torr/offlinetest` | Loopback torrent harness | Use this. Do not open a second client style. |
| `server/torr/state`, `storage/state` | Status DTOs | Returned to HTTP. Do not put live `*torrent.Torrent` in JSON. |
| `server/web` | Gin, listeners, TLS redirect | `serversMu`. Media routes are registered outside the auth group. |
| `server/web/api` | `/stream`, `/play`, `/torrents`, playlists, cache, viewed | Every failure after the handler accepts the request must write a status. An empty return becomes `200` with an empty body. |
| `server/web/auth` | `accs.db` JSON map, Telegram init data | `BasicAuth` sets `auth_required` and does not abort. `CheckAuth` aborts. |
| `server/web/waf` | IP/referer filter | Own `RWMutex` plus `updateMu`. |
| `server/torrfs`, `fuse`, `webdav` | Virtual tree | `FuseAutoMount` calls `os.Exit(1)` on failure. |
| `server/dlna` | anacrolix/dms | `exitFunc` on listen failure unless `settings.Embedded`. |
| `server/bonjour` | mDNS | Interface list is mockable in tests. |
| `server/netbind` | Multi-address `Listen` | DLNA and the web server share it. |
| `server/ffprobe` | External `ffprobe` binary | Mock the binary. Do not shell out to a real one in unit tests. |
| `server/gstreamer` | Optional HLS remux | Build tag `gst`. Default tests compile `bridge/bridge_stub.go` (`!gst`). |
| `server/mcp` | MCP tools over the same torrent API | No second client. |
| `server/tgbot` | Telegram | `httptest` plus `tele.Settings{URL, Offline: true}`. |
| `server/torznab`, `server/rutor` | Search | `LIVE_INDEXER=1` is the only live test. |
| `server/subs`, `torrshash`, `mimetype`, `library` | Subtitles, `torrs://` tokens, MIME, URL helpers | Pure. Prefer tests here over new helpers. |
| `server/log` | File log, 100MB truncate-by-delete, web access log | `TLogln`. No metrics bus. |
| `server/leakcheck` | `AssertNoGoroutineLeak`, `Heap` | Shared by lifecycle tests. No assertions of its own. |
| `server/mobile/torrserverkit` | iOS XCFramework host | `Embedded` must stay free of `os.Exit`. |
| `server/docs` | Generated Swagger | Excluded from `make test`. Do not edit by hand. |

## Lock order

`sync.RWMutex` is not reentrant. `RLock` then `Lock` on the same mutex deadlocks the goroutine. `server/torr/clientrace_test.go` fails the build if `NewTorrent` holds `bt.mu.RLock` across the later `bt.mu.Lock`.

Hold order when both are needed:

1. `settings.dbMigrationLock` then `settings.mu` (`server/settings/torrent.go`). `mu` is a `sync.Mutex`. `ListTorrent` takes `dbMigrationLock.RLock`, then `mu`, and holds `mu` for the whole call, including the bbolt reads (`tdb.List` / `tdb.Get`). `AddTorrent` calls `ListTorrent` first — both locks drop before it returns — then locks `mu` and rewrites every `Torrents` row with `tdb.Set`. Each `Set` is a bbolt write transaction, so `mu` stays held across those database locks. Do not call `ListTorrent` or `AddTorrent` while `mu` is already held: `ListTorrent` locks `mu` again and deadlocks. Do not take `dbMigrationLock` while `mu` is held. Do not add network I/O under `mu`.
2. `BTServer.mu` (`server/torr/btserver.go`). Guards `client`, `storage`, `torrents`, DHT cancel. `NewTorrent` drops `RLock` before `Lock`. `GetTorrent` / `ListTorrents` copy under `RLock` and return the copy. `WriteStatus` holds `RLock` for the whole `client.WriteStatus` call so `Disconnect` cannot nil the client underneath it.
3. `Torrent.muTorrent`. Snapshot `t.Torrent` under the lock and use the local pointer after. `drop` nils `t.Torrent` and `t.cache` at any time (Lampa "drop" during a stream). `WaitInfo` already does this snapshot. `Close` sets `Stat`, unlocks, deletes from `bt.torrents`, then `drop`. Do not call `bt.mu.Lock` while `muTorrent` is held.
4. `torrstor.Storage.mu`. Release it before `Cache.Close`. `CloseHash` deletes the map entry, unlocks, then closes. `Cache.Close` calls `removeCache`, which locks `Storage.mu` again.
5. Cache locks, never nested in the opposite direction: `muRemove` (evict/close) → `muReaders` → `muActive` → `muPrio`. `muTor` is only the `*torrent.Torrent` pointer. `filled` is `atomic.Int64`. `isClosed` is `atomic.Bool`.
6. `Reader.mu`. `resume` drops it around the anacrolix `Seek` because that seek runs cache cleanup, which locks `Reader.mu` again. Copy the fields you need, unlock, seek, relock. `readerOff` refuses to `Seek(0)` while `inFlight > 0` (`reader_park_test.go`).

`Cache.pieces` is filled in `Init` and then read. Do not insert or delete keys after `Init`. `Close` leaves the map in place so `GetState` can still walk it; buffers are nil and `isClosed` blocks new allocs. `activePieces` and `prioPieces` have their own mutexes so `GetState` stays O(active), not O(piece count).

`MemPiece.ReadAt` takes `Lock`, not `RLock`: it writes `Accessed`. `Release` nils the slice under the lock, unlocks, then `putPieceBuf`, so a `ReadAt` that was waiting finishes its copy before the slab is cleared and published.

`Cache.CloseReader` is safe on a nil `*Cache`: it uses `r.cache`. `Torrent.CloseReader` relies on that. `Reader.Close` alone does not unregister; only `Cache.CloseReader` deletes from `readers`, and it holds `muReaders` across `Reader.Close`. A second `clearPriority` after that raced a new reader's priorities. Do not add one.

Do not hold `BTServer.mu`, `Torrent.muTorrent`, `Storage.mu`, or `Cache.muRemove` across tracker HTTP, DHT lookup, `publicip.Get4/Get6`, disk flush, or `os.MkdirAll`. `Connect` already holds `bt.mu` across `torrent.NewClient`; do not add more I/O there.

`SetBTSets` starts a goroutine that walks `TorrentsSavePath` looking for a `.tsc` directory and mutates `BTsets` later. Do not assume the save path is stable the moment `SetBTSets` returns.

## Streaming and memory

`/stream` and `/play` are registered in `server/web/api/route.go` without `auth.CheckAuth`. `stream` still enforces auth unless the query is an unauthenticated play/m3u that passes `utils.TestLink`. `SearchWA` is what exposes `/search` and `/torznab` without auth. `/echo` is the Docker healthcheck and must stay anonymous.

`Torrent.Stream` (`server/torr/stream.go`):

- Answers `404` / `403` / `503` itself. Players treat an empty `200` as a zero-length file.
- `defer t.CloseReader(reader)` on every path after `NewReader`.
- `ResponsiveMode` calls `reader.SetResponsive()` so a read does not wait for the whole piece.
- Wraps the reader in `newBufferedStreamReader(reader, 1<<20)` and `ServeStreamContent`.

Copy path, do not add a third buffer:

- `streamCopySize` is 64KiB. `streamCopyPool` stores `*[streamCopySize]byte`, not a `[]byte` (a bare slice in `sync.Pool` boxes the header).
- `streamCopyWriter.ReadFrom` is what `http.ServeContent` actually hits. gin's writer has no `ReaderFrom`, and `io.CopyN`'s `LimitedReader` has no `WriterTo`.
- `buf[:nr]` passed to `Write` must not escape. The pool reuses the same array on the next read.
- `ServeStreamContent` nils `ResponseWriter` before `Put` so the pool does not pin the request.
- `bufferedStreamReader.Seek` subtracts `buffer.Buffered()` on `SeekCurrent`, then `Reset`. Skipping that seeks to the wrong byte.

Piece buffers (`server/torr/storage/torrstor/piecebuf.go`):

- One `sync.Pool` per exact length, stored as `*pieceSlab` in a `sync.Map`.
- `putPieceBuf` clears the full cap before publish.
- `MemPiece.Release` does not recycle when `cache.isClosed` is set, so `FreeOSMem` after drop can return the RAM. Do not "simplify" that to always `Put`.
- Never retain a sub-slice of a slab after `Release`. The next tenant of that pool owns the backing array.
- `ponytail:` on the global map is intentional. Split pools per cache only if a profile shows the map.

Eviction (`Cache.cleanPieces` / `getRemPieces` / `pieceEvictable`):

- `cleanPieces` uses `muRemove.TryLock`. `Close` holds `muRemove` across `Release`. A failed try skips this pass.
- A piece is not evictable when it sits in a merged reader range, in the begin/end keep window, in `prioPieces`, or when `Size >= piece length` but `Complete` is still false (hash in progress). Evicting that last case makes anacrolix hash a short `ReadAt` and log `unexpected error hashing piece`.
- Reader ranges are closed intervals. `mergeRange` merges only when `End >= next.Start`.
- Keep window is at least 8MiB at the start and end of the torrent when no reader range is active (`isIdInKeepWindow`). `isIdInFileBE` keeps the ends of each open file.
- Evict complete pieces first, oldest `Accessed` first. The client does not re-request chunks it already has; evicting a partial piece makes a responsive reader read zeros.
- Priorities change in `getRemPieces` / `clearPriority` / `setLoadPriority`, guarded by `muPrio`. Recalculate when a reader is created, closed, parked, or crosses a piece. Do not do it on every small `Read`.
- `priorityWatchdog` re-arms every 5s while `GetUseReaders() > 0`, and exits on `stopWatch` or `isClosed`. A `time.Sleep` instead of `select` on `stopWatch` leaks one goroutine per dropped torrent.

Reader window (`Reader.getOffsetRange`): capacity is split across active readers. `ReaderReadAHead` (default 95, clamped 5–100) is the forward fraction. `Seek(0, SeekEnd)` returns the length and does not move the playhead. ServeContent probes size that way; seeking to EOF would pull the tail of a huge file.

Idle park is `minReaderIdleTimeout` (60s), or `TorrentDisconnectTimeout` if that is larger. `checkReader` parks by `SetReadahead(0)` and `Seek(0, SeekStart)` only when `inFlight == 0`. `beginIO` increments `inFlight` and `resume`s under `Reader.mu` before the anacrolix read. A blocked read must not be seeked.

`Torrent.Close` refuses while `ReadOnly` and `GetUseReaders() > 0`.

## BitTorrent client

`BTServer.Connect` calls `shutdownLocked` before building a new client. `PrefetchTrackers` must not block startup. The tracker client is `trackersHTTPClient` in `server/torr/utils/torrent.go`: 5s timeout, 3s dial, 3s TLS, 3s headers, body capped at 256KiB, its own `Transport`. Do not use `http.DefaultClient`. `CloseIdleConnections` on the default transport closes every idle connection in the process. Refresh is every 12h; a failed refresh keeps the previous list.

`SanitizeTrackers` keeps `udp`, `http`, `https`, `ws`, `wss`, drops empty lines, comments, URLs over 2048 bytes, bad ports, and raw-IP `https`/`wss` (no SNI; the tracker client waits its full timeout per torrent). It dedupes. It does not cap the list. Raw-IP dedupe is host:port only (`ponytail:` in `acceptTracker`).

`addExtraTrackers` appends one tier. This fork starts a scraper for every URL in every tier, so one extra tier per URL is an announce storm. `RetrackersMode`: 0 append file list, 1 file plus defaults, 2 replace with file, 3 replace with file plus defaults.

Remote list: `BTSets.TrackersListURL` first, then `DefaultTrackersListURLs` (ngosang mirrors). Tests replace `defaultTrackersListURLs`. `InvalidateTrackersCache` bumps `trackersFetchGen` and starts a new prefetch. `GetDefTrackers` uses the local list until the first prefetch finishes.

DHT bootstrap is `resolveDHTBootstrap` with `dhtBootstrapTimeout` (3s) and `lookupDHTHost`. Tests replace both. Do not switch back to `dht.GlobalBootstrapAddrs`: that lookup uses `context.Background` and starts a refresher that never stops.

`ReadBlockedIP` loads `{Path}/blocklist` (P2P blocklist lines) into `config.IPBlocklist`. `configure` discards the error, so a missing file means no list. Public IPv4/IPv6 from flags are rejected when private; otherwise `publicip.Get4/Get6` runs with a 3s timeout. `privateIPBlocks` is built in `init` and panics on a bad CIDR.

Proxy (`configureProxy`): empty URL is off. Default mode is `tracker`. `full` sets `ProxyURL` and `HTTPProxy`. `peers` sets `ProxyURL` only. Anything else sets `HTTPProxy` only. Schemes: `http`, `https`, `socks4`, `socks4a`, `socks5`, `socks5h`.

`TorrentDisconnectTimeout` is seconds, default 30. `NewTorrent` caps the in-memory expiry used at create to one minute even when the setting is larger. `watch` still uses the setting.

## Settings, CLI, Docker

CLI is `alexflint/go-arg` in `server/cmd/main.go`. Defaults: HTTP `8090` (`settings.DefaultPort`), HTTPS `8091`. Data dir `-d` defaults to the working directory. Auth file is `{Path}/accs.db`. Blocklist is `{Path}/blocklist`.

Flags agents touch most often: `-p` `-i` `-d` `-l` `-w` `-r` (`--rdb`) `-a` (`--httpauth`) `-k` (`--dontkill`) `-t` `-4` `-6` `-s` (`--searchwa`) `-m` `-T` `-f` `--webdav` `--proxyurl` `--proxymode` `--ssl` `--sslport` `--sslcert` `--sslkey` `--force-https` `--http-media` `--https-only` `--pprof`.

`settings.PlainHTTPServesMedia` is false when `--force-https` is set without `--http-media`, and when `--https-only`. Players and TVs that cannot speak TLS need `--http-media`, which sends stream URLs and Basic credentials in the clear. `settings.InternalPort` is the loopback port for ffprobe and GStreamer self-requests. `LoopbackBaseURL` prefers it. Those requests must not follow the public HTTPS redirect.

Process env read by Go:

- `TS_MIGRATION_MODE` forces `SmartMigrate`.
- `TS_SETTINGS_STORAGE=json|bbolt` and `TS_VIEWED_STORAGE=json|bbolt` apply only when no settings row exists yet. Defaults: settings JSON, viewed bbolt.
- `SSL_CERT_FILE` / `SSL_CERT_DIR` on Linux. `configureCARoots` will not override a value the user already set.
- `GODEBUG=madvdontneed=1` is the container default so dropped piece RAM returns to the OS.

`docker-entrypoint.sh` maps `TS_*` onto flags. It enables TLS from `TS_SSL_ENABLE=1`. `docker-compose.yml` and `.env.example` publish `TS_EN_SSL`. Those names are not the same variable. Wiring TLS through Compose requires the entrypoint to see `TS_SSL_ENABLE`, or the flag is never passed. README documents `TS_SSL_ENABLE`.

`BTSets` defaults worth not "cleaning up": cache 64MiB, readahead 95%, preload 50%, connections 25, retrackers mode 1, responsive on, LPD on, Bonjour on, `StoreSettingsInJson` true. `CacheSize == 0` in `SetBTSets` is rewritten to 64MiB. `UseDisk` is forced off when `TorrentsSavePath` is empty.

Storage handles:

- `NewTDB` stores the open `*bolt.DB` in `globalBboltDB`. A second `InitSets` without `CloseDB` returns the existing handle.
- `NewJsonDB` uses `:=`, so the local `globalJsonDB` shadows the package variable and the package variable stays nil. `CloseDB` still assigns `globalJsonDB = nil`. Do not "fix" the shadow in a drive-by: `CloseDB`'s comment is about the bbolt handle (`Leaving them set reused a closed bbolt after Stop/Start`). A real fix assigns the package var and keeps the `CloseDB` nil, with a Stop/Start test.
- `JsonDB.CloseDB` is a no-op. Per-file mutexes live in `jsonDbLocks`.
- `CloseDB` closes the router, then nils `tdb`, `globalBboltDB`, and `globalJsonDB`.
- `ReadOnly` (`-r` / `TS_RDB`) makes `SetBTSets` return immediately and blocks storage switches.

`settings.Embedded` is the in-process iOS host. `dlna.Start` and `apihelper.Shutdown` must return or call `settings.EmbeddedStop` instead of exiting. `FuseAutoMount` still calls `os.Exit(1)`; do not use it on the embedded path.

`os.Exit` / `log.Fatal` that already exist and must stay behind a seam:

- `cmd.exitFunc`, `dlna.exitFunc` (skipped when `Embedded`).
- `torr.Shutdown` exits 0 only when not embedded, after `Disconnect` and `CloseDB`.
- `FuseAutoMount` and `cmd/preconfig_*.go` still call `os.Exit` directly. New library code does not.

Logging is `server/log`. DLNA uses `github.com/anacrolix/log`. There is no event bus. Debug lines are gated by `BTsets.EnableDebug`. A log file at 100MB is deleted and recreated, not rotated.

## Tests

`make test` from the repo root is the race gate:

```bash
cd server && go test -race -count=1 -timeout=10m -coverprofile=coverage.out -covermode=atomic $(go list ./... | grep -v '/docs$')
```

That is `TEST_PKGS` in the Makefile. It excludes generated Swagger only. `rutor` is included. GStreamer CGo is not: those files are `//go:build gst`, and default tests compile `server/gstreamer/bridge/bridge_stub.go`. Do not pass `-tags gst` in the default gate. Do not add a test that needs a real kernel FUSE mount, a real `ffprobe`, or the network.

Coverage (`/.testcoverage.yml`): total 70% on `server/coverage.out`. File and package floors are 0. Do not fail a change for a single package under 70%, and do not lower the total. `make bench` is not the whole module. It is `./subs ./torr/storage/torrstor ./torrshash ./torrfs ./web/api` at `-benchtime=200ms`. The `torr` reader bench blocks on `GotInfo` and is excluded on purpose. CI fails a benchmark that regresses past 200%.

`make lint` is `golangci-lint run ./...` in `server/`. Config is v2, `default: standard`, `max-issues-per-linter: 0`, `max-same-issues: 0`. Format is `gofmt -s` and `goimports` (`make fmt`, `make fmt-check`). CI runs both format checks before the linter. `make lint-all` is the local equivalent.

Web CI, when `web/src` changes: `yarn typecheck`, `yarn lint`, `yarn test`, `yarn i18n:check`, `yarn test:e2e`, `yarn build`.

Leak checks: `leakcheck.AssertNoGoroutineLeak(t, baseline)` polls for 1s. `leakcheck.Heap` does two GCs and `debug.FreeOSMemory` before `ReadMemStats`. Snapshot goroutines before the work, not after `Prepare`, if `Prepare` starts the tracker prefetch.

Live tests stay opted in: `LIVE_TORRENT=1` plus `TORRSERVER_URL` (`server/web/api/live_torrent_test.go`), `LIVE_INDEXER=1` (`server/torznab/jacred_live_test.go`).

Packages with no `*_test.go` today: `gstreamer/bridge` (stub), `leakcheck` (helpers), `mobile/torrserverkit`, `rutor/models`, `rutor/torrsearch`, `rutor/utils`, `torr/offlinetest` (harness), `torr/storage`, `torr/storage/state`. A change in those packages borrows a test from the caller (`rutor`, `torrstor`, `web/api`, `tgbot`).

### Offline torrent harness

`server/torr/offlinetest` is the harness. `Prepare` points `settings` at `t.TempDir()`, calls `InitSets`, turns DHT/PEX/uTP/LPD/UPnP/IPv6/disk off, and restores every global it touched in `t.Cleanup`, including `CloseDB`. Call it once per test. Do not open `config.db` yourself.

```go
func TestStreamFromLoopback(t *testing.T) {
    offlinetest.Prepare(t)
    show := offlinetest.WriteShow(t) // 16KiB pieces, nested files, real hashes
    seed := offlinetest.Seed(t, show) // 127.0.0.1:0, NoDHT, no trackers, no PEX, no IPv6
    _ = seed.Torrent
    // Add show.Spec through torr.NewTorrent / the HTTP handler under test.
}
```

`WriteShow` builds `Season 1/episode1.mkv`, a second episode, an `.srt`, and a JPEG from temp files, then `metainfo.Info.BuildFromFilePath` and `bencode`. `Seed` waits until `BytesCompleted() == Length()` with a 15s deadline. The seeder's `memStore.Piece` is complete and serves `ReadAt` from the concatenated bytes. The fork has no file storage client, so do not "fix" `Seed` to use a directory.

`Reader` tests that cannot construct `torrent.File` (unexported length) set `overrideLen` and leave `file` nil. `fileLen` / `fileOffset` already branch on that. Do not export anacrolix fields.

### ffprobe

`server/ffprobe/exec_test.go` writes a shell script into `t.TempDir()`, then `goffprobe.SetFFProbeBinPath` and assigns `binFile`. Cleanup restores both. Windows skips. `probeTimeout` is a package var; tests shrink it. A missing binary is `Exists() == false`, not a failed test.

### Telegram

`tele.NewBot(tele.Settings{URL: srv.URL, Token: "123:test", Offline: true, Client: srv.Client()})` against `httptest.NewServer`. See `server/tgbot/bot_api_test.go` and `handlers_direct_test.go`. Do not call `api.telegram.org`.

### HTTP handlers

`server/web/api/offline_api_test.go` is the pattern: `offlinetest.Prepare`, a Gin engine, no listen on `8090`. Race coverage is `handlers_race_test.go`. Streaming disconnect coverage is `stream_mem_test.go` (`TestStreamDisconnectReleasesReaders`).

### DHT and trackers

Replace `lookupDHTHost` and `dhtBootstrapHosts` in `btserver`. Replace `defaultTrackersListURLs` in `torr/utils`. Restore them with `t.Cleanup`. A test that fetches the ngosang list is a live test and does not belong in `go test` by default.

## Quality gates before a `server/` change is done

From the repo root:

```bash
make fmt-check
make lint
make test
```

Run `make bench` when the diff touches `server/torr`, `server/torr/storage/torrstor`, or `server/web/api`. A hot-path change needs a bench that fails if the copy buffer or the piece slab stops being pooled (`-benchmem` allocs/op). `piecebuf_test.go`, `stream_buffer_test.go`, and `cache_mem_test.go` are the existing ceilings.

`gofmt` failures are not lint waivers. `//nolint` is not a fix for a lock or a leak.

## Anti-patterns

- `os.Exit` or `log.Fatal` in `settings`, `torr`, `torrstor`, `web`, `dlna`, `bonjour`, `mcp`.
- Holding `bt.mu` as `RLock` and then `Lock` on the same goroutine.
- Calling `settings.ListTorrent` or `settings.AddTorrent` while `settings.mu` is held.
- `defer bt.mu.RUnlock()` across a branch that needs `bt.mu.Lock`.
- Reading `t.Torrent`, `t.cache`, or `bt.client` without a snapshot under the mutex that owns them.
- `Cache.Close` while `Storage.mu` is held.
- `clearPriority` not under `muPrio`, or a sleep "so the new reader can win".
- Evicting a full piece whose `Complete` flag is still false.
- `Put` of a piece slab after `isClosed`, or `Put` of a sub-slice.
- A fresh `[]byte` per `Read` on `/stream` or `/play`.
- `http.DefaultClient` or `http.Get` for tracker lists.
- One tracker tier per URL.
- `dht.GlobalBootstrapAddrs`.
- Upstream anacrolix APIs marked `NE`.
- Editing `server/docs` or `server/web/pages/template/pages` by hand.
- A unit test that listens on the public DHT, talks to Telegram, or requires `ffprobe` on `PATH`.
- Changing `TS_EN_SSL` or `TS_SSL_ENABLE` on only one of Compose, the entrypoint, and the README.
