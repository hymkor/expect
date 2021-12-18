package main

import (
	"fmt"
	"os"

	"github.com/yuin/gopher-lua"
)

func kill(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

// Kill is the implement of the lua-function `kill`
func Kill(L *lua.LState) int {
	pid, ok := L.Get(1).(lua.LNumber)
	if !ok {
		fmt.Fprintln(os.Stderr, "wait: not process-id")
		return 0
	}
	err := kill(int(pid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wait: %s\n", err.Error())
	}
	return 0
}
