package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/mattn/go-colorable"
	"github.com/tidwall/transform"
	"github.com/yuin/gopher-lua"

	"github.com/zetamatta/go-console/input"
)

var (
	escEcho  = "\x1B[40;31;1m"
	escSend  = "\x1B[40;35;1m"
	escSpawn = "\x1B[40;32;1m"
	escEnd   = "\x1B[37;1m"
)

var (
	eOption     = flag.String("e", "", "execute string")
	xOption     = flag.Bool("x", false, "obsoluted option. Lines startings with '@' are always skipped.")
	colorOption = flag.String("color", "always", "colorize the output; can be 'always' (default if omitted), 'auto', or 'never'.")
)

var conIn consoleinput.Handle
var output = colorable.NewColorableStdout()
var echo = io.Discard

// Echo is the implement of the lua-function `echo`
func Echo(L *lua.LState) int {
	value := L.Get(-1)
	if value == lua.LTrue {
		echo = output
	} else if lua.LVIsFalse(value) {
		echo = io.Discard
	} else {
		fmt.Fprintf(output, "%s%s%s\r\n", escEcho, value.String(), escEnd)
	}
	L.Push(lua.LTrue)
	return 1
}

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

// DoFileExceptForAtmarkLines is the same (*lua.LState)DoFile
// but ignores lines starting with '@'
func DoFileExceptForAtmarkLines(L *lua.LState, fname string) (err error) {
	fd, err := os.Open(fname)
	if err != nil {
		return err
	}

	br := bufio.NewReader(fd)
	keepComment := false
	in := transform.NewTransformer(func() ([]byte, error) {
		bin, err := br.ReadBytes('\n')
		if err != nil {
			fd.Close()
			return nil, err
		}
		if keepComment || (len(bin) > 0 && bin[0] == '@') {
			rc := make([]byte, 0, len(bin)+2)
			rc = append(rc, '-')
			rc = append(rc, '-')
			rc = append(rc, bin...)

			trim := bytes.TrimRight(bin, "\r\n")
			keepComment = len(trim) > 0 && bin[len(trim)-1] == '^'
			return rc, nil
		}
		return bin, nil
	})

	f, err := L.Load(in, fname)
	if err != nil {
		return err
	}
	L.Push(f)
	return L.PCall(0, 0, nil)
}

func mains() error {
	if closer := colorable.EnableColorsStdout(nil); closer != nil {
		defer closer()
	}
	flag.Parse()

	if *eOption == "" && len(flag.Args()) < 1 {
		return fmt.Errorf("Usage: %s xxxx.lua", os.Args[0])
	}

	if *colorOption == "never" {
		escEcho = ""
		escSend = ""
		escSpawn = ""
		escEnd = ""
	}

	var err error
	conIn = consoleinput.New()
	defer conIn.Close()

	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("echo", L.NewFunction(Echo))
	L.SetGlobal("send", L.NewFunction(Send))
	L.SetGlobal("sendln", L.NewFunction(Sendln))
	L.SetGlobal("expect", L.NewFunction(Expect))
	L.SetGlobal("spawn", L.NewFunction(Spawn))
	L.SetGlobal("spawnctx", L.NewFunction(SpawnContext))
	L.SetGlobal("kill", L.NewFunction(Kill))

	table := L.NewTable()
	for i, s := range flag.Args() {
		L.SetTable(table, lua.LNumber(i), lua.LString(s))
	}
	L.SetGlobal("arg", table)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	L.SetContext(ctx)

	if *eOption != "" {
		err = L.DoString(*eOption)
	} else {
		err = DoFileExceptForAtmarkLines(L, flag.Arg(0))
	}
	waitGroup.Wait()
	return err
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
