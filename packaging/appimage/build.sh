#!/usr/bin/env bash
# Build a Linux AppImage for one GoReleaser binary.
# No-op (exit 0) unless linux amd64 / arm64 / arm7 and packaging tools can run.
# Safe on macOS and inside the alpine builder image where FUSE/appimagetool are missing.
set -u

BINARY_PATH="${1:-${GORELEASER_ARTIFACT_PATH:-}}"
GOOS_VAL="${2:-${GOOS:-}}"
GOARCH_VAL="${3:-${GOARCH:-}}"
GOARM_VAL="${4:-${GOARM:-}}"

if [[ "${GOOS_VAL}" != "linux" ]]; then
  exit 0
fi

ARCH_NAME=""
APPIMAGE_ARCH=""
case "${GOARCH_VAL}" in
  amd64)
    ARCH_NAME="amd64"
    APPIMAGE_ARCH="x86_64"
    ;;
  arm64)
    ARCH_NAME="arm64"
    APPIMAGE_ARCH="aarch64"
    ;;
  arm)
    if [[ "${GOARM_VAL}" == "7" ]]; then
      ARCH_NAME="arm7"
      APPIMAGE_ARCH="armhf"
    else
      exit 0
    fi
    ;;
  *)
    exit 0
    ;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST_DIR="${REPO_ROOT}/dist"
DESKTOP="${SCRIPT_DIR}/torrserver.desktop"
APPRUN="${SCRIPT_DIR}/AppRun"
ICON="${REPO_ROOT}/web/public/icon.png"

if [[ -z "${BINARY_PATH}" ]]; then
  BINARY_PATH="${DIST_DIR}/TorrServer-linux-${ARCH_NAME}"
fi
if [[ ! -f "${BINARY_PATH}" ]]; then
  echo "appimage: skip (binary not found: ${BINARY_PATH})" >&2
  exit 0
fi
if [[ ! -f "${DESKTOP}" || ! -f "${APPRUN}" || ! -f "${ICON}" ]]; then
  echo "appimage: skip (desktop/AppRun/icon missing)" >&2
  exit 0
fi

mkdir -p "${DIST_DIR}"
APPIMAGE_TOOL="${DIST_DIR}/appimagetool"
if [[ ! -x "${APPIMAGE_TOOL}" ]]; then
  if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
    echo "appimage: skip (curl/wget not available)" >&2
    exit 0
  fi
  URL="https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "${APPIMAGE_TOOL}" "${URL}" || { echo "appimage: skip (download failed)" >&2; exit 0; }
  else
    wget -q -O "${APPIMAGE_TOOL}" "${URL}" || { echo "appimage: skip (download failed)" >&2; exit 0; }
  fi
  chmod +x "${APPIMAGE_TOOL}"
fi

APPDIR="$(mktemp -d "${TMPDIR:-/tmp}/TorrServer.AppDir.XXXXXX")"
cleanup() { rm -rf "${APPDIR}"; }
trap cleanup EXIT

mkdir -p "${APPDIR}/usr/bin"
cp "${BINARY_PATH}" "${APPDIR}/usr/bin/torrserver"
chmod +x "${APPDIR}/usr/bin/torrserver"
cp "${DESKTOP}" "${APPDIR}/torrserver.desktop"
cp "${ICON}" "${APPDIR}/torrserver.png"
cp "${APPRUN}" "${APPDIR}/AppRun"
chmod +x "${APPDIR}/AppRun"

OUT="${DIST_DIR}/TorrServer-linux-${ARCH_NAME}.AppImage"
export APPIMAGE_EXTRACT_AND_RUN="${APPIMAGE_EXTRACT_AND_RUN:-1}"
if ! ARCH="${APPIMAGE_ARCH}" "${APPIMAGE_TOOL}" "${APPDIR}" "${OUT}"; then
  echo "appimage: skip (appimagetool failed for ${ARCH_NAME})" >&2
  rm -f "${OUT}"
  exit 0
fi

echo "appimage: wrote ${OUT}"
exit 0
