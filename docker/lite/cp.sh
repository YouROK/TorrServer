#!/bin/bash

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
mkdir -p /app

cp /src/dist/$binName /app/TorrServer