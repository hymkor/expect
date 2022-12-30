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

while true do
    local rc = expect(
    "password:",
    "Are you sure you want to continue connecting (yes/no/[fingerprint])?",
    "Could not resolve hostname")

    if rc == 0 then
        sendln(password)
        rc = expect("~]$","Permission denied, please try again")
        if rc == 0 then
            sendln("exit")
        else
            kill(sshpid)
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
