#!/bin/bash

ROOT=${PWD}

#### Build web
echo "Build web"
cd "${ROOT}/web" || exit 1
npm install --silent
npm run --silent build-js
cp "${ROOT}/web/dest/index.html" "${ROOT}/server/web/pages/template/pages/"
cd ..

sudo docker run --rm -v "$PWD":/usr/src/torr -w /usr/src/torr golang:1.16 ./build-all.sh
sudo chmod 0777 ./dist/*