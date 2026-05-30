#!/bin/bash

LDFLAGS="-s -w"
TAGS=""
SRC="./cmd/tuns/main.go"

echo "Starting compilation for all platforms..."

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-aarch64 $SRC
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-armv7 $SRC
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-armv5 $SRC
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-mipsle $SRC
CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-mips $SRC
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-x64 $SRC
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-x86 $SRC

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-x64.exe $SRC
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -trimpath -tags=$TAGS -ldflags="$LDFLAGS" -o dist/tuns-x86.exe $SRC

echo "Compilation finished. Check the dist folder."