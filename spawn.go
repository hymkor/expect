package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/yuin/gopher-lua"
)

var waitGroup sync.WaitGroup

func spawn(newCmd func(string, ...string) *exec.Cmd,
	args []string, log io.Writer) (int, error) {

	var cmd *exec.Cmd
	for _, s := range args {
		fmt.Fprintf(echo, "%s\"%s\"%s ", escSpawn, s, escEnd)
	}
	fmt.Fprintln(echo)
	if len(args) <= 1 {
		cmd = newCmd(args[0])
	} else {
		cmd = newCmd(args[0], args[1:]...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	waitGroup.Add(1)
	pid := cmd.Process.Pid
	go func() {
		cmd.Wait()
		waitGroup.Done()
	}()
	return pid, nil
}

func _spawn(L *lua.LState, newCmd func(string, ...string) *exec.Cmd) int {
	n := L.GetTop()
	if n < 1 {
		L.Push(lua.LFalse)
		return 1
	}
	args := make([]string, n)
	for i := 0; i < n; i++ {
		args[i] = L.CheckString(1 + i)
	}
	pid, err := spawn(newCmd, args, echo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 0
	}
	L.Push(lua.LNumber(pid))
	return 1
}

// Spawn is the implement of the lua-function `spawn`
func Spawn(L *lua.LState) int {
	return _spawn(L, exec.Command)
}

// SpawnContext is the implement of the lua-function 'spawnctx`
func SpawnContext(L *lua.LState) int {
	return _spawn(L, func(name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(L.Context(), name, arg...)
	})
}
