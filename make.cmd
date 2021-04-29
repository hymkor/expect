@setlocal
@set "PROMPT=$G "
@pushd "%~dp0."
@for %%I in ("%CD%") do set "NAME=%%~nI"
@call :"%1"
@popd
@endlocal
@exit /b

:""
:"all"
    go fmt
    @for %%I in (386 amd64) do call :build %%I
    @exit /b

:build
    @if not exist "cmd"    mkdir "cmd"
    @if not exist "cmd\%1" mkdir "cmd\%1"
    set "GOARCH=%1"
    go build -o cmd\%1\%NAME%.exe -ldflags "-s -w"
    @exit /b

:"package"
    set /P "VERSION=Version ? "
    for %%I in (386 amd64) do zip -9j "%NAME%-%VERSION%-%%I.zip" "cmd\%%I\%NAME%.exe"
    @exit /b

:"get"
    go get -u
    go mod tidy
    @exit /b
