DEL A ./bin
mkdir "./bin"

SET GOARCH=amd64
SET GOOS=linux

echo "��ʼ��� hbuf linux..........."
go build -o ./bin/hbuf_linux ./pkg/compile/main.go
echo "��� hbuf linux�汾 �ɹ�"

SET GOOS=darwin

echo "��ʼ��� hbuf darwin..........."
go build -o ./bin/hbuf_darwin ./pkg/compile/main.go
echo "��� hbuf darwin�汾 �ɹ�"

SET CGO_ENABdeLED=0
SET GOOS=windows

echo "��ʼ��� hbuf window�汾..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "��� hbuf window�汾 �ɹ�"s

copy E:\dev\3.hbuf\hbuf\bin\hbuf.exe E:\dev\9.slot\hbuf.exe
copy E:\dev\3.hbuf\hbuf\bin\hbuf_darwin E:\dev\9.slot\hbuf_darwin
copy E:\dev\3.hbuf\hbuf\bin\hbuf_linux E:\dev\9.slot\hbuf_linux






