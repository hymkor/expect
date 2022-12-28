do
    echo("*** timeout test ***")
    do
        local backup = timeout
        timeout = 1
        spawn("cmd.exe","/c","timeout 3 >nul & echo hoge")
        local rc = expect("hoge")
        if rc ~= -2 then
            echo("--> [NG]")
            os.exit(1)
        end
        echo("--> [OK]")
    end

    echo("*** non-timeout test ***")
    do
        timeout = backup
        local rc = expect("hoge")
        if rc ~= 0 then
            echo("--> [NG]")
            os.exit(1)
        end
        echo("--> [OK]")
    end

    function testShot(word)
        local screen = shot(2)
        for i=1,#screen do
            if string.find(screen[i],word,1,true) then
                return true
            end
        end
        return false
    end

    echo("*** shot test ***")
    do
        local screen = shot(2)
        if not testShot("shot") then
            echo("--> [NG]")
            os.exit(1)
        end
        echo("--> [OK]")
    end

    echo("*** sleep test ***")
    do
        spawn("cmd.exe","/c","timeout 1 >nul & echo sleeptest")
        if testShot("sleeptest") then
            echo("--> [NG 1st/2]")
            os.exit(1)
        end
        echo("--> [OK 1st/2]")
        sleep(1)
        if not testShot("sleeptest") then
            echo("--> [NG 2nd/2]")
            os.exit(1)
        end
        echo("--> [OK 2nd/2]")
    end

    echo("*** usleep test ***")
    do
        spawn("cmd.exe","/c","timeout 1 >nul & echo usleeptest")
        if testShot("usleeptest") then
            echo("--> [NG 1st/2]")
            os.exit(1)
        end
        echo("--> [OK 1st/2]")
        usleep(1000000)
        if not testShot("usleeptest") then
            echo("--> [NG 2nd/2]")
            os.exit(1)
        end
        echo("--> [OK 2nd/2]")
    end
end
