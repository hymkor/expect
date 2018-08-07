@setlocal
@call :"%1"
@endlocal
@exit /b

:""
    go fmt
    @setlocal
    @for %%I in (386 amd64) do call :buildas %%I
    @endlocal
    @exit /b

:buildas
    @if not exist cmd mkdir cmd
    @if not exist "cmd\%1" mkdir "cmd\%1"
    @setlocal
    set "GOARCH=%1"
    go build -o cmd\%1\expect.exe
    @endlocal
    @exit /b

:"package"
    for %%I in (386 amd64) do zip -9j expect-%DATE:/=%-%%I.zip cmd\%%I\expect.exe
    @exit /b
