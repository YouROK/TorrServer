#!/bin/bash

PLATFORMS=""
PLATFORMS="$PLATFORMS linux/arm64 linux/arm7 linux/arm5"
PLATFORMS="$PLATFORMS linux/amd64 linux/386"
PLATFORMS="$PLATFORMS windows/amd64 windows/386"
PLATFORMS="$PLATFORMS darwin/amd64 darwin/arm64"
PLATFORMS="$PLATFORMS freebsd/amd64"
PLATFORMS="$PLATFORMS linux/mips linux/mipsle linux/mips64 linux/mips64le"

type setopt >/dev/null 2>&1

GOBIN="go"

$GOBIN version

$GOBIN run build_web.go

LDFLAGS="'-s -w'"
FAILURES=""
ROOT=${PWD}
OUTPUT="${ROOT}/dist/TorrServer"

cd "${ROOT}/server" || exit 1

$GOBIN clean -i -r -cache --modcache
$GOBIN mod tidy

BUILD_FLAGS="-ldflags=${LDFLAGS}"

#####################################
### X86 build section
#####

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  if [[ "$GOARCH" =~ arm({5,7}) ]]; then
    GOARCH="arm"
    GOARM="${BASH_REMATCH[1]}"
    GO_ARM="GOARM=${GOARM}"
  else
    GOARM=""
    GO_ARM=""
  fi
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
  # echo "$sound - ${animals[$sound]}"
  export CC="$NDK_TOOLCHAIN/bin/${COMPILERS[$GOARCH]}"
  export CXX="$NDK_TOOLCHAIN/bin/${COMPILERS[$GOARCH]}++"
  if [ "$GOARCH" = "arm7" ]; then
    GOARCH="arm"
    GOARM="7"
    GO_ARM="GOARM=${GOARM}"
  else
    GOARM=""
    GO_ARM=""
  fi
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
