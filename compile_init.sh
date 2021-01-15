#!/bin/bash
set -x
export PATH=$PATH:/usr/local/go/bin/
export GOPATH=`pwd`
export ANDROID_HOME=$HOME'/Android/Sdk'
go get golang.org/x/mobile/cmd/gomobile
./bin/gomobile init -v -ndk /home/yourok/Android/Ndk/android-ndk/
