package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/tidwall/transform"
	"github.com/yuin/gopher-lua"

	"github.com/zetamatta/go-console/input"
	"github.com/zetamatta/go-console/output"
	"github.com/zetamatta/go-console/typekeyas"
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
		fmt.Fprintf(output, "%s%s%s\n", escEcho, value.String(), escEnd)
	}
	L.Push(lua.LTrue)
	return 1
}

func send(str string, wait int) {
	fmt.Fprintf(echo, "%s%s%s", escSend, strings.Replace(str, "\r", "\n", -1), escEnd)
	for _, c := range str {
		typekeyas.Rune(conIn, c)
		if wait > 0 {
			time.Sleep(time.Second * time.Duration(wait) / 1000)
		}
	}
}

func getWaitFrom2ndArg(L *lua.LState) int {
	if val, ok := L.Get(2).(lua.LNumber); ok {
		return int(val)
	}
	return 0
}

// Send is the implement of the lua-function `send`
func Send(L *lua.LState) int {
	send(L.ToString(1), getWaitFrom2ndArg(L))
	L.Push(lua.LTrue)
	return 1
}

// Sendln sends 1st arguments and CR
func Sendln(L *lua.LState) int {
	send(L.ToString(1)+"\r", getWaitFrom2ndArg(L))
	L.Push(lua.LTrue)
	return 1
}

func expect(ctx context.Context, keywords []string, timeover time.Duration) (int, error) {
	tick := time.NewTicker(time.Second / 10)
	defer tick.Stop()
	timer := time.NewTimer(timeover)
	defer timer.Stop()
	for {
		output, err := consoleoutput.GetRecentOutput()
		if err != nil {
			return -1, err
		}
		for i, str := range keywords {
			if strings.Index(output, str) >= 0 {
				return i, nil
			}
		}
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		case <-timer.C:
			return -1, context.DeadlineExceeded
		case <-tick.C:
		}
	}
}

const (
	errnoExpectGetRecentOutput = -1
	errnoExpectTimeOut         = -2
	errnoExpectContextDone     = -3
)

// Expect is the implement of the lua-function `expect`
func Expect(L *lua.LState) int {
	n := L.GetTop()
	keywords := make([]string, n)
	for i := 1; i <= n; i++ {
		keywords[i-1] = L.ToString(i)
	}

	timeout := time.Hour
	if n, ok := L.GetGlobal("timeout").(lua.LNumber); ok {
		timeout = time.Duration(n) * time.Second
	}
	rc, err := expect(L.Context(), keywords, timeout)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			rc = errnoExpectContextDone
		} else if errors.Is(err, context.DeadlineExceeded) {
			rc = errnoExpectTimeOut
		} else {
			rc = errnoExpectGetRecentOutput
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
	L.Push(lua.LNumber(rc))
	return 1
}

var waitGroup sync.WaitGroup

func spawn(args []string, log io.Writer) (int, error) {
	var cmd *exec.Cmd
	for _, s := range args {
		fmt.Fprintf(echo, "%s\"%s\"%s ", escSpawn, s, escEnd)
	}
	fmt.Fprintln(echo)
	if len(args) <= 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
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

// Spawn is the implement of the lua-function `spawn`
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
	pid, err := spawn(args, echo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 0
	}
	L.Push(lua.LNumber(pid))
	return 1
}

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
