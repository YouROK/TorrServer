#!/bin/sh

FLAGS="--path $TS_CONF_PATH --logpath $TS_LOG_PATH --port $TS_PORT --torrentsdir $TS_TORR_DIR"
if [[ -n "$TS_HTTPAUTH" ]]; then FLAGS="${FLAGS} --httpauth"; fi
if [[ -n "$TS_RDB" ]]; then FLAGS="${FLAGS} --rdb"; fi
if [[ -n "$TS_DONTKILL" ]]; then FLAGS="${FLAGS} --dontkill"; fi

if [ ! -d $TS_CONF_PATH ]; then
  mkdir -p $TS_CONF_PATH
fi

if [ ! -d $TS_TORR_DIR ]; then
  mkdir -p $TS_TORR_DIR
fi

if [ ! -f $TS_LOG_PATH ]; then
  touch $TS_LOG_PATH
fi

echo "Running with: ${FLAGS}"

torrserver $FLAGS
