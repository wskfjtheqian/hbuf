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
copy D:\dev\1.hbuf\hbuf\bin\hbuf.exe D:\dev\game_news\hbuf.exe
echo "��� hbuf window�汾 �ɹ�"s




