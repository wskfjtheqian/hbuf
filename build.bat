DEL A ./bin
mkdir "./bin"

SET CGO_ENABdeLED=0
SET GOARCH=amd64
SET GOOS=windows

echo "��ʼ��� hbuf ..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "��� hbuf �ɹ�"


