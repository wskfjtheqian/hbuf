DEL A ./bin
mkdir "./bin"

SET GOARCH=amd64
SET GOOS=linux

echo "开始打包 hbuf linux..........."
go build -o ./bin/hbuf_linux ./pkg/compile/main.go
echo "打包 hbuf linux版本 成功"

SET GOOS=darwin

echo "开始打包 hbuf darwin..........."
go build -o ./bin/hbuf_darwin ./pkg/compile/main.go
echo "打包 hbuf darwin版本 成功"

SET CGO_ENABdeLED=0
SET GOOS=windows

echo "开始打包 hbuf window版本..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
copy D:\dev\1.hbuf\hbuf\bin\hbuf.exe D:\dev\9.recruit\hbuf.exe
copy D:\dev\1.hbuf\hbuf\bin\hbuf_darwin D:\dev\9.recruit\hbuf_darwin

echo "打包 hbuf window版本 成功"s




