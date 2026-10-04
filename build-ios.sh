#!/bin/bash
# Build TorrServerKit.xcframework for iOS (macOS + Xcode required).
# Usage:
#   ./build-ios.sh
#   VERSION=MatriX.144 ./build-ios.sh

set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "build-ios.sh must run on macOS with Xcode" >&2
  exit 1
fi

# gomobile needs full Xcode (not Command Line Tools). Prefer DEVELOPER_DIR if set;
# otherwise switch to Xcode.app when xcode-select points at CLT.
if [[ -z "${DEVELOPER_DIR:-}" ]]; then
  active="$(xcode-select -p 2>/dev/null || true)"
  if [[ "${active}" == *"/CommandLineTools"* ]] || ! xcodebuild -version &>/dev/null; then
    if [[ -d "/Applications/Xcode.app/Contents/Developer" ]]; then
      export DEVELOPER_DIR="/Applications/Xcode.app/Contents/Developer"
      echo "=== Using DEVELOPER_DIR=${DEVELOPER_DIR} ==="
    else
      echo "Full Xcode is required (xcode-select currently: ${active:-unknown})." >&2
      echo "Install Xcode, then run: sudo xcode-select -s /Applications/Xcode.app/Contents/Developer" >&2
      exit 1
    fi
  fi
fi
if ! xcodebuild -version &>/dev/null; then
  echo "xcodebuild is not usable. Point xcode-select at Xcode.app:" >&2
  echo "  sudo xcode-select -s /Applications/Xcode.app/Contents/Developer" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST="${ROOT}/dist"
GOBIN="${GOBIN:-go}"
IOS_VERSION_MIN="${IOS_VERSION_MIN:-18.0}"
export NODE_OPTIONS="${NODE_OPTIONS:---openssl-legacy-provider}"

if [[ -z "${VERSION:-}" && -f "${ROOT}/server/version/version.go" ]]; then
  VERSION="$(sed -n 's/.*var Version = "\([^"]*\)".*/\1/p' "${ROOT}/server/version/version.go" | head -1)"
fi
VERSION="${VERSION:-dev}"

LDFLAGS="-s -w -checklinkname=0 -X server/version.Version=${VERSION}"
XCFRAMEWORK="${DIST}/TorrServerKit.xcframework"
ZIP_NAME="TorrServer-ios-TorrServerKit.xcframework.zip"

echo "=== TorrServer iOS XCFramework (${VERSION}), min iOS ${IOS_VERSION_MIN} ==="
mkdir -p "${DIST}"

echo "=== Build web ==="
cd "${ROOT}"
$GOBIN run gen_web.go

echo "=== Build swagger docs ==="
$GOBIN install github.com/swaggo/swag/cmd/swag@latest
cd "${ROOT}/server"
"$(go env GOPATH)/bin/swag" init -g web/server.go --parseInternal --parseDepth 5

echo "=== Install gomobile ==="
$GOBIN install golang.org/x/mobile/cmd/gomobile@latest
$GOBIN install golang.org/x/mobile/cmd/gobind@latest
export PATH="$(go env GOPATH)/bin:${PATH}"
gomobile init

echo "=== gomobile bind ==="
rm -rf "${XCFRAMEWORK}"
cd "${ROOT}/server/mobile/torrserverkit"
go mod tidy
gomobile bind \
  -target=ios,iossimulator \
  -iosversion="${IOS_VERSION_MIN}" \
  -tags=nosqlite \
  -ldflags="${LDFLAGS}" \
  -trimpath \
  -o "${XCFRAMEWORK}" \
  .

echo "=== Zip XCFramework ==="
cd "${DIST}"
rm -f "${ZIP_NAME}"
zip -r -9 "${ZIP_NAME}" TorrServerKit.xcframework
echo "Done: ${DIST}/${ZIP_NAME}"
