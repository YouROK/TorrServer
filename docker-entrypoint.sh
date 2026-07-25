#!/bin/sh

FLAGS="--path $TS_CONF_PATH --logpath $TS_LOG_PATH --port $TS_PORT --torrentsdir $TS_TORR_DIR"
if [ -n "$TS_IP" ]; then FLAGS="${FLAGS} --ip ${TS_IP}"; fi
if [ "$TS_HTTPAUTH" = "1" ]; then FLAGS="${FLAGS} --httpauth"; fi
if [ "$TS_RDB" = "1" ]; then FLAGS="${FLAGS} --rdb"; fi
if [ "$TS_DONTKILL" = "1" ]; then FLAGS="${FLAGS} --dontkill"; fi
if [ "$TS_SSL_ENABLE" = "1" ]; then FLAGS="${FLAGS} --ssl"; fi
if [ "$TS_SEARCH_WA_ENABLE" = "1" ]; then FLAGS="${FLAGS} --searchwa"; fi
if [ "$TS_WEBDAV_ENABLE" = "1" ]; then FLAGS="${FLAGS} --webdav"; fi
if [ "$TS_FORCE_HTTPS_ENABLE" = "1" ]; then FLAGS="${FLAGS} --force-https"; fi
if [ -n "$TS_SSL_PORT" ]; then FLAGS="${FLAGS} --sslport ${TS_SSL_PORT}"; fi
if [ -n "$TS_SSL_CERT_PATH" ]; then FLAGS="${FLAGS} --sslcert ${TS_SSL_CERT_PATH}"; fi
if [ -n "$TS_SSL_KEY_PATH" ]; then FLAGS="${FLAGS} --sslkey ${TS_SSL_KEY_PATH}"; fi
if [ -n "$TS_PROXYURL" ]; then FLAGS="${FLAGS} --proxyurl ${TS_PROXYURL}"; fi
if [ -n "$TS_PROXYMODE" ]; then FLAGS="${FLAGS} --proxymode ${TS_PROXYMODE}"; fi
if [ -n "$TS_WEB_LOG_PATH" ]; then FLAGS="${FLAGS} --weblogpath ${TS_WEB_LOG_PATH}"; fi
if [ -n "$TS_TORR_ADDR" ]; then FLAGS="${FLAGS} --torrentaddr ${TS_TORR_ADDR}"; fi
if [ -n "$TS_PUBLIC_IPV4_ADDR" ]; then FLAGS="${FLAGS} --pubipv4 ${TS_PUBLIC_IPV4_ADDR}"; fi
if [ -n "$TS_PUBLIC_IPV6_ADDR" ]; then FLAGS="${FLAGS} --pubipv6 ${TS_PUBLIC_IPV6_ADDR}"; fi
if [ -n "$TS_MAX_SIZE" ]; then FLAGS="${FLAGS} --maxsize ${TS_MAX_SIZE}"; fi
if [ -n "$TS_TELEGRAM_TOKEN" ]; then FLAGS="${FLAGS} --tg ${TS_TELEGRAM_TOKEN}"; fi
if [ -n "$TS_FUSE_PATH" ]; then FLAGS="${FLAGS} --fuse ${TS_FUSE_PATH}"; fi

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
