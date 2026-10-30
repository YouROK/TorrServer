# Development

See **[BUILD.md](BUILD.md)** for Makefile commands, CI, tagged releases, GHCR images, and fork setup.

## Go server

Run from source:

```bash
make run
# or
cd server && go run ./cmd
```

## Web development

React 19 + Vite 8 + HeroUI v3 UI. See **[web/README.md](../web/README.md)** for stack, scripts, and env setup.

```bash
cd web && yarn && yarn start
# ship into binary:
cd web && yarn build && cd .. && make webgen-clean
```

## Build

Use the **Makefile** (wraps [GoReleaser](https://goreleaser.com/)):

- **Local dev:** `.goreleaser.local.yaml` — linux/darwin amd64+arm64 (`make binary`, `make build`, `make dist`)
- **CI / release:** `.goreleaser.yaml` — full matrix (`make dist-full`, `make release`)

```bash
make help              # command overview
make install-tools     # goreleaser v2 + swag
make start-build       # build host binary, install to data/, run
make binary            # build host platform → dist/TorrServer-<os>-<arch>
make binary-gst         # GStreamer build for host platform
make dist              # local snapshot (4 platforms, no publish)
make dist-full         # full snapshot (all platforms)
make update            # web embed + swagger
```

**Tagged releases** (`.goreleaser.yaml`, workflow `.github/workflows/release.yml`) use the `MatriX.*` tag scheme and publish to `ghcr.io/<owner>/<repo>` where `<owner>/<repo>` is the GitHub repository lowercased. Override locally with `REGISTRY_IMAGE=owner/repo make …`.

Direct GoReleaser (without Make):

```bash
goreleaser build --snapshot --clean --single-target --id binary
GOOS=linux GOARCH=amd64 goreleaser build --snapshot --clean --single-target --id binary
```

Web UI build inside `gen_web.go` needs **Node.js 22+** (see `web/.nvmrc`).

Go toolchain versions are pinned in [`.github/versions.env`](../.github/versions.env) (`GO_VERSION` for the server, `GO_ANDROID_VERSION` for Android; `server/go.mod` carries the module minimum). See [BUILD.md](BUILD.md) for the toolchain wrappers, cross-compilation, Docker builder mode (`USE_DOCKER_BUILDER=1`), and Android NDK setup.

### Server

- Install [Golang](https://golang.org/doc/install) (`GO_VERSION` in `.github/versions.env`; the minimum is the `go` line in `server/go.mod`, and `GO_ANDROID_VERSION` is additionally needed for Android via GoReleaser)
- Go to the TorrServer source directory
- Run `make binary` or GoReleaser as above

### Web

- Install **Node.js 22+** and **yarn**
- Go to the web directory, run `yarn build`
- From the repo root, embed with `make webgen-clean` (or `go run gen_web.go --clean`)

### Android

To build an Android server you will need the Android Toolchain (release workflow / full `.goreleaser.yaml`).

### iOS

iOS cannot run TorrServer as a subprocess. Releases include an in-process XCFramework built with `gomobile bind`:

- Asset: `TorrServer-ios-TorrServerKit.xcframework.zip`
- Minimum iOS: **18.0**
- Slices: `ios-arm64` (device), `ios-arm64-simulator`, `ios-amd64-simulator`

Build on macOS with Xcode:

```bash
./build-ios.sh
```

Link `TorrServerKit.xcframework` into the host app and call the gomobile API:

```swift
import TorrServerKit

let err = TorrserverkitStartServer(8090, dataDir)
if !err.isEmpty { /* handle error */ }
// HTTP API: http://127.0.0.1:8090  (/echo, /stream, /torrents, …)
TorrserverkitStopServer()
```

Host `Info.plist`:

```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
    <key>NSAllowsLocalNetworking</key>
    <true/>
</dict>
<key>NSLocalNetworkUsageDescription</key>
<string>TorrServer discovers local peers and streams media over local HTTP.</string>
```

Limits: FUSE and GStreamer are not available on iOS. Background playback is the host app’s responsibility (`UIBackgroundModes`). The XCFramework is statically linked, so the combined iOS app is a GPL-3.0 derivative work. Public App Store submission of BitTorrent clients may be rejected under Guideline 5.2.3.

### Swagger

```bash
make install-swag
make update-swag
# or manually:
cd server && swag init -g web/server.go
swag fmt
```

Standard binaries serve a filtered Swagger spec at runtime (only `/gst/settings`); `-gst` builds include all `/gst/*` endpoints.
