DEL A ./bin
mkdir "./bin"

SET CGO_ENABdeLED=0
SET GOARCH=amd64
SET GOOS=windows

echo "开始打包 hbuf ..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "打包 hbuf 成功"


