#!/bin/bash
workspace=$(pwd)
export GOPATH=$workspace/vender:$GOPATH

go build -ldflags "-X main.buildstamp=`date '+%Y-%m-%d_%I:%M:%S'` -X main.githash=`git rev-parse HEAD` -X main.goversion=`go version|awk '{print $3}'`" -o zk-proxy main.go
