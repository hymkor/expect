[![Go Report Card](https://goreportcard.com/badge/github.com/hymkor/expect)](https://goreportcard.com/report/github.com/hymkor/expect)

Expect-lua for Windows
======================

- A tool like `expect` on Linux.
- The syntax of scripts is exactly same with Lua 5.1 except for some functions.
    - The reference manuals of Lua 5.1 exist in the Lua official site.  
        Please see [https://www.lua.org/docs.html](https://www.lua.org/docs.html)
    - Expect-lua uses [GopherLua](https://github.com/yuin/gopher-lua) as the VM for Lua.
- These functions are extended in Expect-lua
    - `RC=expect(A,B,C...)` accesses CONOUT$ directly and watches the cursor-line and abobe N-lines (N+1 can be set by `capturelines`:default is 2) (0.1 seconds interval)
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
    - `kill(PROCESS-ID)` kills the process. (since v0.4.0)
    - `spawnctx(NAME,ARG1,ARG2,...)` is similar with spawn() but the process started by spawnctx is killed automatically when Ctrl-C is pressed. (since v0.5.0)
    - `wait(PID)` waits the process of PID terminates.
    - `shot(N)` reads N-lines from the console buffer and returns them. (since v0.8.0)
    - `sleep(N)` sleeps N-seconds. (since v0.9.0)
    - `usleep(MICROSECOND)` sleep N-micrseconds. (since v0.9.0)
    - `local OBJ=create_object()` creates OLE-Object (since v0.10.0)
        - `OBJ:method(...)` calls method
        - `OBJ:_get("PROPERTY")` returns the value of the property.
        - `OBJ:_set("PROPERTY",value)` sets the value to the property.
        - `OBJ:_iter()` returns an enumerator of the collection.
        - `OBJ:_release()` releases the COM-instance.
        - `local N=to_ole_integer(10)` creates the integer value for OLE.

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

or

```
scoop bucket add hymkor https://github.com/hymkor/scoop-bucket
scoop install expect-lua
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

local sshpid = spawn(sshexe,"-p","22",account)
if not sshpid then
    print("ssh.exe is not found")
    os.exit(1)
end
timeout = 10
capturelines = 3 -- default is 2
local promptkeyword = "$"

while true do
    local rc = expect(
    "password:",
    "Are you sure you want to continue connecting (yes/no/[fingerprint])?",
    "Could not resolve hostname")

    if rc == 0 then
        sendln(password)
        sleep(1)
        rc = expect(promptkeyword,"Permission denied, please try again")
        if rc == 0 then
            sleep(1)
            sendln("exit")
            sleep(1)
            echo()
            echo("-- This is the sample script")
            echo("-- Write your code to do in your server instead of sendln(\"exit\")")
        else
            print()
            echo(string.format(
            "-- The expected prompt keyword(%s) was not found.",
            promptkeyword))

            -- kill(sshpid)
        end
        break
    elseif rc == 1 then
        sendln("yes")
        print() -- move cursor down not to capture same keyword
    elseif rc == -2 then
        echo("TIMEOUT")
        break
    elseif rc == -1 then
        echo("ERROR")
        break
    else
        echo(string.format("Error keyword found \"%s\". Exit",_MATCH))
        break
    end
end
```

On the command prompt:

```console
$ .\expect sample.lua example@example.com PASSW0RD
Expect-lua v0.10.0-3-ga2986b0-windows-amd64
example@example.com's password:
Last login: Mon Dec 26 23:18:11 2022 from XXXXXXXX-XXXXX.XXXX.XX.XXX.XXX.XXX.XX.XX
FreeBSD 9.1-RELEASE-p24 (XXXXXXXX) #0: Thu Feb  5 10:03:29 JST 2015

Welcome to FreeBSD!

[example@XXXXXXX ~]$ exit
logout
Connection to example.com closed.
-- This is the sample script
-- Write your code to do in your server instead of sendln("exit")
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

FAQ
---

### `expect sample.lua > log` does not work.

Expect.lua directly accesses the terminal device linked to STDOUT 
or STDERR to retrieve the printed text.
Therefore, if the STDOUT/STDERR is redirected to something other 
than the terminal, the screen data of the terminal cannot be 
accessed, resulting in an error.
This is difficult to solve, and it is currently impossible to work
with redirects.

It was ideal that like "expect" on Linux, Expect.lua got the stdout
and stderr of the child process via a pipeline,
and then re-output them to the original destination after parsing .
However, it is unavailable because on Windows the output is suppressed
until CRLF is found when the destination is not the terminal and the expect function can never get a prompt without a newline.
