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

copy E:\develop\hbuf\hbuf\bin\hbuf.exe E:\develop\client-web\hbuf.exe
copy E:\develop\hbuf\hbuf\bin\hbuf.darwin E:\develop\client-web\hbuf.darwin
copy E:\develop\hbuf\hbuf\bin\hbuf.linux E:\develop\client-web\hbuf.linux

copy E:\develop\hbuf\hbuf\bin\hbuf.exe E:\develop\hanber\hbuf.exe
copy E:\develop\hbuf\hbuf\bin\hbuf.darwin E:\develop\hanber\hbuf.darwin
copy E:\develop\hbuf\hbuf\bin\hbuf.linux E:\develop\hanber\hbuf.linux





