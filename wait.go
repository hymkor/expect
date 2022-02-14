package main

import (
	"github.com/yuin/gopher-lua"
	"os"
)

func wait(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	_, err = process.Wait()
	return err
}

// Wait is the implement of the lua-function `wait`
func Wait(L *lua.LState) int {
	pid, ok := L.Get(1).(lua.LNumber)
	if !ok {
		L.Push(lua.LFalse)
		L.Push(lua.LString("wait: argument error"))
		return 2
	}
	err := wait(int(pid))
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}
