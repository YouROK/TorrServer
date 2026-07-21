# Release image: binaries are built and compressed by GoReleaser (see .goreleaser.yaml).
FROM alpine:latest
# in case if dockerhub rate limits are exceeded:
#FROM ghcr.io/linuxcontainers/alpine:latest

ENV TS_CONF_PATH="/opt/ts/config"
ENV TS_LOG_PATH="/opt/ts/log"
ENV TS_TORR_DIR="/opt/ts/torrents"
ENV TS_PORT=8090
ENV GODEBUG=madvdontneed=1

ARG TARGETARCH
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/TorrServer-linux-$TARGETARCH* /usr/bin/torrserver
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN apk add --no-cache --update ffmpeg

CMD /docker-entrypoint.sh
