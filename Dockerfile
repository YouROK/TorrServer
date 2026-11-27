### FRONT BUILD START ###
FROM --platform=$BUILDPLATFORM node:20-alpine AS front

WORKDIR /app

ARG REACT_APP_SERVER_HOST='.'
ARG REACT_APP_TMDB_API_KEY=''
ARG PUBLIC_URL=''

ENV REACT_APP_SERVER_HOST=$REACT_APP_SERVER_HOST
ENV REACT_APP_TMDB_API_KEY=$REACT_APP_TMDB_API_KEY
ENV PUBLIC_URL=$PUBLIC_URL
ENV NODE_OPTIONS=--openssl-legacy-provider

COPY ./web/package.json ./web/yarn.lock ./
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn \
    yarn install --frozen-lockfile --network-timeout 600000

# Build front once upon multiarch build
COPY ./web .
RUN yarn run build
### FRONT BUILD END ###


### BUILD TORRSERVER MULTIARCH START ###
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS builder

WORKDIR /opt/src/server

# Modules get their own layer (a plain layer, so the gha cache keeps it too):
# it is reused until go.mod/go.sum change
COPY ./server/go.mod ./server/go.sum ./
RUN go mod download

COPY . /opt/src
COPY --from=front /app/build /opt/src/web/build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION=MatriX.Docker

# Same flags as the release binaries in ts_build.yml. The compiler always runs
# natively on the build platform; only the output is for the target platform.
RUN --mount=type=cache,target=/root/.cache/go-build \
    cd /opt/src && go run gen_web.go \
    && cd server \
    && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
       go build -trimpath -tags=nosqlite \
       -ldflags "-s -w -checklinkname=0 -X server/version.Version=${VERSION}" \
       -o /opt/src/server/torrserver ./cmd
### BUILD TORRSERVER MULTIARCH END ###


### BUILD MAIN IMAGE START ###
FROM alpine

ENV TS_CONF_PATH="/opt/ts/config"
ENV TS_LOG_PATH="/opt/ts/log"
ENV TS_TORR_DIR="/opt/ts/torrents"
ENV TS_PORT=8090
ENV GODEBUG=madvdontneed=1

# Rarely changing layers first: an image update then only ships the new binary
RUN apk add --no-cache --update ffmpeg

COPY ./docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=builder /opt/src/server/torrserver /usr/bin/torrserver

CMD /docker-entrypoint.sh
### BUILD MAIN IMAGE end ###
