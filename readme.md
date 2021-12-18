[![Go Report Card](https://goreportcard.com/badge/github.com/zetamatta/expect)](https://goreportcard.com/report/github.com/zetamatta/expect)

Expect.lua for Windows
======================

- A tool like `expect` on Linux.
- Scripts are to be written with Lua ([GopherLua](https://github.com/yuin/gopher-lua))
- Some built-in functions exists:
    - `RC=expect(A,B,C...)` accesses CONOUT$ directly and watches the cursor-line (0.1 seconds interval)
        - When A was found in cursor-line, RC=0
        - When B was found in cursor-line, RC=1
        - When C was found in cursor-line, RC=2
        - :
        - When error occured, RC=-1
        - When timeout occurs, RC=-2 (set variable like `timeout=(SECONDS)`,default 1 hour)
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

Install
-------

Download the binary package from [Releases](https://github.com/zetamatta/expect/releases) and extract the executable.

Sample
------

sample.lua:

```sample.lua
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

On the command prompt:

```console
$ expect.exe sample.lua
foo@example.com's password:
Last login: Thu Jun 15 13:21:57 2017 from XXXXXXXXXXXX.XXXX.XX.XXX.XXX.XXXXXXX.XX.XX
FreeBSD 9.1-RELEASE-p24 (XXXXXXXX) #0: Thu Feb  5 10:03:29 JST 2015

Welcome to FreeBSD!

[foo@XXXXXXX ~]$ exit
logout
Connection to example.com closed.
```

The script embeded in the batchfile:

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
