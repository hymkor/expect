if #arg < 1 then
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
