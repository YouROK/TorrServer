# Release image: binaries are built by GoReleaser (see .goreleaser.yaml dockers_v2).
# For a full in-Docker build (e.g. master branch CI), use Dockerfile.standalone.

### UPX COMPRESSING START ###
FROM ubuntu AS compressed

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/torrserver ./torrserver

RUN apt-get update && apt-get install -y --no-install-recommends upx-ucl \
    && upx --best --lzma ./torrserver \
    && rm -rf /var/lib/apt/lists/*
### UPX COMPRESSING END ###


### BUILD MAIN IMAGE START ###
FROM alpine

ENV TS_CONF_PATH="/opt/ts/config"
ENV TS_LOG_PATH="/opt/ts/log"
ENV TS_TORR_DIR="/opt/ts/torrents"
ENV TS_PORT=8090
ENV GODEBUG=madvdontneed=1

COPY --from=compressed ./torrserver /usr/bin/torrserver
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN apk add --no-cache --update ffmpeg

CMD /docker-entrypoint.sh
### BUILD MAIN IMAGE end ###
