package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/yuin/gopher-lua"
	"github.com/zetamatta/go-getch/consoleinput"
	"github.com/zetamatta/go-getch/consoleoutput"
	"github.com/zetamatta/go-getch/typekeyas"
)

var conIn consoleinput.Handle

var echoOn = false

func Echo(L *lua.LState) int {
	value := L.Get(-1)
	if value == lua.LTrue {
		echoOn = true
	} else if lua.LVIsFalse(value) {
		echoOn = false
	} else {
		fmt.Println(value.String())
	}
	L.Push(lua.LTrue)
	return 1
}

func send(str string) {
	if echoOn {
		str_ := strings.Replace(str, "\r", "\n", -1)
		fmt.Print(str_)
	}
	typekeyas.String(conIn, str)
}

func Send(L *lua.LState) int {
	send(L.ToString(1))
	L.Push(lua.LTrue)
	return 1
}

var conOut consoleoutput.Handle

func expect(str string) bool {
	for {
		output, err := conOut.GetRecentOutput()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return false
		}
		if strings.Index(output, str) >= 0 {
			return true
		}
		time.Sleep(time.Second / time.Duration(10))
	}
}

func Expect(L *lua.LState) int {
	if expect(L.ToString(1)) {
		L.Push(lua.LTrue)
	} else {
		L.Push(lua.LFalse)
	}
	return 1
}

var waitProcess = []*exec.Cmd{}

func spawn(args []string) bool {
	var cmd *exec.Cmd
	if echoOn {
		for _, s := range args {
			fmt.Printf("\"%s\" ", s)
		}
		fmt.Println()
	}
	if len(args) <= 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	} else {
		waitProcess = append(waitProcess, cmd)
		return true
	}
}

func Spawn(L *lua.LState) int {
	n := L.GetTop()
	if n < 1 {
		L.Push(lua.LFalse)
		return 1
	}
	args := make([]string, n)
	for i := 0; i < n; i++ {
		args[i] = L.CheckString(1 + i)
	}
	if spawn(args) {
		L.Push(lua.LTrue)
	} else {
		L.Push(lua.LFalse)
	}
	return 1
}

func Main() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("Usage: %s xxxx.lua", os.Args[0])
	}
	var err error
	conIn, err = consoleinput.New()
	if err != nil {
		return err
	}
	defer conIn.Close()

	conOut, err = consoleoutput.New()
	if err != nil {
		return err
	}
	defer conOut.Close()

	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("echo", L.NewFunction(Echo))
	L.SetGlobal("send", L.NewFunction(Send))
	L.SetGlobal("expect", L.NewFunction(Expect))
	L.SetGlobal("spawn", L.NewFunction(Spawn))

	err = L.DoFile(os.Args[1])

	for _, c := range waitProcess {
		c.Wait()
	}
	return err
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
