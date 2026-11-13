### FRONT BUILD START ###
FROM --platform=$BUILDPLATFORM node:16-alpine AS front
WORKDIR /app
ARG REACT_APP_SERVER_HOST='.'
ARG REACT_APP_TMDB_API_KEY=''
ARG PUBLIC_URL=''
ENV REACT_APP_SERVER_HOST=$REACT_APP_SERVER_HOST
ENV REACT_APP_TMDB_API_KEY=$REACT_APP_TMDB_API_KEY
ENV PUBLIC_URL=$PUBLIC_URL
COPY ./web/package.json ./web/yarn.lock ./
RUN yarn install
COPY ./web .
RUN yarn run build
### FRONT BUILD END ###
### BUILD TORRSERVER MULTIARCH START ###
FROM --platform=$BUILDPLATFORM golang:1.27.1-bookworm AS builder
COPY . /opt/src
COPY --from=front /app/build /opt/src/web/build
WORKDIR /opt/src
ARG TARGETARCH
ARG VERSION=MatriX.Docker-GST
ENV GOARCH=$TARGETARCH
RUN apt-get update && apt-get install -y --no-install-recommends g++ \
&& go run gen_web.go \
&& cd server \
&& go mod tidy \
&& go clean -i -r -cache \
&& go build -tags "nosqlite gst" -ldflags "-s -w -checklinkname=0 -X server/version.Version=${VERSION}" --o "torrserver" ./cmd
### BUILD TORRSERVER MULTIARCH END ###
### UPX COMPRESSING START ###
FROM ubuntu AS compressed
COPY --from=builder /opt/src/server/torrserver ./torrserver
RUN apt update && apt install -y upx-ucl && upx --best --lzma ./torrserver
### UPX COMPRESSING END ###
### BUILD MAIN IMAGE START ###
FROM debian:bookworm-slim
ENV TS_CONF_PATH="/opt/ts/config"
ENV TS_LOG_PATH="/opt/ts/log"
ENV TS_TORR_DIR="/opt/ts/torrents"
ENV TS_PORT=8090
ENV GODEBUG=madvdontneed=1
COPY --from=compressed ./torrserver /usr/bin/torrserver
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-base-apps \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    gstreamer1.0-plugins-ugly \
    gstreamer1.0-libav \
    gstreamer1.0-tools \
    && rm -rf /var/lib/apt/lists/* /usr/share/man /usr/share/locale /usr/share/info
COPY --chmod=755 ./docker-entrypoint.sh /docker-entrypoint.sh
CMD ["/docker-entrypoint.sh"]
### BUILD MAIN IMAGE end ###
