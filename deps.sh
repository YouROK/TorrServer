#!/bin/bash

export GOPATH="${PWD}"
export GO111MODULE=auto

go get -v github.com/pkg/browser
go get -v github.com/alexflint/go-arg
go get -v golang.org/x/time/rate
go get -v github.com/anacrolix/dht
go get -v github.com/anacrolix/missinggo/httptoo
go get -v github.com/anacrolix/torrent
go get -v github.com/anacrolix/torrent/iplist
go get -v github.com/anacrolix/torrent/metainfo
go get -v github.com/anacrolix/utp
go get -u github.com/gin-gonic/gin
go get -v github.com/pion/webrtc/v2
go get -v go.etcd.io/bbolt

ln -s . src/github.com/pion/webrtc/v2
go get -v github.com/pion/webrtc/v2

go get -v github.com/gin-contrib/cors