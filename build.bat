DEL A ./bin
mkdir "./bin"

SET GOARCH=amd64
SET GOOS=linux
SET CGO_ENABdeLED=0

echo "��ʼ��� hbuf linux..........."
go build -o ./bin/hbuf.linux ./pkg/compile/main.go
echo "��� hbuf linux�汾 �ɹ�"

SET GOOS=darwin

echo "��ʼ��� hbuf darwin..........."
go build -o ./bin/hbuf.darwin ./pkg/compile/main.go
echo "��� hbuf darwin�汾 �ɹ�"

SET CGO_ENABdeLED=0
SET GOOS=windows

echo "��ʼ��� hbuf window�汾..........."
go build -o ./bin/hbuf.exe ./pkg/compile/main.go
echo "��� hbuf window�汾 �ɹ�"s

copy D:\dev\3.hbuf\hbuf\bin\hbuf.exe D:\dev\p_game\hbuf.exe
copy D:\dev\3.hbuf\hbuf\bin\hbuf.darwin D:\dev\p_game\hbuf.darwin
copy D:\dev\3.hbuf\hbuf\bin\hbuf.linux D:\dev\p_game\hbuf.linux



