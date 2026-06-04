# TorrServer — local build (.goreleaser.local.yaml)
# Usage: make help

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

export NODE_OPTIONS ?= --openssl-legacy-provider

# Toolchain versions for .goreleaser.yaml (see .github/versions.env)
-include .github/versions.env
export GO_VERSION GO_ANDROID_VERSION

# ── GoReleaser ──────────────────────────────────────────────────────────────
GORELEASER        ?= goreleaser
GORELEASER_CONFIG ?= .goreleaser.local.yaml
GORELEASER_ARGS   ?=
BUILD_ID          ?= torrserver
TARGET            ?=
OUTPUT            ?=
SKIP              ?=
DIST_DIR          ?= dist
DATA_DIR          ?= data

GR_COMMON := --config $(GORELEASER_CONFIG) $(GORELEASER_ARGS)
GR_SKIP   := $(if $(SKIP),--skip=$(SKIP),)

# ── Host platform (native binary in dist/ and data/) ─────────────────────────
HOST_UNAME_M  := $(shell uname -m)
HOST_GOOS     := $(shell cd server && go env GOOS 2>/dev/null)
HOST_GOARCH   := $(shell cd server && go env GOARCH 2>/dev/null)
HOST_TARGET   := $(HOST_GOOS)_$(HOST_GOARCH)

ifeq ($(HOST_GOOS)-$(HOST_GOARCH),darwin-amd64)
HOST_BIN := TorrServer-darwin-amd64
else ifeq ($(HOST_GOOS)-$(HOST_GOARCH),darwin-arm64)
HOST_BIN := TorrServer-darwin-arm64
else ifeq ($(HOST_GOOS)-$(HOST_GOARCH),linux-amd64)
HOST_BIN := TorrServer-linux-amd64
else ifeq ($(HOST_GOOS)-$(HOST_GOARCH),linux-arm64)
HOST_BIN := TorrServer-linux-arm64
else
HOST_BIN := torrserver
endif

# ── Docker / GHCR (auto from git origin; same rule as .github/workflows/release.yml) ──
# ghcr.io/<owner>/<repo> with lowercase owner/repo (GitHub Container Registry convention)
GIT_ORIGIN_SLUG := $(shell git remote get-url origin 2>/dev/null | tr '[:upper:]' '[:lower:]' | sed 's/.*github.com[:/]//; s/.git$$//')
REGISTRY_IMAGE    ?= $(if $(GIT_ORIGIN_SLUG),$(GIT_ORIGIN_SLUG),yourok/torrserver)
DOCKER_IMAGE      ?= torrserver
DOCKER_TAG        ?= local
DOCKER_CONTEXT    ?= $(DIST_DIR)/docker
IMAGE_RELEASE     ?= ghcr.io/$(REGISTRY_IMAGE)

ifeq ($(HOST_UNAME_M),arm64)
HOST_DOCKER_PLATFORM := linux/arm64
else ifeq ($(HOST_UNAME_M),aarch64)
HOST_DOCKER_PLATFORM := linux/arm64
else
HOST_DOCKER_PLATFORM := linux/amd64
endif
DOCKER_PLATFORM   ?= $(or $(DOCKER_DEFAULT_PLATFORM),$(HOST_DOCKER_PLATFORM))
DOCKER_TAG_RELEASE ?= latest-$(lastword $(subst /, ,$(DOCKER_PLATFORM)))
DOCKER_LINUX_TARGET = $(subst /,_,$(DOCKER_PLATFORM))
DOCKER_LINUX_BIN    = TorrServer-linux-$(lastword $(subst /, ,$(DOCKER_PLATFORM)))

DOCKER_DATA_ENV = -e TS_CONF_PATH=/opt/ts \
                  -e TS_LOG_PATH=/opt/ts/torrserver.log \
                  -e TS_TORR_DIR=/opt/ts/torrents

GR_DOCKER_ENV         := REGISTRY_IMAGE=$(REGISTRY_IMAGE)
GR_RELEASE_SNAPSHOT   := $(GR_DOCKER_ENV) $(GORELEASER) release --snapshot --clean --skip=publish

.DEFAULT_GOAL := help

# ===========================================================================
# Help
# ===========================================================================

.PHONY: help help-all

help:
	@echo "TorrServer — $(GORELEASER_CONFIG)"
	@echo "Host: $(HOST_TARGET) → $(HOST_BIN)   Docker: $(DOCKER_PLATFORM)"
	@echo "Registry: $(REGISTRY_IMAGE) (override: REGISTRY_IMAGE=owner/repo make …)"
	@echo "Docs: docs/BUILD.md   All targets: make help-all"
	@echo ""
	@echo "Workflows:"
	@echo "  make start-build          build host binary → data/ → run"
	@echo "  make build && make data-sync"
	@echo "  make run                  go run (source, no data/ binary)"
	@echo "  make release-snapshot     binaries + docker images, no publish"
	@echo "  make docker-image && make docker-start"
	@echo "  make update               web embed + swagger"
	@echo ""
	@echo "Build → $(DIST_DIR)/"
	@printf "  \033[36m%-22s\033[0m %s\n" build "All platforms (+ flatten)"
	@printf "  \033[36m%-22s\033[0m %s\n" build-host "Host only ($(HOST_TARGET))"
	@printf "  \033[36m%-22s\033[0m %s\n" build-one "One platform (TARGET=linux_amd64)"
	@printf "  \033[36m%-22s\033[0m %s\n" build-no-hooks "Skip pre-build hooks (SKIP=before)"
	@printf "  \033[36m%-22s\033[0m %s\n" flatten "Flat release names in dist/"
	@printf "  \033[36m%-22s\033[0m %s\n" dist "ls dist/"
	@echo ""
	@echo "Release"
	@printf "  \033[36m%-22s\033[0m %s\n" release-snapshot "Snapshot, no publish"
	@printf "  \033[36m%-22s\033[0m %s\n" release-no-docker "Snapshot, SKIP=docker"
	@printf "  \033[36m%-22s\033[0m %s\n" release "Tagged release + push"
	@echo ""
	@echo "Local run → $(DATA_DIR)/"
	@printf "  \033[36m%-22s\033[0m %s\n" data-sync "Copy host binary dist/ → data/"
	@printf "  \033[36m%-22s\033[0m %s\n" build-sync "build-host + data-sync"
	@printf "  \033[36m%-22s\033[0m %s\n" start "Run $(DATA_DIR)/$(HOST_BIN)"
	@printf "  \033[36m%-22s\033[0m %s\n" start-build "build-sync + start"
	@printf "  \033[36m%-22s\033[0m %s\n" run "go run ./cmd (source)"
	@echo ""
	@echo "Docker"
	@printf "  \033[36m%-22s\033[0m %s\n" docker-image "Build $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@printf "  \033[36m%-22s\033[0m %s\n" docker-start "Run local image + $(DATA_DIR)/"
	@printf "  \033[36m%-22s\033[0m %s\n" docker-start-release "Run $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE) + $(DATA_DIR)/"
	@printf "  \033[36m%-22s\033[0m %s\n" docker-push "Push release image"
	@echo ""
	@echo "Web & tools"
	@printf "  \033[36m%-22s\033[0m %s\n" update "Web embed + swagger"
	@printf "  \033[36m%-22s\033[0m %s\n" update-clean "Clean web rebuild + swagger"
	@printf "  \033[36m%-22s\033[0m %s\n" check "Validate goreleaser config"
	@printf "  \033[36m%-22s\033[0m %s\n" version "Tool versions"
	@printf "  \033[36m%-22s\033[0m %s\n" clean-all "Clean dist, web/build, caches"

help-all: ## Alphabetical target list
	@grep -hE '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | sort \
		| awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ===========================================================================
# Tools
# ===========================================================================

.PHONY: check healthcheck show-config version install-tools install-goreleaser install-swag deps

check: ## Validate GoReleaser config
	$(GORELEASER) check $(GR_COMMON)

healthcheck: ## Check GoReleaser dependencies
	$(GORELEASER) healthcheck $(GR_COMMON)

show-config: ## Print goreleaser config file
	@cat $(GORELEASER_CONFIG)

version: ## Tool and platform info
	@echo "host:     $(HOST_TARGET) → $(HOST_BIN)"
	@echo "data:     $(CURDIR)/$(DATA_DIR)/"
	@echo "registry: $(REGISTRY_IMAGE) → $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)"
	@echo "docker:   $(DOCKER_PLATFORM) (local tag $(DOCKER_IMAGE):$(DOCKER_TAG))"
	@echo "config:   $(GORELEASER_CONFIG)"
	@echo "go:       $$(go version)"
	@echo "goreleaser: $$($(GORELEASER) --version 2>/dev/null | head -1 || echo not installed)"

install-tools: install-goreleaser install-swag ## Install goreleaser + swag
install-goreleaser: ## Install GoReleaser v2
	go install github.com/goreleaser/goreleaser/v2@latest
install-swag: ## Install swag
	go install github.com/swaggo/swag/cmd/swag@latest

deps: ## go mod download
	cd server && go mod download

# ===========================================================================
# Build → dist/
# ===========================================================================

.PHONY: build build-host build-one build-no-hooks flatten

build: ## All platforms + flatten
	$(GORELEASER) build --snapshot --clean $(GR_COMMON) $(GR_SKIP)
	$(MAKE) flatten

build-host: ## Host platform only ($(HOST_TARGET))
	$(MAKE) build-one TARGET=$(HOST_TARGET)

build-one: ## Single platform (TARGET=linux_amd64)
ifndef TARGET
	$(error TARGET required, e.g. make build-one TARGET=linux_amd64)
endif
	TARGET="$(TARGET)" $(GORELEASER) build --snapshot --clean --single-target --id $(BUILD_ID) \
		$(GR_COMMON) $(GR_SKIP) $(if $(OUTPUT),-o $(OUTPUT),)
	@if [ -z "$(OUTPUT)" ]; then $(MAKE) flatten; fi

build-no-hooks: ## build with SKIP=before (skip pre-build hooks)
	$(MAKE) build SKIP=before

flatten: ## dist/torrserver_* → TorrServer-os-arch flat names
	@test -d $(DIST_DIR) || { echo "$(DIST_DIR)/ missing — run make build first" >&2; exit 1; }
	@shopt -s nullglob; \
	_name_from_dir() { \
	  case "$$1" in \
	    torrserver_linux_amd64_v1) echo TorrServer-linux-amd64 ;; \
	    torrserver_linux_arm64_v8.0) echo TorrServer-linux-arm64 ;; \
	    torrserver_darwin_amd64_v1) echo TorrServer-darwin-amd64 ;; \
	    torrserver_darwin_arm64_v8.0) echo TorrServer-darwin-arm64 ;; \
	    torrserver_linux_arm_5) echo TorrServer-linux-arm5 ;; \
	    torrserver_linux_arm_7) echo TorrServer-linux-arm7 ;; \
	    torrserver_windows_amd64_v1) echo TorrServer-windows-amd64.exe ;; \
	    torrserver_windows_386_sse2) echo TorrServer-windows-386.exe ;; \
	    torrserver_freebsd_amd64_v1) echo TorrServer-freebsd-amd64 ;; \
	    torrserver_freebsd_arm_7) echo TorrServer-freebsd-arm7 ;; \
	    torrserver_android_arm_7) echo TorrServer-android-arm7 ;; \
	    torrserver_android_arm64_v8.0) echo TorrServer-android-arm64 ;; \
	    torrserver_android_386_sse2) echo TorrServer-android-386 ;; \
	    torrserver_android_amd64_v1) echo TorrServer-android-amd64 ;; \
	    *) return 1 ;; \
	  esac; \
	}; \
	n=0; \
	for dir in $(DIST_DIR)/torrserver_*/; do \
	  bin="$$dir/torrserver"; [ -f "$$bin" ] || continue; \
	  name=$$(_name_from_dir "$$(basename "$${dir%/}")") || continue; \
	  cp -f "$$bin" "$(DIST_DIR)/$$name"; echo "  $(DIST_DIR)/$$name"; n=$$((n+1)); \
	done; \
	test "$$n" -gt 0 || { echo "No torrserver_* in $(DIST_DIR)/" >&2; exit 1; }

dist: ## ls dist/
	@ls -la $(DIST_DIR)/ 2>/dev/null || echo "$(DIST_DIR)/ empty"

checksums: ## cat dist/checksums.txt
	@test -f $(DIST_DIR)/checksums.txt && cat $(DIST_DIR)/checksums.txt \
		|| echo "No checksums — run make release-snapshot"

# ===========================================================================
# Release (binaries + archives + docker)
# ===========================================================================

.PHONY: release release-snapshot release-no-docker

release-snapshot: ## Snapshot, no publish
	$(GR_RELEASE_SNAPSHOT) $(GR_COMMON) $(GORELEASER_ARGS) $(GR_SKIP)

release: ## Tagged release + push (on tag commit, GITHUB_TOKEN)
	$(GR_DOCKER_ENV) $(GORELEASER) release --clean --skip=validate \
		$(GR_COMMON) $(GORELEASER_ARGS) $(GR_SKIP)

release-no-docker: ## Snapshot without docker
	$(MAKE) release-snapshot SKIP=docker

# ===========================================================================
# Local run → data/
# ===========================================================================

.PHONY: data-sync build-sync start start-build run

data-sync: ## Copy host binary from dist/ to $(DATA_DIR)/
	@mkdir -p $(DATA_DIR)/torrents
	@_target="$(HOST_TARGET)"; _flat="$(DIST_DIR)/$(HOST_BIN)"; \
	if [ -f "$$_flat" ]; then _src="$$_flat"; \
	else _src=$$(find "$(DIST_DIR)" -path "*$$_target*" -name torrserver -type f 2>/dev/null | head -1); fi; \
	if [ -z "$${_src:-}" ] || [ ! -f "$$_src" ]; then \
	  echo "No binary for $(HOST_TARGET) — run: make build or make build-host" >&2; exit 1; fi; \
	cp -f "$$_src" "$(DATA_DIR)/$(HOST_BIN)"; chmod +x "$(DATA_DIR)/$(HOST_BIN)"; \
	echo "synced $(DATA_DIR)/$(HOST_BIN) <- $$_src"

build-sync: build-host data-sync ## Build host binary + sync to data/

start: ## Run binary from $(DATA_DIR)/
	@bin="$(CURDIR)/$(DATA_DIR)/$(HOST_BIN)"; \
	test -x "$$bin" || { echo "Missing $$bin — run: make data-sync or make build-sync" >&2; exit 1; }; \
	cd "$(DATA_DIR)" && exec "./$(HOST_BIN)"

start-build: build-sync start ## Build, sync, run (main local workflow)

run: ## Run from Go source (not data/ binary)
	cd server && CGO_ENABLED=0 go run -tags nosqlite ./cmd

# ===========================================================================
# Docker
# ===========================================================================

.PHONY: docker-image docker-image-amd64 docker-image-arm64
.PHONY: docker-start docker-start-release docker-push docker-clean
.PHONY: _docker-stage _docker-build-linux

docker-image: _docker-build-linux _docker-stage ## Build $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "docker-image: $(DOCKER_PLATFORM) → $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker buildx build --platform $(DOCKER_PLATFORM) --load \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) -f Dockerfile $(DOCKER_CONTEXT)

docker-image-amd64: ## Build local image for linux/amd64
	$(MAKE) docker-image DOCKER_PLATFORM=linux/amd64

docker-image-arm64: ## Build local image for linux/arm64
	$(MAKE) docker-image DOCKER_PLATFORM=linux/arm64

_docker-stage:
	@_target="$(DOCKER_LINUX_TARGET)"; _flat="$(DIST_DIR)/$(DOCKER_LINUX_BIN)"; \
	if [ -f "$$_flat" ]; then _src="$$_flat"; \
	else _src=$$(find "$(DIST_DIR)" -path "*$$_target*" -name torrserver -type f 2>/dev/null | head -1); fi; \
	test -n "$${_src:-}" || { echo "No binary for $(DOCKER_LINUX_TARGET) — run make build-one TARGET=$(DOCKER_LINUX_TARGET)" >&2; exit 1; }; \
	rm -rf $(DOCKER_CONTEXT); mkdir -p "$(DOCKER_CONTEXT)/$(DOCKER_PLATFORM)"; \
	cp -f "$$_src" "$(DOCKER_CONTEXT)/$(DOCKER_PLATFORM)/torrserver"; \
	cp -f docker-entrypoint.sh "$(DOCKER_CONTEXT)/"; echo "staged $$_src"

_docker-build-linux:
	$(MAKE) build-one TARGET=$(DOCKER_LINUX_TARGET)

docker-start: ## Run $(DOCKER_IMAGE):$(DOCKER_TAG) with $(DATA_DIR)/
	@mkdir -p $(DATA_DIR)/torrents
	docker run --rm -it --name torrserver $(DOCKER_DATA_ENV) \
		-v $(CURDIR)/$(DATA_DIR):/opt/ts -p 8090:8090 $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-start-release: ## Run release image with $(DATA_DIR)/
	@mkdir -p $(DATA_DIR)/torrents
	docker run --rm -it --name torrserver $(DOCKER_DATA_ENV) \
		-v $(CURDIR)/$(DATA_DIR):/opt/ts -p 8090:8090 $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)

docker-push: ## Push $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)
	docker push $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)

docker-clean: ## Remove $(DOCKER_CONTEXT) staging dir
	rm -rf $(DOCKER_CONTEXT)

# ===========================================================================
# Web UI & API
# ===========================================================================

.PHONY: web-deps web-build web-embed update-web update-web-clean update-swag update update-clean

web-deps: ## yarn install
	cd web && yarn

web-build: ## yarn build
	cd web && yarn run build

web-embed: ## gen_web.go embed
	go run gen_web.go

update-web: web-build web-embed ## Rebuild web + embed
update-web-clean: ## Clean web/build + rebuild embed
	go run gen_web.go --clean

update-swag: ## Regenerate swagger
	@command -v swag >/dev/null 2>&1 || $(MAKE) install-swag
	cd server && swag init -g web/server.go

update: update-web update-swag ## Web embed + swagger
update-clean: update-web-clean update-swag ## Clean web rebuild + swagger

# ===========================================================================
# Clean
# ===========================================================================

.PHONY: clean clean-web clean-cache clean-all

clean: ## Remove dist/
	rm -rf $(DIST_DIR)

clean-web: ## Remove web/build/
	rm -rf web/build

clean-cache: ## go clean -cache
	cd server && go clean -cache -testcache

clean-all: clean clean-web clean-cache docker-clean ## dist + web/build + caches (git-safe)
