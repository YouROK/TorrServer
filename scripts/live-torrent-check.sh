#!/usr/bin/env bash
# Live torrent / stream / ffp / optional Torznab check using repo data/ (gitignored).
# Usage (from repo root):
#   ./scripts/live-torrent-check.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${LIVE_PORT:-18091}"
DATADIR="${DATA_DIR:-$ROOT/data}"
HASH="${LIVE_TORRENT_HASH:-e5a5bdb8ff6152657a1e051024b618dd37d76957}"
export TORRSERVER_URL="${TORRSERVER_URL:-http://127.0.0.1:${PORT}}"
export TS_USER="${TS_USER:-ts}"
export TS_PASS="${TS_PASS:-ts}"
export LIVE_TORRENT_HASH="$HASH"
export LIVE=1
export LIVE_TORRENT=1
export LIVE_INDEXER=1

if [[ ! -f "$DATADIR/accs.db" ]]; then
  echo "missing $DATADIR/accs.db" >&2
  exit 1
fi

if lsof -nP -iTCP:"$PORT" -sTCP:LISTEN >/dev/null 2>&1; then
  echo "port $PORT already in use — reusing TORRSERVER_URL=$TORRSERVER_URL"
else
  echo "starting TorrServer on $PORT with -d $DATADIR"
  (
    cd "$ROOT/server"
    CGO_ENABLED=0 go run -tags nosqlite ./cmd -p "$PORT" -d "$DATADIR" -a -i 127.0.0.1
  ) > /tmp/ts-live-check.log 2>&1 &
  SERVER_PID=$!
  trap 'kill "$SERVER_PID" 2>/dev/null || true' EXIT
  for _ in $(seq 1 60); do
    if curl -sf "$TORRSERVER_URL/echo" >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done
  curl -sf "$TORRSERVER_URL/echo" >/dev/null
fi

echo "=== Go live torrent ==="
(
  cd "$ROOT/server"
  go test ./web/api -run TestLiveTorrentStreamStatPlayFFP -count=1 -timeout 3m
)

echo "=== Go JacRed (jacred.stream) indexer searches ==="
(
  cd "$ROOT/server"
  go test ./torznab -count=1 -timeout 4m
)

echo "=== Playwright live tests ==="
(
  cd "$ROOT/web"
  yarn test:e2e:live
)

echo "live-torrent-check ok"
