# TorrServer — local build (.goreleaser.local.yaml); release uses .goreleaser.yaml
# Usage: make help

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

-include .github/versions.env

GORELEASER_CONFIG         ?= .goreleaser.local.yaml
GORELEASER_RELEASE_CONFIG ?= .goreleaser.yaml
GR_CONFIG         := --config $(GORELEASER_CONFIG)
GR_RELEASE_CONFIG := --config $(GORELEASER_RELEASE_CONFIG)

PROJECT_NAME := TorrServer
RUNTIME_NAME := torrserver

# ── Host platform ─────────────────────────────────────────────────────────────
# Prefer go$(GO_VERSION) when installed (CI / release); fall back to system `go`
GO_BINARY         ?= $(shell command -v go$(GO_VERSION) >/dev/null 2>&1 && echo go$(GO_VERSION) || echo go)
GO_ANDROID_BINARY ?= $(shell command -v go$(GO_ANDROID_VERSION) >/dev/null 2>&1 && echo go$(GO_ANDROID_VERSION) || echo go)
GOOS              ?= $(shell go env GOOS)
GOARCH            ?= $(shell go env GOARCH)
GOARM             ?= $(shell go env GOARM)
NDK_TOOLCHAIN     ?= $(ANDROID_NDK_LATEST_HOME)/toolchains/llvm/prebuilt/linux-x86_64

# Binary name produced by goreleaser:
#   {{ .ProjectName }}-{{ .Os }}-{{ if eq .Arch "arm" }}arm{{ .Arm }}{{ else }}{{ .Arch }}{{ end }}
GOARM_SUFFIX := $(if $(filter arm,$(GOARCH)),$(GOARM),)
DIST_BINARY     := dist/$(PROJECT_NAME)-$(GOOS)-$(GOARCH)$(GOARM_SUFFIX)
DIST_GST_BINARY := dist/$(PROJECT_NAME)-gst-$(GOOS)-$(GOARCH)$(GOARM_SUFFIX)

# ── Install paths ─────────────────────────────────────────────────────────────
DATA_DIR ?= data
DATA_ABS := $(abspath $(DATA_DIR))

# ── Docker / GHCR (auto from git origin; same rule as .github/workflows/release.yml) ──
GIT_ORIGIN_SLUG := $(shell git remote get-url origin 2>/dev/null | tr '[:upper:]' '[:lower:]' | sed 's/.*github.com[:/]//; s/.git$$//')
REGISTRY_IMAGE    ?= $(if $(GIT_ORIGIN_SLUG),$(GIT_ORIGIN_SLUG),yourok/torrserver)
DOCKER_IMAGE_ID   ?= ghcr.io/$(REGISTRY_IMAGE)
IMAGE_BUILDER     := $(RUNTIME_NAME).builder
DOCKER_TAG        ?= local
IMAGE_RELEASE     ?= $(DOCKER_IMAGE_ID)
DIST_DIR          ?= dist

HOST_UNAME_M := $(shell uname -m)
ifeq ($(HOST_UNAME_M),arm64)
DOCKER_PLATFORM      := linux/arm64
DOCKER_ARCH          := arm64
DOCKER_PLATFORM_DIR  := linux_arm64
else ifeq ($(HOST_UNAME_M),aarch64)
DOCKER_PLATFORM      := linux/arm64
DOCKER_ARCH          := arm64
DOCKER_PLATFORM_DIR  := linux_arm64
else
DOCKER_PLATFORM      := linux/amd64
DOCKER_ARCH          := amd64
DOCKER_PLATFORM_DIR  := linux_amd64
endif
DOCKER_TAG_RELEASE   ?= latest-$(DOCKER_ARCH)
DOCKER_CONTEXT_DIR := dist/docker-context
DIST_LINUX_BINARY     := dist/$(PROJECT_NAME)-linux-$(DOCKER_ARCH)
DIST_LINUX_GST_BINARY := dist/$(PROJECT_NAME)-gst-linux-$(DOCKER_ARCH)

DOCKER_DATA_ENV = -e TS_CONF_PATH=/opt/ts \
                  -e TS_LOG_PATH=/opt/ts/torrserver.log \
                  -e TS_TORR_DIR=/opt/ts/torrents

# ── Go build cache ────────────────────────────────────────────────────────────
CACHE_DIR      := $(CURDIR)/.cache
CACHE_GO_MOD   := $(CACHE_DIR)/go-mod
CACHE_GO_BUILD := $(CACHE_DIR)/go-build

# ── Goreleaser flags ──────────────────────────────────────────────────────────
ifeq ($(VERBOSE),1)
GORELEASER_COMMON_FLAGS := --verbose --clean
else
GORELEASER_COMMON_FLAGS := --clean
endif
GORELEASER_BUILD_FLAGS  := --snapshot --single-target --skip=validate

# UPX compression (optional):
#   binary / binary-gst / release / dist — off by default
#   docker / USE_DOCKER_BUILDER=1       — on by default
# Override: ENABLE_UPX=1 (force on) or SKIP_UPX=1 (force off)
ifdef SKIP_UPX
UPX_ENABLED := false
else ifneq ($(filter 1,$(ENABLE_UPX) $(USE_DOCKER_BUILDER)),)
UPX_ENABLED := true
else
UPX_ENABLED := false
endif
SKIP_BEFORE_FLAG := $(if $(SKIP_BEFORE),--skip=before,)
GORELEASER_SKIP_FLAG := $(if $(SKIP),--skip=$(SKIP),)

# Host may use go1.26.4 wrapper; builder image only has plain `go`
ifeq ($(USE_DOCKER_BUILDER),1)
GR_GO_BINARY         := go
GR_GO_ANDROID_BINARY := go
else
GR_GO_BINARY         := $(GO_BINARY)
GR_GO_ANDROID_BINARY := $(GO_ANDROID_BINARY)
endif

GORELEASER_ENVS := \
	GO_BINARY=$(GR_GO_BINARY) \
	GO_ANDROID_BINARY=$(GR_GO_ANDROID_BINARY) \
	REGISTRY_IMAGE=$(REGISTRY_IMAGE) \
	DOCKER_IMAGE_ID=$(DOCKER_IMAGE_ID) \
	UPX_ENABLED=$(UPX_ENABLED)

ifeq ($(USE_DOCKER_BUILDER),1)
GORELEASER_RUN = docker run --platform=linux/amd64 --rm \
	$(foreach v,$(GORELEASER_ENVS),-e $(v)) \
	-v $(CURDIR):/go/src/app \
	-v $(CACHE_GO_MOD):/go/pkg/mod \
	-v $(CACHE_GO_BUILD):/root/.cache/go-build \
	-w /go/src/app
else
GORELEASER_RUN := env $(GORELEASER_ENVS) NDK_TOOLCHAIN=$(NDK_TOOLCHAIN) goreleaser
endif

# ── Guards ────────────────────────────────────────────────────────────────────

guard-docker-amd64:
	@docker run --rm --platform=linux/amd64 alpine uname -m >/dev/null 2>&1 || { \
		echo "\033[31mError: Docker cannot run linux/amd64 containers on this host.\033[0m"; \
		echo "On macOS: enable Rosetta in Docker Desktop → Settings → General."; \
		echo "See docs/BUILD.md"; \
		exit 1; }

guard-goreleaser:
ifeq ($(USE_DOCKER_BUILDER),1)
	@:
else
	@command -v goreleaser >/dev/null 2>&1 || { \
		echo "\033[31mError: 'goreleaser' not found.\033[0m"; \
		echo "Install: go install github.com/goreleaser/goreleaser/v2@latest"; \
		echo "Or use: make USE_DOCKER_BUILDER=1 binary"; \
		exit 1; }
endif

guard-upx:
ifeq ($(UPX_ENABLED),true)
ifneq ($(USE_DOCKER_BUILDER),1)
	@command -v upx >/dev/null 2>&1 || { \
		echo "\033[31mError: 'upx' not found.\033[0m"; \
		echo "Install upx, or run: make SKIP_UPX=1 …"; \
		exit 1; }
endif
endif

guard-ndk:
ifneq ($(USE_DOCKER_BUILDER),1)
	@test -d "$(NDK_TOOLCHAIN)" || { \
		echo "\033[31mError: Android NDK toolchain not found at $(NDK_TOOLCHAIN)\033[0m"; \
		echo "Set ANDROID_NDK_LATEST_HOME or NDK_TOOLCHAIN to the correct path."; \
		echo "Or use: make USE_DOCKER_BUILDER=1 android GOARCH=arm64"; \
		exit 1; }
endif

guard-yarn:
	@command -v yarn >/dev/null 2>&1 || { \
		echo "\033[31mError: 'yarn' not found.\033[0m"; \
		echo "Please read docs/BUILD.md"; \
		exit 1; }

guard-swag:
	@command -v swag >/dev/null 2>&1 || { \
		echo "\033[31mError: 'swag' not found.\033[0m"; \
		echo "Please read docs/BUILD.md"; \
		exit 1; }

# ── Setup: builder image + caches ─────────────────────────────────────────────

setup-builder: guard-docker-amd64 ## Build torrserver.builder goreleaser image
	@docker image inspect $(IMAGE_BUILDER) >/dev/null 2>&1 || \
		docker build --platform=linux/amd64 -t $(IMAGE_BUILDER) -f Dockerfile.builder .

$(CACHE_GO_MOD) $(CACHE_GO_BUILD):
	@mkdir -p $@

setup-cache: $(CACHE_GO_MOD) $(CACHE_GO_BUILD)

prereqs = guard-goreleaser guard-upx
ifneq ($(USE_DOCKER_BUILDER),1)
ifneq ($(SKIP_BEFORE),1)
prereqs += guard-yarn
endif
endif
ifeq ($(USE_DOCKER_BUILDER),1)
prereqs += guard-docker-amd64 setup-builder setup-cache
endif

# NDK required for full release matrix (android) on host only
release_prereqs = $(prereqs)
ifneq ($(USE_DOCKER_BUILDER),1)
release_prereqs += guard-ndk
endif

# Extra args when USE_DOCKER_BUILDER=1 (empty otherwise — must stay on one line)
GORELEASER_DOCKER_BUILD_ARGS := $(if $(filter 1,$(USE_DOCKER_BUILDER)),-e GOOS=$(GOOS) -e GOARCH=$(GOARCH) -e GOARM=$(GOARM) $(IMAGE_BUILDER),)
GORELEASER_DOCKER_ANDROID_ARGS := $(if $(filter 1,$(USE_DOCKER_BUILDER)),-e GOOS=android -e GOARCH=$(GOARCH) -e GOARM=$(GOARM) $(IMAGE_BUILDER),)
GORELEASER_DOCKER_RELEASE_ARGS := $(if $(filter 1,$(USE_DOCKER_BUILDER)),-v /var/run/docker.sock:/var/run/docker.sock $(IMAGE_BUILDER),)

.DEFAULT_GOAL := binary

binary: $(prereqs) ## Build binary for host platform (default)
	@echo "Building binary  GOOS=$(GOOS)  GOARCH=$(GOARCH)  GOARM=$(GOARM)  GO_BINARY=$(GR_GO_BINARY)..."
	@$(GORELEASER_RUN) $(GORELEASER_DOCKER_BUILD_ARGS) build $(GR_CONFIG) $(GORELEASER_COMMON_FLAGS) $(GORELEASER_BUILD_FLAGS) $(SKIP_BEFORE_FLAG) --id=binary
	@[ "$(GOOS)" = linux ] && $(call _verify_static_linux,$(DIST_BINARY)) || true

binary-gst: $(prereqs) ## Build GStreamer binary for host platform
	@echo "Building binary-gst  GOOS=$(GOOS)  GOARCH=$(GOARCH)  GOARM=$(GOARM)..."
	@$(GORELEASER_RUN) $(GORELEASER_DOCKER_BUILD_ARGS) build $(GR_CONFIG) $(GORELEASER_COMMON_FLAGS) $(GORELEASER_BUILD_FLAGS) $(SKIP_BEFORE_FLAG) --id=binary-gst
	@[ "$(GOOS)" = linux ] && $(call _verify_gst_linux,$(DIST_GST_BINARY)) || true

android: $(prereqs) guard-ndk ## Build Android binary (release config)
	@echo "Building Android binary  GOARCH=$(GOARCH)  GOARM=$(GOARM)..."
	@GOOS=android $(GORELEASER_RUN) $(GORELEASER_DOCKER_ANDROID_ARGS) build $(GR_RELEASE_CONFIG) $(GORELEASER_COMMON_FLAGS) $(GORELEASER_BUILD_FLAGS) $(SKIP_BEFORE_FLAG) --id=android

dist: $(prereqs) ## Local snapshot — $(GORELEASER_CONFIG), no publish
	@echo "Local snapshot ($(GORELEASER_CONFIG))..."
	@$(GORELEASER_RUN) release $(GR_CONFIG) $(GORELEASER_COMMON_FLAGS) $(SKIP_BEFORE_FLAG) $(GORELEASER_SKIP_FLAG) --snapshot --skip=publish

dist-full: $(release_prereqs) ## Full snapshot — $(GORELEASER_RELEASE_CONFIG), all platforms
	@echo "Full snapshot ($(GORELEASER_RELEASE_CONFIG))..."
	@$(GORELEASER_RUN) $(GORELEASER_DOCKER_RELEASE_ARGS) release $(GR_RELEASE_CONFIG) $(GORELEASER_COMMON_FLAGS) $(SKIP_BEFORE_FLAG) $(GORELEASER_SKIP_FLAG) --snapshot --skip=publish

release: $(release_prereqs) ## Tagged release via goreleaser (publish)
	@echo "Releasing..."
	@$(GORELEASER_RUN) $(GORELEASER_DOCKER_RELEASE_ARGS) release $(GR_RELEASE_CONFIG) $(GORELEASER_COMMON_FLAGS) $(SKIP_BEFORE_FLAG) $(GORELEASER_SKIP_FLAG)

docker:
	$(MAKE) ENABLE_UPX=1 _docker_image

docker-image: docker
docker-image-amd64:
	$(MAKE) docker DOCKER_PLATFORM=linux/amd64 DOCKER_ARCH=amd64 DOCKER_PLATFORM_DIR=linux_amd64
docker-image-arm64:
	$(MAKE) docker DOCKER_PLATFORM=linux/arm64 DOCKER_ARCH=arm64 DOCKER_PLATFORM_DIR=linux_arm64

# id=binary:     static ELF (CGO_ENABLED=0) — required for Docker/Alpine
# id=binary-gst: static Go ELF, loads libgstreamer*.so at runtime (purego) — not for Docker
define _verify_static_linux
$(SHELL) -eu -c '\
	bin="$(1)"; \
	test -f "$$bin" || { echo "Binary not found: $$bin" >&2; exit 1; }; \
	case "$$(basename "$$bin")" in \
		*-gst-*) echo "gst binary is not for Docker — use id=binary: $$bin" >&2; exit 1;; \
	esac; \
	ftype=$$(file -b "$$bin"); \
	echo "$$ftype" | grep -q ELF || { echo "Not an ELF binary: $$ftype" >&2; exit 1; }; \
	echo "$$ftype" | grep -qi Mach-O && { echo "Mach-O binary — use GOOS=linux for Docker" >&2; exit 1; }; \
	if command -v readelf >/dev/null 2>&1; then \
		if readelf -d "$$bin" 2>/dev/null | grep -q NEEDED; then \
			echo "Expected static Linux binary (no DT_NEEDED):" >&2; \
			readelf -d "$$bin" | grep NEEDED >&2 || true; exit 1; \
		fi; \
	elif ! echo "$$ftype" | grep -q "statically linked"; then \
		echo "Expected statically linked ELF, got: $$ftype" >&2; exit 1; \
	fi; \
	echo "OK: static Linux binary ($$ftype)"'
endef

define _verify_gst_linux
$(SHELL) -eu -c '\
	bin="$(1)"; \
	test -f "$$bin" || { echo "Binary not found: $$bin" >&2; exit 1; }; \
	ftype=$$(file -b "$$bin"); \
	echo "$$ftype" | grep -q ELF || { echo "Not ELF: $$ftype" >&2; exit 1; }; \
	echo "$$ftype" | grep -qi Mach-O && { echo "Mach-O binary — expected Linux gst build" >&2; exit 1; }; \
	echo "OK: gst Linux binary ($$ftype)"; \
	echo "Note: loads libgstreamer/glib via dlopen at runtime — not for Docker/Alpine (use binary)"'
endef

verify-linux-static: ## Verify static Linux ELF (BIN= or GOOS=linux + GOARCH=)
ifeq ($(GOOS),linux)
	@$(call _verify_static_linux,$(if $(BIN),$(BIN),$(DIST_BINARY)))
else
	@$(call _verify_static_linux,$(if $(BIN),$(BIN),dist/$(PROJECT_NAME)-linux-$(GOARCH)$(GOARM_SUFFIX)))
endif

verify-linux-gst: ## Verify gst Linux binary; warns runtime .so deps (BIN= or GOOS=linux)
ifeq ($(GOOS),linux)
	@$(call _verify_gst_linux,$(if $(BIN),$(BIN),$(DIST_GST_BINARY)))
else
	@$(call _verify_gst_linux,$(if $(BIN),$(BIN),dist/$(PROJECT_NAME)-gst-linux-$(GOARCH)$(GOARM_SUFFIX)))
endif

_docker_build_linux:
	$(MAKE) binary GOOS=linux GOARCH=$(DOCKER_ARCH)
	@$(call _verify_static_linux,$(DIST_LINUX_BINARY))

_docker_image: _docker_build_linux
	@echo "Building docker image $(DOCKER_IMAGE_ID):$(DOCKER_TAG) for $(DOCKER_PLATFORM)..."
	@_platform_dir="$${DOCKER_PLATFORM_DIR:-$(DOCKER_PLATFORM_DIR)}"; \
	_platform="$${DOCKER_PLATFORM:-$(DOCKER_PLATFORM)}"; \
	_arch="$${DOCKER_ARCH:-$(DOCKER_ARCH)}"; \
	_ctx="$(DOCKER_CONTEXT_DIR)"; \
	_src="$(DIST_LINUX_BINARY)"; \
	case "$$(basename "$$_src")" in \
		*-gst-*) echo "Refusing gst binary for Docker: $$_src" >&2; exit 1;; \
	esac; \
	test -f "$$_src" || { echo "Binary not found: $$_src — build failed?" >&2; exit 1; }; \
	rm -rf "$$_ctx"; mkdir -p "$$_ctx/$$_platform_dir"; \
	cp -f "$$_src" "$$_ctx/$$_platform_dir/$(PROJECT_NAME)-linux-$$_arch"; \
	cp -f docker-entrypoint.sh "$$_ctx/"; \
	docker buildx build --load \
		--platform "$$_platform" \
		-t $(DOCKER_IMAGE_ID):$(DOCKER_TAG) \
		--build-arg TARGETPLATFORM="$$_platform_dir" \
		--build-arg TARGETARCH="$$_arch" \
		-f Dockerfile "$$_ctx"

docker-start: ## Run local image with data/ mounted
	@mkdir -p $(DATA_DIR)/torrents
	docker run --rm -it --name $(RUNTIME_NAME) $(DOCKER_DATA_ENV) \
		-v $(DATA_ABS):/opt/ts -p 8090:8090 $(DOCKER_IMAGE_ID):$(DOCKER_TAG)

docker-start-release: ## Run GHCR release image with data/ mounted
	@mkdir -p $(DATA_DIR)/torrents
	docker run --rm -it --name $(RUNTIME_NAME) $(DOCKER_DATA_ENV) \
		-v $(DATA_ABS):/opt/ts -p 8090:8090 $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)

docker-push: ## Push local/release image tag
	docker push $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)

docker-clean: ## Remove docker build staging dir
	rm -rf $(DOCKER_CONTEXT_DIR)

install:
	@mkdir -p $(DATA_DIR)/torrents; \
	if [ -f "$(DIST_BINARY)" ]; then _src="$(DIST_BINARY)"; \
	else _src=$$(find "$(DIST_DIR)" -maxdepth 1 -name '$(PROJECT_NAME)-$(GOOS)-*' -type f 2>/dev/null | head -1); fi; \
	test -n "$${_src:-}" && [ -f "$$_src" ] || { \
		echo "\033[31mError: binary not found for $(GOOS)/$(GOARCH)\033[0m"; \
		echo "Run 'make binary' first."; exit 1; }; \
	echo "Installing $$_src → $(DATA_DIR)/$(RUNTIME_NAME)..."; \
	cp "$$_src" "$(DATA_DIR)/$(RUNTIME_NAME)"; chmod +x "$(DATA_DIR)/$(RUNTIME_NAME)"; \
	echo "Done. Data directory: $(DATA_DIR)"

start:
	@bin="$(DATA_DIR)/$(RUNTIME_NAME)"; \
	test -x "$$bin" || { echo "Missing $$bin — run: make install or make start-build" >&2; exit 1; }; \
	cd "$(DATA_DIR)" && exec "./$(RUNTIME_NAME)"

start-build: binary install start

webgen: guard-yarn
	go run gen_web.go $(WEBGEN_EXTRA_FLAGS)

webgen-clean: WEBGEN_EXTRA_FLAGS = --clean
webgen-clean: webgen

swag: guard-swag
	cd server && swag init -g web/server.go

update: webgen swag
update-clean: webgen-clean swag

run:
	cd server && CGO_ENABLED=0 go run -tags nosqlite ./cmd

check: ## Validate local goreleaser config
	goreleaser check $(GR_CONFIG)

check-release: ## Validate release goreleaser config
	goreleaser check $(GR_RELEASE_CONFIG)

# Lint / format (run from repo root; operates on server/)
fmt:
	cd server && gofmt -w .
	cd server && golangci-lint fmt ./...

fmt-check:
	@cd server && \
	unformatted=$$(gofmt -l .) && \
	if [ -n "$$unformatted" ]; then \
		echo "Files need gofmt:" >&2; echo "$$unformatted" >&2; exit 1; \
	fi
	@cd server && \
	diff=$$(golangci-lint fmt -d ./...) && \
	if [ -n "$$diff" ]; then \
		echo "Files need golangci-lint fmt:" >&2; echo "$$diff" >&2; exit 1; \
	fi
	@echo "OK: formatting clean"

lint:
	cd server && golangci-lint run ./...

lint-all: fmt-check lint

healthcheck:
	goreleaser healthcheck $(GR_CONFIG)

show-config:
	@cat $(GORELEASER_CONFIG)

show-config-release:
	@cat $(GORELEASER_RELEASE_CONFIG)

version:
	@echo "host:     $(GOOS)_$(GOARCH)$(GOARM_SUFFIX) → $(DIST_BINARY)"
	@echo "data:     $(DATA_DIR)/$(RUNTIME_NAME)"
	@echo "registry: $(REGISTRY_IMAGE) → $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)"
	@echo "docker:   $(DOCKER_PLATFORM) (local tag $(DOCKER_IMAGE_ID):$(DOCKER_TAG))"
	@echo "go:       $$($(GO_BINARY) version 2>/dev/null || go version)"
	@echo "goreleaser: $$(goreleaser --version 2>/dev/null | head -1 || echo not installed)"
	@echo "config:   local=$(GORELEASER_CONFIG)  release=$(GORELEASER_RELEASE_CONFIG)"

deps:
	cd server && go mod download

# All platforms (no --single-target), same as old `make build`
build: $(prereqs)
	@echo "Building all binary platforms ($(GORELEASER_CONFIG))..."
	@$(GORELEASER_RUN) $(GORELEASER_DOCKER_BUILD_ARGS) build $(GR_CONFIG) --snapshot --clean --skip=validate $(SKIP_BEFORE_FLAG) --id=binary

build-gst: $(prereqs)
	@echo "Building all binary-gst platforms ($(GORELEASER_CONFIG))..."
	@$(GORELEASER_RUN) $(GORELEASER_DOCKER_BUILD_ARGS) build $(GR_CONFIG) --snapshot --clean --skip=validate $(SKIP_BEFORE_FLAG) --id=binary-gst

# TARGET=linux_amd64 or TARGET=linux_arm_7 (old style)
build-one:
ifndef TARGET
	$(error TARGET required, e.g. make build-one TARGET=linux_amd64)
endif
	@$(MAKE) binary GOOS=$(word 1,$(subst _, ,$(TARGET))) \
		GOARCH=$(word 2,$(subst _, ,$(TARGET))) \
		$(if $(word 3,$(subst _, ,$(TARGET))),GOARM=$(word 3,$(subst _, ,$(TARGET))),)

build-gst-one:
ifndef TARGET
	$(error TARGET required, e.g. make build-gst-one TARGET=linux_amd64)
endif
	@$(MAKE) binary-gst GOOS=$(word 1,$(subst _, ,$(TARGET))) \
		GOARCH=$(word 2,$(subst _, ,$(TARGET))) \
		$(if $(word 3,$(subst _, ,$(TARGET))),GOARM=$(word 3,$(subst _, ,$(TARGET))),)

ls-dist:
	@ls -la $(DIST_DIR)/ 2>/dev/null || echo "$(DIST_DIR)/ empty"

checksums:
	@test -f $(DIST_DIR)/checksums.txt && cat $(DIST_DIR)/checksums.txt \
		|| echo "No checksums — run make dist"

clean:
	rm -rf $(DIST_DIR) web/build $(DOCKER_CONTEXT_DIR)

clean-web:
	rm -rf web/build

clean-all: clean clean-web clean-cache
	cd server && go clean -cache -testcache

clean-cache:
	rm -rf $(CACHE_DIR)

# ── Backward-compatible aliases ───────────────────────────────────────────────

build-host: binary
gst: binary-gst
build-gst-host: binary-gst
build-no-hooks: SKIP_BEFORE=1
build-no-hooks: binary
release-snapshot: dist
release-no-docker:
	$(MAKE) dist SKIP=docker
release-no-docker-full:
	$(MAKE) dist-full SKIP=docker
data-sync: install
build-sync: binary install
web-deps:
	cd web && yarn
web-build:
	cd web && yarn run build
web-embed: webgen
install-goreleaser:
	go install github.com/goreleaser/goreleaser/v2@latest
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest
install-tools: install-goreleaser install-swag install-golangci-lint

help:
	@$(SHELL) -c '\
		b=$$(tput bold 2>/dev/null || true); \
		d=$$(tput dim 2>/dev/null || true); \
		c=$$(tput setaf 6 2>/dev/null || true); \
		g=$$(tput setaf 2 2>/dev/null || true); \
		y=$$(tput setaf 3 2>/dev/null || true); \
		r=$$(tput sgr0 2>/dev/null || true); \
		rule() { printf "%s\n" "$$d  ─────────────────────────────────────────────────────────────$$r"; }; \
		sec() { printf "\n  %s%s%s\n" "$$c" "$$b" "$$1"; rule; }; \
		tgt() { printf "  %s%-20s%s %s\n" "$$g" "$$1" "$$r" "$$2"; }; \
		opt() { printf "  %s%-22s%s %s\n" "$$y" "$$1" "$$r" "$$2"; }; \
		printf "\n"; \
		printf "  %s%s╔══════════════════════════════════════════════════════════════════════╗%s\n" "$$c" "$$b" "$$r"; \
		printf "  %s%s║%s  %-68s  %s%s║%s\n" "$$c" "$$b" "$$r" "$(PROJECT_NAME) — build system" "$$c" "$$b" "$$r"; \
		printf "  %s%s╚══════════════════════════════════════════════════════════════════════╝%s\n" "$$c" "$$b" "$$r"; \
		printf "\n"; \
		printf "  %sUsage:%s  make [target] [OPTIONS]          %sDocs:%s  docs/BUILD.md\n" "$$b" "$$r" "$$b" "$$r"; \
		sec "Build (host)"; \
		tgt "binary" "Host platform binary  (default)"; \
		tgt "binary-gst" "Host platform + GStreamer"; \
		tgt "build" "All platforms — $(GORELEASER_CONFIG)"; \
		tgt "build-one" "Single platform  TARGET=linux_amd64"; \
		tgt "build-gst" "All binary-gst platforms (local config)"; \
		tgt "build-gst-one" "Single gst build  TARGET=linux_amd64"; \
		sec "Release"; \
		tgt "dist" "Local snapshot — $(GORELEASER_CONFIG)"; \
		tgt "dist-full" "Full snapshot — $(GORELEASER_RELEASE_CONFIG)"; \
		tgt "release" "Tagged release + publish"; \
		tgt "release-no-docker" "dist without docker images"; \
		tgt "release-no-docker-full" "dist-full without docker images"; \
		tgt "android" "Android binary ($(GORELEASER_RELEASE_CONFIG))"; \
		sec "Docker"; \
		tgt "docker" "Build local image (UPX on)"; \
		tgt "docker-image-amd64" "Image for linux/amd64"; \
		tgt "docker-image-arm64" "Image for linux/arm64"; \
		tgt "docker-start" "Run local image, data/ mounted"; \
		tgt "docker-start-release" "Run $(IMAGE_RELEASE):$(DOCKER_TAG_RELEASE)"; \
		tgt "docker-push" "Push release image tag"; \
		tgt "setup-builder" "Build goreleaser builder image ($(IMAGE_BUILDER))"; \
		sec "Run & install"; \
		tgt "install" "Copy binary → data/$(RUNTIME_NAME)"; \
		tgt "start" "Run from data/"; \
		tgt "start-build" "binary + install + start"; \
		tgt "run" "Run from Go source (nosqlite)"; \
		sec "Web & API"; \
		tgt "webgen" "Build web assets and embed"; \
		tgt "webgen-clean" "Clean embed, then webgen"; \
		tgt "swag" "Regenerate swagger docs"; \
		tgt "update" "webgen + swag"; \
		sec "Utilities"; \
		tgt "ls-dist" "List dist/ contents"; \
		tgt "checksums" "Show dist/checksums.txt"; \
		tgt "version" "Toolchain and platform info"; \
		tgt "healthcheck" "GoReleaser dependency check"; \
		tgt "deps" "go mod download"; \
		tgt "fmt" "gofmt + golangci-lint fmt (write)"; \
		tgt "fmt-check" "Check gofmt + golangci-lint fmt (CI)"; \
		tgt "lint" "golangci-lint run ./server/..."; \
		tgt "lint-all" "fmt-check + lint"; \
		tgt "check" "Validate $(GORELEASER_CONFIG)"; \
		tgt "check-release" "Validate $(GORELEASER_RELEASE_CONFIG)"; \
		tgt "verify-linux-static" "Check static Linux ELF (id=binary)"; \
		tgt "verify-linux-gst" "Check gst Linux binary (runtime .so)"; \
		tgt "show-config" "Print local goreleaser config"; \
		tgt "show-config-release" "Print release goreleaser config"; \
		sec "Cleanup"; \
		tgt "clean" "Remove dist/"; \
		tgt "clean-web" "Remove embedded web assets"; \
		tgt "clean-all" "clean + clean-web + clean-cache"; \
		tgt "clean-cache" "Remove .cache/"; \
		sec "UPX compression"; \
		opt "binary, release, dist, dist-full" "off by default"; \
		opt "docker" "on by default"; \
		opt "ENABLE_UPX=1" "force on  (e.g. make ENABLE_UPX=1 release)"; \
		opt "SKIP_UPX=1" "force off (e.g. make SKIP_UPX=1 docker)"; \
		sec "Options"; \
		opt "USE_DOCKER_BUILDER=1" "goreleaser in Docker builder (UPX on)"; \
		opt "VERBOSE=1" "verbose goreleaser output"; \
		opt "SKIP=docker" "skip docker in dist/release"; \
		opt "GOOS / GOARCH / GOARM" "override target platform"; \
		opt "REGISTRY_IMAGE=o/r" "override GHCR slug (default: $(REGISTRY_IMAGE))"; \
		opt "DOCKER_IMAGE_ID=img" "override image (default: $(DOCKER_IMAGE_ID))"; \
		opt "DATA_DIR=path" "install dir (default: ./data)"; \
		opt "GORELEASER_CONFIG=file" "local config (default: $(GORELEASER_CONFIG))"; \
		opt "GORELEASER_RELEASE_CONFIG=file" "release config (default: $(GORELEASER_RELEASE_CONFIG))"; \
		printf "\n"; \
	'

.PHONY: binary binary-gst gst android dist dist-full release docker _docker_image _docker_build_linux docker-start docker-start-release docker-push docker-clean install start start-build \
        build build-one build-gst build-gst-one ls-dist checksums version healthcheck deps show-config \
        verify-linux-static verify-linux-gst \
        webgen webgen-clean swag update update-clean run check check-release clean clean-web clean-all clean-cache help \
        setup-builder setup-cache \
        guard-docker-amd64 guard-goreleaser guard-upx guard-ndk guard-yarn guard-swag \
        build-host build-gst-host build-no-hooks release-snapshot release-no-docker release-no-docker-full \
        data-sync build-sync web-deps web-build web-embed install-tools install-goreleaser install-swag install-golangci-lint \
        fmt fmt-check lint lint-all \
        docker-image docker-image-amd64 docker-image-arm64 show-config show-config-release
