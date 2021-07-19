#!/bin/bash

ROOT=${PWD}

#### Build web
echo "Build web"
$GOBIN run gen_web.go

sudo docker run --rm -v "$PWD":/usr/src/torr -w /usr/src/torr golang:1.16 ./build-all.sh
sudo chmod 0777 ./dist/*