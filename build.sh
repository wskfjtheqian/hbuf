#!/bin/zsh

rm -rf ./bin
mkdir "./bin"

export GOARCH=amd64
export GOOS=linux

echo "开始打包 hbuf linux..........."
go build -o ./bin/hbuf.linux ./pkg/compile/main.go
chmod 777 ./bin/hbuf.linux
echo "打包 hbuf linux版本 成功"

export GOOS=darwin

echo "开始打包 hbuf darwin..........."
go build -o ./bin/hbuf.darwin ./pkg/compile/main.go
chmod 777 ./bin/hbuf.darwin
echo "打包 hbuf darwin版本 成功"
cp ./bin/hbuf.darwin  /Users/dev/7.hanber/hbuf.darwin

export CGO_ENABdeLED=0
export GOOS=windows

echo "开始打包 hbuf window版本..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "打包 hbuf window版本 成功"
#
#cp ./bin/hbuf.exe /Users/dev/5.client-web/hbuf.exe
#cp ./bin/hbuf.darwin /Users/dev/5.client-web/hbuf.darwin
#cp ./bin/hbuf.linux /Users/dev/5.client-web/hbuf.linux
#
#cp ./bin/hbuf.exe /Users/dev/7.hanber/hbuf.exe
#cp ./bin/hbuf.darwin /Users/dev/7.hanber/hbuf.darwin
#cp ./bin/hbuf.linux /Users/dev/7.hanber/hbuf.linux
#
#cp ./bin/hbuf.exe /Users/dev/10.account/hbuf.exe
#cp ./bin/hbuf.darwin /Users/dev/10.account/hbuf.darwin
#cp ./bin/hbuf.linux /Users/dev/10.account/hbuf.linux

cp ./bin/hbuf.exe /Users/dev/8.p_game/hbuf.exe
cp ./bin/hbuf.darwin /Users/dev/8.p_game/hbuf.darwin
cp ./bin/hbuf.linux /Users/dev/8.p_game/hbuf.linux

chmod 777 ./bin/*

