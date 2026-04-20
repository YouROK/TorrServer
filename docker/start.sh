#!/bin/sh

case $(uname -m) in
    i386) architecture="386" ;;
    i686) architecture="386" ;;
    x86_64) architecture="amd64" ;;
    aarch64) architecture="arm64" ;;
    armv7|armv7l) architecture="arm7" ;;
    armv6|armv6l) architecture="arm5" ;;
#    armv5|armv5l) architecture="arm5" ;;
    *) echo "Unsupported Arch. Can't continue."; exit 1 ;;
esac

binName="TorrServer-linux-${architecture}"

mkdir -p /opt/torrserver
cd /opt/torrserver

rm -f ${binName}*

wget -O $binName "https://github.com/9000000/TorrServer/releases/latest/download/$binName"
chmod +x $binName

FLAGS=""

#sets start flags
[ ! -z "$TS_PORT" ] && echo "TS_PORT: $TS_PORT" && FLAGS="${FLAGS} --port ${TS_PORT}"

# Support both TS_CONF_PATH (full) and TS_PATH (lite)
CONF_PATH="${TS_CONF_PATH:-$TS_PATH}"
if [ ! -z "$CONF_PATH" ]; then
    echo "CONF_PATH: $CONF_PATH"
    FLAGS="${FLAGS} --path ${CONF_PATH}"
    [ ! -d "$CONF_PATH" ] && mkdir -p "$CONF_PATH"
fi

# Support both TS_LOG_PATH (full) and TS_LOGFILE (lite)
if [ ! -z "$TS_LOG_PATH" ]; then
    echo "TS_LOG_PATH: $TS_LOG_PATH"
    FLAGS="${FLAGS} --logpath ${TS_LOG_PATH}"
    LOG_DIR=$(dirname "$TS_LOG_PATH")
    [ ! -d "$LOG_DIR" ] && mkdir -p "$LOG_DIR"
elif [ ! -z "$TS_LOGFILE" ]; then
    echo "TS_LOGFILE: $TS_LOGPATHDIR/$TS_LOGFILE"
    FLAGS="${FLAGS} --logpath $TS_LOGPATHDIR/${TS_LOGFILE}"
    [ ! -d "$TS_LOGPATHDIR" ] && mkdir -p "$TS_LOGPATHDIR"
fi

[ ! -z "$TS_WEBLOGFILE" ] && echo "TS_WEBLOGFILE: $TS_LOGPATHDIR/$TS_WEBLOGFILE" && FLAGS="${FLAGS} --weblogpath $TS_LOGPATHDIR/${TS_WEBLOGFILE}"
[ ! -z "$TS_RDB" ] | [ "$TS_RDB" = "true" ] && echo "TS_RDB: $TS_RDB" && FLAGS="${FLAGS} --rdb"
[ ! -z "$TS_HTTPAUTH" ] && echo "TS_HTTPAUTH: $TS_HTTPAUTH" && FLAGS="${FLAGS} --httpauth"
[ ! -z "$TS_DONTKILL" ] && echo "TS_DONTKILL: $TS_DONTKILL" && FLAGS="${FLAGS} --dontkill"

# Support both TS_TORR_DIR (full) and TS_TORRENTSDIR (lite)
TORR_DIR="${TS_TORR_DIR:-$TS_TORRENTSDIR}"
if [ ! -z "$TORR_DIR" ]; then
    echo "TORR_DIR: $TORR_DIR"
    FLAGS="${FLAGS} --torrentsdir ${TORR_DIR}"
    [ ! -d "$TORR_DIR" ] && mkdir -p "$TORR_DIR"
fi

[ ! -z "$TS_TORRENTADDR" ] && echo "TS_TORRENTADDR: $TS_TORRENTADDR" && FLAGS="${FLAGS} --torrentaddr ${TS_TORRENTADDR}"
[ ! -z "$TS_PUBIPV4" ] && echo "TS_PUBIPV4: $TS_PUBIPV4" && FLAGS="${FLAGS} --pubipv4 ${TS_PUBIPV4}"
[ ! -z "$TS_PUBIPV6" ] && echo "TS_PUBIPV6: $TS_PUBIPV6" && FLAGS="${FLAGS} --pubipv6 ${TS_PUBIPV6}"
[ ! -z "$TS_SEARCHWA" ]&& echo "TS_SEARCHWA: $TS_SEARCHWA" && FLAGS="${FLAGS} --searchwa"
[ ! -z "$TS_PROXYURL" ] && echo "TS_PROXYURL: $TS_PROXYURL" && FLAGS="${FLAGS} --proxyurl ${TS_PROXYURL}"
[ ! -z "$TS_PROXYMODE" ] && echo "TS_PROXYMODE: $TS_PROXYMODE" && FLAGS="${FLAGS} --proxymode ${TS_PROXYMODE}"

echo "Running with: ${FLAGS}"
export GODEBUG=madvdontneed=1

/opt/torrserver/${binName} ${FLAGS}
