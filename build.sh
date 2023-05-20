#!/bin/zsh

DEL A ./bin
mkdir "./bin"

export GOARCH=amd64
export GOOS=linux

echo "开始打包 hbuf linux..........."
go build -o ./bin/hbuf_linux ./pkg/compile/main.go
echo "打包 hbuf linux版本 成功"

export GOOS=darwin

echo "开始打包 hbuf darwin..........."
go build -o ./bin/hbuf_darwin ./pkg/compile/main.go
echo "打包 hbuf darwin版本 成功"
cp ./bin/hbuf_darwin /Users/heqian/dev/hbuf/hbuf_frame/hbuf_darwin
cp ./bin/hbuf_darwin /Users/heqian/dev/apk_rebuild/hbuf_darwin
cp ./bin/hbuf_darwin /Users/heqian/dev/h_im/hbuf_darwin

export CGO_ENABdeLED=0
export GOOS=windows

echo "开始打包 hbuf window版本..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "打包 hbuf window版本 成功"


chmod 777 ./bin/*

