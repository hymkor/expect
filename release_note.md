v0.11.0
-------
Feb.13,2024

- ([#35]) Add the new function: `sendvkey(VIRTUAL_KEYCODE)`.
    - It sends [a virtual key code](https://learn.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes)
- Fix the Go language version used for building to 1.20.14 for Windows 7,8,Server2008, and 2012R1/R2

[#35]: https://github.com/hymkor/expect/issues/35

### Example for sendvkey

```lua
local pid = spawn("cmd.exe")
if not pid then
    os.exit(1)
end
send("rem exit")
sleep(1)
sendvkey(0x24) -- HOME
sendvkey(0x2E) -- DELETE
sendvkey(0x2E) -- DELETE
sendvkey(0x2E) -- DELETE
sendvkey(0x2E) -- DELETE
sendln("")
wait(pid)
```

v0.10.0
-------
Jan.03,2023

- Add `create_object` and OLE functions.

v0.9.0
------
Dec.28,2022

- Add the new global variable: capturelines (default:2)
- Add the new function: sleep(SECOND),usleep(MICROSECOND)

v0.8.0
------
Dec.25,2022

- Add the new function: shot(N) that reads N-lines from the console buffer.

v0.7.1
------
Nov.29,2022

- (#31) Fix: expect() crashed on timeout

v0.7.0
------
Sep.25,2022

- Use the default-background-color `ESC[49m` instead of black `ESC[40m`
- Failed to call console-api, show API-name as error
- Add `-nologo` option
- expect(): when the console of STDOUT can not be read, try STDERR.
- (#30) expect(): Set matching information into the global variables: `_MATCHPOSITION` , `_MATCHLINE` , `_MATCH` , `_PREMATCH`, and `_POSTMATCH`.

v0.6.2
------
Feb.16,2022

- Fix: filehandle for script was not closed when script did not end with CRLF
- Show the version string

v0.6.1
------
Feb.14,2022

- For #23, fix that when a script did not end with CRLF, the last line was ignored.
- Include the source code of go-console in the package.
- readme.md and go.mod: change the URLs (Change username: zetamatta to hymkor)
- Fix warning for golint that the function does not have the header comments.

v0.6.0
------
Dec.18,2021

- Add the new function wait(PID) that waits the process of PID terminates.

v0.5.0
------
Apr.29,2021

- Implement Lua function: spawnctx()
    - The new function spawnctx is the similar one with spawn, but the process started by spawnctx is killed when Ctrl-C is pressed.

v0.4.0
------
Apr.15,2021

- Add a new function: kill(PROCESS-ID)
- spawn() returns PROCESS-ID on success , or nil on failure. (It returned true or false before )
- Remove import "io/ioutil" from the source file.
- (#20) -color=nerver args and batch suport long lines with the caret (^) (Thanks to @a690700752 )

v0.3.3
------
Dec.20,2019

- (#14) Fixed that wRepeatCount (the parameter for WriteConsoleInput) was not set 1 to send key-events.  
By this bug, some console applications cound not recieve keys from expect.exe .

v0.3.2
------
Dec.18,2018

- Fix bug that scripts embeded in batchfile could not be executed sometimes

v0.3.1
------
Aug.28,2018

- Rebuild with Go 1.11 (the files of the previous version are built with Go 1.10)

v0.3.0
------
Aug.08,2018

- Add send() the second parameter as mili-second to wait per 1 character sent #6
v0.2.0
------
Aug.07,2018

- #5, Implemented arg[] which are stored commandline arguments
- #4, Skip ByteOrderMark (EF BB BF) from source Lua script
- Lines startings with '@' are always skipped without -x option,

v0.1.1
------
Sep.21,2017

- Fix: runtime error: makeslice: len out of range #2
- Source file are not modified. You have to update your local source of go-getch to include the change hymkor/go-getch@a82c486

v0.1.0
------
Jun.27,2017

- colored
- expect() returns the number of found string
- Add echo()

v0.0.0
------
Jun.15,2017

- The first version
