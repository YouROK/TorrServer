# Building TorrServer

## Prerequisites

By default, builds run **goreleaser on the host**. Install:

| Tool | Purpose | Install |
|------|---------|---------|
| [goreleaser](https://goreleaser.com/getting-started/install/oss/) | Build & release automation | `go install github.com/goreleaser/goreleaser/v2@latest` |
| [Go](https://go.dev/dl/) | Server build (`go$(GO_VERSION)` from `.github/versions.env`) | See link |
| [yarn](https://yarnpkg.com/getting-started/install) | Web asset bundling (`make webgen`) | See link |
| [Node.js 22+](https://nodejs.org/) | Vite web build (see `web/.nvmrc`) | See link |
| [swag](https://github.com/swaggo/swag) | Swagger docs (before hooks) | `go install github.com/swaggo/swag/cmd/swag@latest` |

| [upx](https://upx.github.io) | Optional binary compression | `brew install upx` / `apt install upx` |

### UPX compression (optional)

| Target | UPX by default |
|--------|----------------|
| `make binary`, `make binary-gst` | off |
| `make release`, `make dist`, `make dist-full` | off |
| `make docker` | **on** |
| `make USE_DOCKER_BUILDER=1 …` | **on** (upx in builder image) |

Override anytime:

```sh
make ENABLE_UPX=1 binary          # compress local binary
make ENABLE_UPX=1 release         # compress release artifacts
make SKIP_UPX=1 docker            # docker image without UPX
```

CI (`release.yml`, `ci.yml`) sets `UPX_ENABLED=false` — same as local release defaults.

---

## Docker builder mode (`USE_DOCKER_BUILDER=1`)

Optional: run goreleaser inside a Docker image with Go, yarn, swag, upx, and Android NDK preinstalled. Useful when you don't want to install toolchains locally.

```sh
make setup-builder                  # build torrserver.builder image once
make USE_DOCKER_BUILDER=1 binary
make USE_DOCKER_BUILDER=1 dist
```

Only **make** and **Docker** (with `linux/amd64` support) are required on the host in this mode. Builder image tag: **`torrserver.builder`** (lowercase — Docker requirement).

> **Why amd64 only?** The Android NDK only ships prebuilt toolchains for `x86_64` Linux. The builder image runs as `linux/amd64`, so the host must be able to execute amd64 containers.

---

## Docker setup

Docker is **optional**. You only need it for:

- `make docker` / `make docker-start` — run TorrServer in a container
- `make USE_DOCKER_BUILDER=1 …` — goreleaser inside `Dockerfile.builder`
- `make release` / `make dist` — multi-arch release images (needs buildx)

For everyday development, `make binary` uses the host toolchain and does **not** require Docker.

### Linux

Install [Docker Engine](https://docs.docker.com/engine/install/) or your distro packages. For multi-arch release builds:

```sh
docker run --privileged --rm tonistiigi/binfmt --install all
```

### macOS

#### Docker Desktop

1. Install [Docker Desktop for Mac](https://docs.docker.com/desktop/setup/install/mac-install/).
2. On Apple Silicon: enable **Use Rosetta for x86_64/amd64 emulation** in Docker Desktop → Settings → General. This is needed only for `USE_DOCKER_BUILDER=1` (the builder image is `linux/amd64`).
3. Verify:

```sh
docker run --rm alpine uname -m          # native linux/arm64 on Apple Silicon
docker run --rm --platform=linux/amd64 alpine uname -m   # x86_64 when Rosetta is enabled
```

#### macOS Containers

On supported macOS versions you can use Apple's native [Container CLI](https://github.com/apple/container) instead of Docker Desktop for running Linux containers:

```sh
brew install container
container system start
```

Use this for `make docker-start` style workflows when you prefer Apple's runtime. GoReleaser release builds (`make dist` / `make release`) still expect **Docker** with buildx for publishing multi-arch images to GHCR.

> **Note:** `linux/arm/v7` release images need platform emulation. Docker Desktop handles this via built-in binfmt; on Linux you may need the `binfmt` step above. This matters only for full `make release`, not for local `make binary`.

---

## Building

```sh
# Build binary for the host platform (default target)
make

# Build binary for the host platform with GStreamer support
make binary-gst

# Build Android binary (GOARCH and GOARM are required)
make android GOARCH=arm GOARM=7

# Build all local platforms (fast, no publish)
make build

# Local snapshot — archives + checksums from local config
make dist

# Full snapshot — all platforms + Android + docker (needs NDK or USE_DOCKER_BUILDER=1)
make dist-full

# Cut a real release (on a tag)
make release
```

Run `make help` for the full list of targets and options.

### Quick local workflow

```sh
make start-build       # build host binary → data/torrserver → run
make install           # copy dist binary to data/
make docker && make docker-start
```

### Linux binaries: static vs gst vs Docker

| Build id | ELF | Runtime deps | Docker image |
|----------|-----|--------------|--------------|
| `binary` | static (`CGO_ENABLED=0`) | none | **yes** — `make docker` |
| `binary-gst` | static Go binary | `libgstreamer` via dlopen (purego) | **no** — host install only |
| `android` | dynamic (`CGO_ENABLED=1`) | bionic/NDK | **no** |

GoReleaser `dockers_v2` uses **`id=binary` only** (`TorrServer-linux-<arch>`, never `*-gst-*`).

`make binary GOOS=linux` and `make docker` verify static ELF (no `DT_NEEDED`).  
`make binary-gst GOOS=linux` runs `verify-linux-gst` (warns about runtime `.so`).

Inside Alpine, `ldd torrserver` on a **static** binary may print `Not a valid dynamic program` — that is normal on musl.

```sh
make verify-linux-static GOOS=linux GOARCH=arm64
make verify-linux-gst GOOS=linux GOARCH=arm64   # after binary-gst
```

---

## Cross-compiling

The `binary` target respects the standard Go platform variables. Override them to build for a different target:

```sh
# Linux amd64 on any host
make GOOS=linux GOARCH=amd64

# Linux arm64
make GOOS=linux GOARCH=arm64

# Linux ARMv7 (e.g. for Raspberry Pi 2/3 in 32-bit mode)
make GOOS=linux GOARCH=arm GOARM=7

# Windows amd64
make GOOS=windows GOARCH=amd64
```

In `USE_DOCKER_BUILDER=1` mode the platform variables are forwarded into the builder via `-e` flags. In the default local mode your host Go toolchain must support the target.

Android builds require `GOARCH` and `GOARM` to be set explicitly.

```sh
export ANDROID_NDK_LATEST_HOME=$ANDROID_HOME/ndk/27.0.12077973
make android GOARCH=arm64
make android GOARCH=arm GOARM=7
make USE_DOCKER_BUILDER=1 android GOARCH=arm64   # no local NDK needed
```

---

## GoReleaser configs

| File | Use |
|------|-----|
| `.goreleaser.local.yaml` | **Local dev** — `make binary`, `make build`, `make dist` (linux/darwin amd64+arm64) |
| `.goreleaser.yaml` | **CI and release** — full matrix, Android, `make dist-full`, `make release` |

| Target | Config | Notes |
|--------|--------|-------|
| `make binary`, `make build` | local | linux/darwin amd64+arm64 |
| `make dist` | local | lightweight snapshot (old `release-snapshot`) |
| `make dist-full` | release | all platforms + docker + Android |
| `make release` | release | tagged publish to GHCR |

Override: `make build GORELEASER_CONFIG=.goreleaser.yaml` for full matrix via build.

Local config targets (fast):

- `binary`: linux_amd64, linux_arm64, darwin_amd64, darwin_arm64
- `binary-gst`: same four platforms
- Docker images: linux/amd64, linux/arm64 only

Full release config adds Windows, FreeBSD, MIPS, RISC-V, Android, linux/arm/v7 docker, etc.

- Flat binary names in `dist/` (e.g. `TorrServer-linux-amd64`) — no `flatten` step
- UPX compression via goreleaser (optional; off for binary/release/dist, on for docker)
- Multi-arch Docker images published to `ghcr.io/<owner>/<repo>`

### Registry image naming

**GitHub Actions** sets `DOCKER_IMAGE_ID=ghcr.io/${{ github.repository }}` on release.

**Local `make`** derives `REGISTRY_IMAGE` from `git remote get-url origin` (lowercase `owner/repo`):

```bash
REGISTRY_IMAGE=myorg/torrserver make release
```

### Toolchains

Versions are defined in `.github/versions.env`:

- Main server build: **Go 1.26.4**
- Android build: **Go 1.25.7** + Android NDK (release workflow / `Dockerfile.builder`)

**Local `make`** uses `go$(GO_VERSION)` when that wrapper is on `PATH`; otherwise it falls back to system `go`. For exact CI parity:

```bash
go install golang.org/dl/go1.26.4@latest && go1.26.4 download
go install golang.org/dl/go1.25.7@latest && go1.25.7 download   # Android only
```

Override explicitly: `make GO_BINARY=go binary` or `make GO_BINARY=go1.26.4 binary`.

---

## CI (`.github/workflows/ci.yml`)

Runs on push/PR to `master`:

1. **GoReleaser check** — validates `.goreleaser.yaml` and `.goreleaser.local.yaml`
2. **Snapshot build** — single `linux_amd64` target, verifies `dist/TorrServer-linux-amd64`

## Release workflow (`.github/workflows/release.yml`)

Triggers on **any tag push**:

1. Install Go toolchains + Android NDK path
2. Docker Buildx (+ QEMU on Linux CI runners) for multi-arch images
3. Login to `ghcr.io` with `GITHUB_TOKEN`
4. `goreleaser release --clean --skip=validate`

**Forks**: enable Actions and push a tag; images publish to `ghcr.io/<your-org>/<repo>`.

---

## Docker usage (published images)

```bash
docker run --rm -d --name torrserver -p 8090:8090 ghcr.io/<owner>/<repo>:latest
```

With persistent data:

```bash
docker run --rm -d --name torrserver \
  -v ./data:/opt/ts \
  -e TS_CONF_PATH=/opt/ts \
  -e TS_LOG_PATH=/opt/ts/torrserver.log \
  -e TS_TORR_DIR=/opt/ts/torrents \
  -p 8090:8090 \
  ghcr.io/<owner>/<repo>:latest
```

## Fork checklist

1. Fork on GitHub; clone with `origin` pointing at your fork.
2. `make start-build` — verify local build.
3. Enable GitHub Actions on the fork.
4. Push a tag (e.g. `MatriX.141.4-test`) to test release + GHCR push.

Install scripts default to **YouROK/TorrServer** releases. Forks can override:

```bash
TORRSERVER_GITHUB_REPO=myorg/TorrServer sudo bash ./installTorrServerLinux.sh --install --silent
```

## Clean

| Command | Removes |
|---------|---------|
| `make clean` | `dist/`, `web/build/`, docker staging |
| `make clean-cache` | `.cache/` Go module/build caches |

Does **not** delete committed embed/swagger or `server/docs/`.
