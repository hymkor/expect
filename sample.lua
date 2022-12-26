if #arg < 1 then
    print("expect.exe sample.lua USERNAME@DOMAIN PASSWD")
    os.exit(0)
end
local account = arg[1]
local password = arg[2]
local sshexe = os.getenv("windir") .. "\\System32\\OpenSSH\\ssh.exe"

spawn(sshexe,"-p","22",account)
timeout = 10
capturelines = 3 -- default is 2

while true do
    local rc = expect(
    "password:",
    "Are you sure you want to continue connecting (yes/no/[fingerprint])?",
    "Could not resolve hostname")
    if rc == 0 then
        sendln(password)
        rc = expect("~]$")
        if rc == 0 then
            sendln("exit")
        end
        break
    elseif rc == 1 then
        sendln("yes")
        print() -- move cursor down not to capture same keyword
    else
        if _MATCH then
            echo(string.format("Error keyword found \"%s\". Exit",_MATCH))
        else
            echo("TIMEOUT")
        end
        break
    end
end
