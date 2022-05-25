### FRONT BUILD START ###
FROM --platform=$BUILDPLATFORM node:16-alpine as front
COPY ./web /app
WORKDIR /app
# Build front once upon multiarch build
RUN yarn install && yarn run build
### FRONT BUILD END ###

### BUILD TORRSERVER MULTIARCH START ###
FROM --platform=$BUILDPLATFORM golang:1.18-alpine as builder

COPY . /opt/src
COPY --from=front /app/build /opt/src/web/build

WORKDIR /opt/src

ARG TARGETARCH

# Step for multiarch build with docker buildx
ENV GOARCH=$TARGETARCH

# Build torrserver
RUN apk add --update g++ \
&& go run gen_web.go \
&& cd server \
&& go clean -i -r -cache \
&& go mod tidy \
&& go build -ldflags '-w -s' --o "torrserver" ./cmd 
### BUILD TORRSERVER MULTIARCH END ###


### UPX COMPRESSING START ###
FROM debian:buster-slim as compressed

COPY --from=builder /opt/src/server/torrserver ./torrserver

RUN apt-get update && apt-get install -y upx-ucl && upx --best --lzma ./torrserver
# Compress torrserver only for amd64 and arm64 no variant platforms
# ARG TARGETARCH
# ARG TARGETVARIANT
# RUN if [ "$TARGETARCH" == 'amd64' ]; then compress=1; elif [ "$TARGETARCH" == 'arm64' ] && [ -z "$TARGETVARIANT"  ]; then compress=1; else compress=0; fi \
# && if [[ "$compress" -eq 1 ]]; then ./upx --best --lzma ./torrserver; fi
### UPX COMPRESSING END ###

### BUILD MAIN IMAGE START ###
FROM alpine

ENV TS_CONF_PATH="/opt/ts/config"
ENV TS_LOG_PATH="/opt/ts/log"
ENV TS_TORR_DIR="/opt/ts/torrents"
ENV TS_PORT=8090
ENV GODEBUG=madvdontneed=1

COPY --from=compressed ./torrserver /usr/bin/torrserver
COPY ./docker-entrypoint.sh /docker-entrypoint.sh

RUN apk add --no-cache --update ffmpeg

CMD /docker-entrypoint.sh
### BUILD MAIN IMAGE end ###
