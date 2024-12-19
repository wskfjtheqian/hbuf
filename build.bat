DEL A ./bin
mkdir "./bin"

SET GOARCH=amd64
SET GOOS=linux
SET CGO_ENABdeLED=0

echo "开始打包 hbuf linux..........."
go build -o ./bin/hbuf.linux ./pkg/compile/main.go
echo "打包 hbuf linux版本 成功"

SET GOOS=darwin

echo "开始打包 hbuf darwin..........."
go build -o ./bin/hbuf.darwin ./pkg/compile/main.go
echo "打包 hbuf darwin版本 成功"

SET CGO_ENABdeLED=0
SET GOOS=windows

echo "开始打包 hbuf window版本..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "打包 hbuf window版本 成功"s

copy E:\develop\hbuf\hbuf\bin\hbuf.exe E:\develop\client-web\hbuf.exe
copy E:\develop\hbuf\hbuf\bin\hbuf.darwin E:\develop\client-web\hbuf.darwin
copy E:\develop\hbuf\hbuf\bin\hbuf.linux E:\develop\client-web\hbuf.linux

copy E:\develop\hbuf\hbuf\bin\hbuf.exe E:\develop\hanber\hbuf.exe
copy E:\develop\hbuf\hbuf\bin\hbuf.darwin E:\develop\hanber\hbuf.darwin
copy E:\develop\hbuf\hbuf\bin\hbuf.linux E:\develop\hanber\hbuf.linux





