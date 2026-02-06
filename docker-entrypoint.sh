#!/bin/sh

FLAGS="--path $TS_CONF_PATH --logpath $TS_LOG_PATH --port $TS_PORT --torrentsdir $TS_TORR_DIR"
if [ -n "$TS_IP" ]; then FLAGS="${FLAGS} -i ${TS_IP}"; fi
if [ "$TS_HTTPAUTH" = "1" ]; then FLAGS="${FLAGS} --httpauth"; fi
if [ "$TS_RDB" = "1" ]; then FLAGS="${FLAGS} --rdb"; fi
if [ "$TS_DONTKILL" = "1" ]; then FLAGS="${FLAGS} --dontkill"; fi
if [ "$TS_EN_SSL" = "1" ]; then FLAGS="${FLAGS} --ssl"; fi
if [ -n "$TS_SSL_PORT" ]; then FLAGS="${FLAGS} --sslport ${TS_SSL_PORT}"; fi
if [ -n "$TS_PROXYURL" ]; then FLAGS="${FLAGS} --proxyurl ${TS_PROXYURL}"; fi
if [ -n "$TS_PROXYMODE" ]; then FLAGS="${FLAGS} --proxymode ${TS_PROXYMODE}"; fi

if [ ! -d "$TS_CONF_PATH" ]; then
  mkdir -p "$TS_CONF_PATH"
fi

if [ ! -d "$TS_TORR_DIR" ]; then
  mkdir -p "$TS_TORR_DIR"
fi

if [ ! -f "$TS_LOG_PATH" ]; then
  touch "$TS_LOG_PATH"
fi

echo "Running with: ${FLAGS}"

exec torrserver $FLAGS
