package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	anko_core "github.com/mattn/anko/builtins"
	"github.com/mattn/anko/vm"
	"github.com/mattn/anko/parser"
	"github.com/yuin/gopher-lua"
	"github.com/zetamatta/go-getch/consoleinput"
	"github.com/zetamatta/go-getch/consoleoutput"
	"github.com/zetamatta/go-getch/typekeyas"
)

var conIn consoleinput.Handle

func send(str string) {
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

	if strings.HasSuffix(strings.ToLower(os.Args[1]), ".ank") {
		env := vm.NewEnv()

		anko_core.LoadAllBuiltins(env)

		env.Define("send", func(s string)int{send(s);return 0})
		env.Define("expect", func(s string)int{if expect(s) { return 1; } else{ return 0 }})
		env.Define("spawn", func(ss []string)int{if spawn(ss){ return 1; } else{ return 0 }})
		println("!")

		b, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			return err
		}
		println("!!")
		parser.EnableErrorVerbose()
		_, err = env.Execute(string(b))
		println("!!!")
	} else {
		L := lua.NewState()
		defer L.Close()

		L.SetGlobal("send", L.NewFunction(Send))
		L.SetGlobal("expect", L.NewFunction(Expect))
		L.SetGlobal("spawn", L.NewFunction(Spawn))

		err = L.DoFile(os.Args[1])
	}

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
