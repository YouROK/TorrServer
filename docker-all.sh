#!/bin/bash
sudo docker run --rm -v "$PWD":/usr/src/torr -w /usr/src/torr golang:1.16 ./build-all.sh
sudo chmod 0777 ./dist/*