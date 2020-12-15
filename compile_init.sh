#!/bin/bash
set -x

export PATH=$PATH:/usr/local/go/bin/
export GOPATH=`pwd`
export ANDROID_HOME=$HOME'/android-sdk'
export ANDROID_NDK_HOME=$ANDROID_HOME'/ndk-bundle'
export PATH=$PATH:$ANDROID_NDK_HOME'/ndk-build'
go get golang.org/x/mobile/cmd/gomobile
./bin/gomobile init -v
