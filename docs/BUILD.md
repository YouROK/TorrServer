# Build, release, and Docker

This document describes how to build TorrServer locally, run CI-equivalent checks, publish releases, and publish container images. The setup is **fork-friendly**: registry names are derived from the GitHub repository when possible, not hardcoded to a single maintainer.

## Quick start

```bash
make help                 # grouped command list
make version              # host platform, registry image, paths
make start-build          # build host binary → data/ → run
```

Install tools once:

```bash
make install-tools        # goreleaser v2 + swag
```

## Configuration

| Variable | Where set | Purpose |
|----------|-----------|---------|
| `REGISTRY_IMAGE` | env or auto from `git remote origin` | GHCR image slug (`owner/repo`, lowercase) |
| `TORRSERVER_GITHUB_REPO` | env | GitHub releases for install scripts (default `YouROK/TorrServer`) |
| `GORELEASER_CONFIG` | env / Makefile | `.goreleaser.local.yaml` (local) or `.goreleaser.yaml` (CI/release) |
| `TARGET` | CLI | Single GoReleaser platform, e.g. `linux_amd64` |
| `SKIP` | CLI | GoReleaser skip list, e.g. `before`, `docker` |
| `DATA_DIR` | env / Makefile | Local runtime dir (default `data/`) |

### Registry image naming

Container images are published as:

```text
ghcr.io/<REGISTRY_IMAGE>:<tag>
```

**GitHub Actions** (`.github/workflows/release.yml`) sets `REGISTRY_IMAGE` from `${{ github.repository }}` lowercased — e.g. `YouROK/TorrServer` → `yourok/torrserver`.

**Local `make`** uses the same rule: lowercase `owner/repo` from `git remote get-url origin`. Override on the command line:

```bash
REGISTRY_IMAGE=myorg/torrserver make release-snapshot
```

GoReleaser configs (`.goreleaser.yaml`, `.goreleaser.local.yaml`) use:

```yaml
ghcr.io/{{ envOrDefault "REGISTRY_IMAGE" "yourok/torrserver" }}
```

The fallback `yourok/torrserver` matches the upstream project; forks and CI override via `REGISTRY_IMAGE`.

## Makefile workflows

### Build binaries → `dist/`

| Command | Description |
|---------|-------------|
| `make build` | All platforms in `.goreleaser.local.yaml` + flatten |
| `make build-host` | Current machine only |
| `make build-one TARGET=linux_amd64` | Single platform + flatten |
| `make build-no-hooks` | Skip pre-build hooks (`SKIP=before`) |
| `make flatten` | `dist/torrserver_*` → `TorrServer-os-arch` flat names |

### Local run → `data/`

Runtime files (binary, `config.db`, `settings.json`, `torrents/`) live under `data/` (gitignored).

| Command | Description |
|---------|-------------|
| `make data-sync` | Copy host binary from `dist/` to `data/` |
| `make build-sync` | `build-host` + `data-sync` |
| `make start` | Run `data/TorrServer-…` |
| `make start-build` | Build, sync, run |
| `make run` | `go run` from source (no `data/` binary) |

### Docker (local)

| Command | Description |
|---------|-------------|
| `make docker-image` | Build `torrserver:local` for host Linux arch |
| `make docker-start` | Run local image with `data/` mounted at `/opt/ts` |
| `make docker-start-release` | Run `ghcr.io/$(REGISTRY_IMAGE):latest-<arch>` + `data/` |
| `make release-snapshot` | Binaries + docker images locally, no publish |

Platform-specific local images:

```bash
make docker-image-amd64
make docker-image-arm64
```

### Web UI & API docs

| Command | Description |
|---------|-------------|
| `make update` | Rebuild web embed + swagger |
| `make update-clean` | Clean web rebuild + swagger |
| `make web-deps` | `yarn install` in `web/` |

### Release (publish)

| Command | Description |
|---------|-------------|
| `make release-snapshot` | Snapshot: artifacts + docker, `--skip=publish` |
| `make release-no-docker` | Snapshot without docker |
| `make release` | Tagged release + push (needs clean tree, tag, `GITHUB_TOKEN`) |
| `make docker-push` | Push `ghcr.io/$(REGISTRY_IMAGE):latest-<arch>` |

## GoReleaser configs

| File | Use |
|------|-----|
| `.goreleaser.yaml` | **CI and tagged releases** — full platform matrix, Android, docker |
| `.goreleaser.local.yaml` | **Local dev** — reduced targets, same docker/registry rules |

Tagged releases use the `MatriX.*` version scheme (not strict semver). Releases run with `--skip=validate` so GoReleaser accepts these tags.

### Toolchains

- Main server build: **Go 1.26.4** (`go1.26.4` via GoReleaser)
- Android build: **Go 1.25.7** + Android NDK (release workflow only)

Install wrappers for local builds:

```bash
go install golang.org/dl/go1.26.4@latest && go1.26.4 download
go install golang.org/dl/go1.25.7@latest && go1.25.7 download   # Android only
```

Web UI build needs Node 16–18, or Node 17+ with OpenSSL legacy (`NODE_OPTIONS=--openssl-legacy-provider`, set automatically in Makefile and GoReleaser).

## CI (`.github/workflows/ci.yml`)

Runs on push/PR to `master`:

1. **GoReleaser check** — validates `.goreleaser.yaml`
2. **Snapshot build** — single target `linux_amd64`, verifies binary exists

No secrets required beyond default `GITHUB_TOKEN` (unused for build-only job).

## Release workflow (`.github/workflows/release.yml`)

Triggers on **any tag push**. Steps:

1. Set `REGISTRY_IMAGE` from `${{ github.repository }}` (lowercase)
2. Install Go toolchains + Android NDK path for Android builds
3. QEMU + Docker Buildx for multi-arch images
4. Login to `ghcr.io` with `GITHUB_TOKEN`
5. `goreleaser release --clean --skip=validate`

**Forks**: enable Actions and push a tag; images publish to `ghcr.io/<your-org>/<repo>`. No workflow edits needed.

**Secrets**:

- `GITHUB_TOKEN` — automatic (packages + releases)
- `TMDB_API_KEY` — optional; web build embeds TMDB in CI if set

## Docker usage (published images)

Replace `<owner>/<repo>` with your registry slug (lowercase GitHub `owner/repo`):

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

GoReleaser publishes platform tags such as `latest-amd64` and `latest-arm64` in addition to release tags.

Upstream example:

```bash
docker run --rm -d --name torrserver -p 8090:8090 ghcr.io/yourok/torrserver:latest
```

## Fork checklist

1. Fork on GitHub; clone with `origin` pointing at your fork.
2. `make install-tools && make build-host && make start-build` — verify local build.
3. Enable GitHub Actions on the fork.
4. Push a tag (e.g. `MatriX.141.4-test`) to test release + GHCR push.
5. Pull container: `docker pull ghcr.io/<your-org>/<repo>:<tag>`.

Install scripts (`installTorrServerLinux.sh`, `installTorrServerMac.sh`) default to **YouROK/TorrServer** (correct for upstream end users). Forks that publish their own releases can point at another repo without editing the scripts:

```bash
TORRSERVER_GITHUB_REPO=myorg/TorrServer sudo bash ./installTorrServerLinux.sh --install --silent
```

## Clean

| Command | Removes |
|---------|---------|
| `make clean` | `dist/` |
| `make clean-web` | `web/build/` |
| `make clean-all` | dist, web/build, go cache, docker staging (git-safe) |

Does **not** delete committed embed/swagger or `server/docs/`.
