setlocal
set DT=%DATE:/=%

set GOARCH=386
go build
zip -9 expect-%DT%-%GOARCH%.zip expect.exe
set GOARCH=amd64
go build
zip -9 expect-%DT%-%GOARCH%.zip expect.exe

endlocal
