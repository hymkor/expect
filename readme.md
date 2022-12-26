[![Go Report Card](https://goreportcard.com/badge/github.com/hymkor/expect)](https://goreportcard.com/report/github.com/hymkor/expect)

Expect-lua for Windows
======================

- A tool like `expect` on Linux.
- The syntax of scripts is exactly same with Lua 5.1 except for some functions.
    - The reference manuals of Lua 5.1 exist in the Lua official site.  
        Please see [https://www.lua.org/docs.html](https://www.lua.org/docs.html)
    - Expect-lua uses [GopherLua](https://github.com/yuin/gopher-lua) as the VM for Lua.
- These functions are extended in Expect-lua
    - `RC=expect(A,B,C...)` accesses CONOUT$ directly and watches the cursor-line (0.1 seconds interval)
        - When A was found in cursor-line, RC=0
        - When B was found in cursor-line, RC=1
        - When C was found in cursor-line, RC=2
        - :
        - When error occured, RC=-1
        - When timeout occurs, RC=-2 (set variable like `timeout=(SECONDS)`,default 1 hour)
        - When RC &gt;= 0, these global variables are set.
            - `_MATCH` - The string matched.
            - `_MATCHPOSITION` - The position where matched.
            - `_MATCHLINE` - The matched whole line.
            - `_PREMATCH` - The string preceding matched.
            - `_POSTMATCH` - The string following matched.
    - `send(TEXT)` sends TEXT to the terminal as keyboard events.
        - `send(TEXT,MS)` waits MS [m-seconds] per 1-character (for plink.exe)
    - `sendln()` is same as send() but append CR.
    - `PID=spawn(NAME,ARG1,ARG2,...)` starts applications and
        - On success, `PID` is process-id(integer).
        - On failure, `PID` is nil.
    - `echo()` controls echoback
        - `echo(true)`: echo on
        - `echo(false)`: echo off
        - `echo("...")`: print a string
    - `arg[]` contains commandline arguments (`arg[0]` is scriptname)
    - `kill(PROCESS-ID)` kills the process. (v0.4.0~)
    - `spawnctx(NAME,ARG1,ARG2,...)` is similar with spawn() but the process started by spawnctx is killed automatically when Ctrl-C is pressed. (v0.5.0~)
    - `wait(PID)` waits the process of PID terminates.
    - `shot(N)` reads N-lines from the console buffer and returns them. (v0.8.0~)

``` lua
local screen = assert(shot(25))
for i = 1,#screen do
    print( i,screen[i] )
end
```

Install
-------

Download the binary package from [Releases](https://github.com/hymkor/expect/releases) and extract the executable.

### for scoop-installer

```
scoop install https://raw.githubusercontent.com/hymkor/expect/master/expect-lua.json
```

Sample
------

sample.lua:

``` lua
if #arg < 2 then
    print("expect.exe sample.lua USERNAME@DOMAIN PASSWD")
    os.exit(0)
end
local account = arg[1]
local password = arg[2]
local sshexe = os.getenv("windir") .. "\\System32\\OpenSSH\\ssh.exe"

spawn(sshexe,"-p","22",account)
timeout = 10

while true do
    local rc = expect(
    "Are you sure you want to continue connecting (yes/no/[fingerprint])?",
    "password:")

    if rc == 0 then
        sendln("yes")
    elseif rc == 1 then
        sendln(password)
        rc = expect("~]$")
        if rc == 0 then
            sendln("exit")
        end
        break
    else
        if _MATCH then
            echo(string.format("Error keyword found \"%s\". Exit",_MATCH))
        else
            echo("TIMEOUT")
        end
        break
    end
end
```

On the command prompt:

```console
$ .\expect sample.lua example@example.com PASSW0RD
Expect-lua v0.8.0-6-g456fe3e-windows-amd64
example@example.com's password:
Last login: Mon Dec 26 23:18:11 2022 from XXXXXXXX-XXXXX.XXXX.XX.XXX.XXX.XXX.XX.XX
FreeBSD 9.1-RELEASE-p24 (XXXXXXXX) #0: Thu Feb  5 10:03:29 JST 2015

Welcome to FreeBSD!

[example@XXXXXXX ~]$ exit
logout
Connection to example.com closed.
$
```

The script embedded in the batchfile:

```sample.cmd
@expect.exe "%~f0"
@exit /b

-- Lines starting with '@' are replaced to '--@' by expect.exe
-- to embed the script into the batchfile.

echo(true)
if spawn([[c:\Program Files\Git\usr\bin\ssh.exe]],"foo@example.com") then
    expect("password:")
    echo(false)
    send("PASSWORD\r")
    expect("~]$")
    echo(true)
    send("exit\r")
end
```
