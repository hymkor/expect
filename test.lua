do
    echo("*** timeout test ***")
    local backup = timeout
    timeout = 1
    spawn("cmd.exe","/c","timeout 3 >nul & echo hoge")
    local rc = expect("hoge")
    if rc ~= -2 then
        echo("--> [NG]")
        os.exit(1)
    end
    echo("--> [OK]")
    echo("*** non-timeout test ***")
    timeout = backup
    rc = expect("hoge")
    if rc ~= 0 then
        echo("--> [NG]")
        os.exit(1)
    end
    echo("--> [OK]")

end
