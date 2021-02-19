#!/bin/bash

PLATFORMS=""
PLATFORMS="$PLATFORMS linux/amd64 linux/386"
PLATFORMS="$PLATFORMS windows/amd64 windows/386"
PLATFORMS="$PLATFORMS darwin/amd64 darwin/arm64"
PLATFORMS="$PLATFORMS freebsd/amd64"
PLATFORMS="$PLATFORMS linux/mips linux/mipsle linux/mips64 linux/mips64le"

type setopt >/dev/null 2>&1

GOBIN="/usr/local/go/bin/go"
#GOBIN="/usr/local/go116b/bin/go"

$GOBIN version

$GOBIN run build_web.go

LDFLAGS="'-s -w'"
FAILURES=""
ROOT=${PWD}
OUTPUT="${ROOT}/dist/TorrServer"

cd "${ROOT}/server"
$GOBIN clean -i -r -cache
rm -f "${ROOT}/dist/TorrServer*"

$GOBIN mod tidy

BUILD_FLAGS="-tags disable_libutp -ldflags=${LDFLAGS}"

#####################################
### ARM build section
#####

GOOS="linux"
GOARCH="arm64"
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}"
CMD="GOOS=linux GOARCH=${GOARCH} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"

GOARCH="arm"
GOARM="7"
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
CMD="GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"

#####################################
### X86 build section
#####

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}"
  if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
  if [[ "${GOOS}" == "linux" ]]; then
    CMD="GOOS=${GOOS} GOARCH=${GOARCH} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
  else
    CMD="GOOS=${GOOS} GOARCH=${GOARCH} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
  fi
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
done

#####################################
### Android build section
#####

export NDK_TOOLCHAIN=$ROOT/toolchain

GOOS=android

export CC=$NDK_TOOLCHAIN/bin/armv7a-linux-androideabi21-clang
export CXX=$NDK_TOOLCHAIN/bin/armv7a-linux-androideabi21-clang++
GOARCH="arm"
GOARM="7"
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
CMD="GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} CGO_ENABLED=1 ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"

export CC=$NDK_TOOLCHAIN/bin/aarch64-linux-android21-clang
export CXX=$NDK_TOOLCHAIN/bin/aarch64-linux-android21-clang++
GOARCH="arm64"
GOARM=""
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
CMD="GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"

export CC=$NDK_TOOLCHAIN/bin/i686-linux-android21-clang
export CXX=$NDK_TOOLCHAIN/bin/i686-linux-android21-clang++
GOARCH="386"
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
CMD="GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"

export CC=$NDK_TOOLCHAIN/bin/x86_64-linux-android21-clang
export CXX=$NDK_TOOLCHAIN/bin/x86_64-linux-android21-clang++
GOARCH="amd64"
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
CMD="GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "failed on: ${FAILURES}"
  exit 1
fi
