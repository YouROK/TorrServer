#!/bin/bash

ROOT=${PWD}

#### Build web
#echo "Build web"
#go run gen_web.go

sudo docker run --rm -v "$PWD":/usr/src/torr -v ~/go/pkg/mod:/go/pkg/mod -w /usr/src/torr golang:1.17.5-stretch ./build-all.sh
sudo chmod 0777 ./dist/*