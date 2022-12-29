@setlocal
@set "PROMPT=$$ "
@pushd "%~dp0."
@for %%I in ("%CD%") do @set "NAME=%%~nI"
@call :"%1"
@popd
@endlocal
@exit /b

:""
    go fmt
    for /F %%I in ('git.exe describe --tags') do set "VERSION=%%I"
    go build -ldflags "-s -w -X main.version=%VERSION%"
    exit /b

:"all"
    go fmt
    @for %%I in (386 amd64) do call :build %%I
    @exit /b

:build
    @if not exist "bin"    mkdir "bin"
    @if not exist "bin\%1" mkdir "bin\%1"
    set "GOARCH=%1"
    for /F %%I in ('git.exe describe --tags') do set "VERSION=%%I"
    go build -o bin\%1\%NAME%.exe -ldflags "-s -w -X main.version=%VERSION%"
    @exit /b

:"package"
    call :"all"
    for /F %%I in ('git describe --tags') do set "VERSION=%%I"
    for %%I in (386 amd64) do zip -9j "%NAME%-%VERSION%-%%I.zip" "bin\%%I\%NAME%.exe"
    @exit /b

:"get"
    go get -u
    go mod tidy
    @exit /b

:"install"
    for /F "skip=1" %%I in ('where expect.exe') do copy /-Y expect.exe "%%I"
    exit /b

:"test"
    @echo off
    expect.exe test.lua
    if errorlevel 1 (
        echo TEST FAILED
    ) else (
        echo TEST SUCCEEDED
    )
    exit /b

:"manifest"
    go run mkmanifest.go expect-v*.zip > expect-lua.json
    exit /b
