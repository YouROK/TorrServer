#!/bin/sh

FLAGS="--path $TS_CONF_PATH --logpath $TS_LOG_PATH --port $TS_PORT --torrentsdir $TS_TORR_DIR"
if [[ -n "$TS_HTTPAUTH" ]]; then FLAGS="${FLAGS} --httpauth"; fi
if [[ -n "$TS_RDB" ]]; then FLAGS="${FLAGS} --rdb"; fi
if [[ -n "$TS_DONTKILL" ]]; then FLAGS="${FLAGS} --dontkill"; fi

echo "Running with: ${FLAGS}"

torrserver $FLAGS
