#!/bin/bash

rm -rf ./bin
mkdir "./bin"

export GOARCH=amd64
export GOOS=linux

echo "开始打包 hbuf linux..........."
go build -o ./bin/hbuf_linux ./pkg/compile/main.go
echo "打包 hbuf linux版本 成功"
cp ./bin/hbuf_linux /home/heqian/dev/1.recruit/hbuf_linux

export GOOS=darwin

echo "开始打包 hbuf darwin..........."
go build -o ./bin/hbuf_darwin ./pkg/compile/main.go
echo "打包 hbuf darwin版本 成功"
cp ./bin/hbuf_darwin /home/heqian/dev/1.recruit/hbuf_darwin

export CGO_ENABdeLED=0
export GOOS=windows

echo "开始打包 hbuf window版本..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "打包 hbuf window版本 成功"
cp ./bin/hbuf.exe /home/heqian/dev/1.recruit/hbuf.exe

chmod 777 ./bin/*

