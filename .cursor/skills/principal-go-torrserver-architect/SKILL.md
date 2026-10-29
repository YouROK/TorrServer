---
name: principal-go-torrserver-architect
description: >-
  Principal Go systems and BitTorrent streaming architect for the TorrServer
  `server/` module (anacrolix/torrent fork, torrstor piece cache, HTTP Range
  /stream and /play). Use for all Go work in this repo: concurrency and race
  debugging, cache and streaming memory, settings and HTTP handlers, and tests.
  Read AGENTS.md before editing. There is no internal/ engine.
---

# TorrServer Go architect

Read [AGENTS.md](../../../AGENTS.md) before editing `server/`. That file is the map of locks, forks, traps, and commands. This skill is the working rule. Do not restate Go basics. Ship thread-safe, bounded, cancellable changes with no compatibility regressions.

Module path is `server`. Run Go from `server/`. `go 1.25.7` in `server/go.mod`. No root `go.mod`, no `internal/` tree.

Hot path: `GET/HEAD /stream` and `/play` → `torr.Torrent.Stream` → `torrstor.Reader` → piece cache → `anacrolix/torrent`. DLNA, Bonjour, FUSE, WebDAV, Telegram, MCP, and GStreamer stay off that path.

## Forks

Import upstream paths. `server/go.mod` replaces them:

- `github.com/anacrolix/torrent` → `github.com/lampac-nextgen/torrent`
- `github.com/anacrolix/utp` → `github.com/lampac-nextgen/utp`
- `github.com/anacrolix/upnp` → `github.com/lampac-nextgen/upnp`

`NE` comments are APIs this fork does not have (directory file storage, `DisableWebtorrent`, `HeaderObfuscationPolicy`, storage `Capacity`). Do not call them. `OE` comments are the live behavior.

Peer id prefix stays `-qB4390-` / `qBittorrent/4.3.9`. `UpnpID` stays `TorrServer/<version>`. `TotalHalfOpenConns` is 500.

## Do not touch unless the task names it

- `server/docs` (generated Swagger) and `server/web/pages/template/pages` (generated UI). Edit `web/src`, then rebuild.
- `server/gstreamer` CGo. Files are `//go:build gst`. Default tests compile `bridge/bridge_stub.go`. Do not pass `-tags gst` and do not grep `gstreamer` out of `make test`.
- New JS runtimes, plugin engines, or an FFmpeg pipeline. Remux already exists behind `gst`.
- `rutor` is in the default test set. Do not exclude it.

## Locks

`sync.RWMutex` is not reentrant. `RLock` then `Lock` on `BTServer.mu` deadlocks. `server/torr/clientrace_test.go` fails if `NewTorrent` holds the read lock across the write lock.

Order when both are needed: `settings.dbMigrationLock` then `settings.mu`; `BTServer.mu`; `Torrent.muTorrent`; `torrstor.Storage.mu` released before `Cache.Close`; cache `muRemove` → `muReaders` → `muActive`. `Reader.resume` drops `Reader.mu` around the anacrolix `Seek`. `applyZones` drops `muReaders` before `SetZones`.

Snapshot `t.Torrent`, `t.cache`, and `bt.client` under the mutex that owns them. `drop` nils the first two at any time. `Cache.pieces` is read-only after `Init`. `MemPiece.ReadAt` uses `Lock` because it writes `Accessed`.

Do not hold those mutexes across tracker HTTP, DHT, `publicip.Get4/Get6`, or disk flush. Full order and the `NewJsonDB` `:=` shadow are in AGENTS.md.

`FuseAutoMount` still calls `os.Exit(1)`. New library code returns an error or uses the existing `exitFunc` (`cmd`, `dlna`). `settings.Embedded` must not exit the host process.

## Streaming and trackers

- Copy buffer is `streamCopyPool`: `*[64KiB]byte`, not a bare `[]byte`. Read-ahead is `newBufferedStreamReader(..., 1<<20)`. Nil the pooled `ResponseWriter` before `Put`. `buf[:nr]` must not escape.
- Piece slabs are `*pieceSlab` pooled by exact length. `putPieceBuf` clears the full cap. `Release` does not recycle when `cache.isClosed`. Never keep a sub-slice after `Release`.
- Do not evict a piece inside a merged reader range, the begin/end keep window (at least 8MiB), or a full piece whose `Complete` flag is still false.
- Download windows are zone `(capacity/readers)*ReaderReadAHead/100`, readahead `min(16MiB/readers, zone)`, next `min(4MiB/readers, readahead)`. Push them on reader create, close, and the 1s `updateRA` tick. Not on every small `Read`. Do not call `Piece.SetPriority` from the cache: it raises over those tiers. Evicted pieces call `Piece.Forget` (fork v1.2.42), not `UpdateCompletion`, so the client re-requests their chunks.
- Idle park is `minReaderIdleTimeout` (60s), or `TorrentDisconnectTimeout` if larger. `readerOff` must not `Seek` while `inFlight > 0`.
- `Reader.Close` does not unregister. `Cache.CloseReader` does, and it is safe on a nil `*Cache`.
- A handler that accepted the request must write a status. An empty return becomes `200` with an empty body.
- Tracker client is `trackersHTTPClient`: 5s timeout, 3s dial, 3s TLS, body cap 256KiB, its own `Transport`. Do not use `http.DefaultClient`. A failed refresh keeps the previous list and must not block startup.
- `SanitizeTrackers` does not cap the list. `addExtraTrackers` appends one tier. One tier per URL is an announce storm on this fork.
- DHT bootstrap uses `lookupDHTHost` and a 3s timeout. Do not switch it to `dht.GlobalBootstrapAddrs`.

## Tests

Use `server/torr/offlinetest`. The fork has no directory file storage.

```go
offlinetest.Prepare(t)       // temp DB, DHT/PEX/uTP/IPv6 off, restores globals
show := offlinetest.WriteShow(t)
seed := offlinetest.Seed(t, show) // 127.0.0.1:0 memStore, already complete
```

`ffprobe`: shell script in `t.TempDir()`, assign `binFile`, `goffprobe.SetFFProbeBinPath`, restore in `t.Cleanup`. Windows skips. Telegram: `httptest` plus `tele.Settings{URL: srv.URL, Offline: true, Client: srv.Client()}`. Leaks: `leakcheck.AssertNoGoroutineLeak` and `leakcheck.Heap`. Replace `lookupDHTHost`, `dhtBootstrapHosts`, and `defaultTrackersListURLs` in tests. `LIVE_TORRENT=1` and `LIVE_INDEXER=1` stay opt-in.

## Done when

From the repo root:

```bash
make fmt-check
make lint
make test
```

`make test` is `-race -count=1 -timeout=10m -covermode=atomic` on every package except `server/docs`. Coverage floor is **70% total** (`.testcoverage.yml`). File and package floors are 0.

`make bench` when the diff touches `server/torr`, `server/torr/storage/torrstor`, or `server/web/api`. It benches only `./subs ./torr/storage/torrstor ./torrshash ./torrfs ./web/api`. The `torr` reader bench blocks on `GotInfo` and is excluded on purpose. A hot-path change needs a `-benchmem` check that the copy buffer or piece slab is still pooled.
