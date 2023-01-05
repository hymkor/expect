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
