#!/bin/bash

PLATFORMS=(
  'linux/arm64' 'linux/arm7' 'linux/amd64'
  'linux/arm5' 'linux/386'
  'windows/amd64' 'windows/386'
  'darwin/amd64' 'darwin/arm64'
  'freebsd/amd64'
  'linux/mips' 'linux/mipsle' 'linux/mips64' 'linux/mips64le'
)

type setopt >/dev/null 2>&1

set_goarm() {
  if [[ "$1" =~ arm([5,7]) ]]; then
    GOARCH="arm"
    GOARM="${BASH_REMATCH[1]}"
    GO_ARM="GOARM=${GOARM}"
  else
    GOARM=""
    GO_ARM=""
  fi
}

GOBIN="go"

$GOBIN version

LDFLAGS="'-s -w'"
FAILURES=""
ROOT=${PWD}
OUTPUT="${ROOT}/dist/TorrServer"

#### Build web
echo "Build web"
cd "${ROOT}/web" || exit 1
npm run --silent build-js
cp "${ROOT}/web/dest/index.html" "${ROOT}/server/web/pages/template/pages/"

#### Build server
echo "Build server"
cd "${ROOT}/server" || exit 1
$GOBIN clean -i -r -cache #--modcache
$GOBIN mod tidy

BUILD_FLAGS="-ldflags=${LDFLAGS}"

#####################################
### X86 build section
#####

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  set_goarm "$GOARCH"
  BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
  if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
  CMD="GOOS=${GOOS} GOARCH=${GOARCH} ${GO_ARM} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
  echo "${CMD}"
  eval "$CMD" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"
done

#####################################
### Android build section
#####

declare -A COMPILERS=(
  ["arm7"]="armv7a-linux-androideabi21-clang"
  ["arm64"]="aarch64-linux-android21-clang"
  ["386"]="i686-linux-android21-clang"
  ["amd64"]="x86_64-linux-android21-clang"
)

export NDK_TOOLCHAIN=$ROOT/toolchain

GOOS=android

for GOARCH in "${!COMPILERS[@]}"; do
  export CC="$NDK_TOOLCHAIN/bin/${COMPILERS[$GOARCH]}"
  export CXX="$NDK_TOOLCHAIN/bin/${COMPILERS[$GOARCH]}++"
  set_goarm "$GOARCH"
  BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
  CMD="GOOS=${GOOS} GOARCH=${GOARCH} ${GO_ARM} CGO_ENABLED=1 ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
  echo "${CMD}"
  eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"
done

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "failed on: ${FAILURES}"
  exit 1
fi
