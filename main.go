package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

const (
	escEcho  = "\x1B[40;31;1m"
	escSend  = "\x1B[40;35;1m"
	escSpawn = "\x1B[40;32;1m"
	escEnd   = "\x1B[37;1m"
)

var (
	eOption = flag.String("e", "", "execute string")
	xOption = flag.Bool("x", false, "obsoluted option. Lines startings with '@' are always skipped.")
)

var conIn consoleinput.Handle
var output = colorable.NewColorableStdout()
var echo = ioutil.Discard

// Echo is the implement of the lua-function `echo`
func Echo(L *lua.LState) int {
	value := L.Get(-1)
	if value == lua.LTrue {
		echo = output
	} else if lua.LVIsFalse(value) {
		echo = ioutil.Discard
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
	} else {
		return 0
	}
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

func expect(ctx context.Context, keywords []string, until time.Time) int {
	for time.Now().Before(until) {
		if IsContextCanceled(ctx) {
			return -3
		}
		output, err := consoleoutput.GetRecentOutput()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return -1
		}
		for i, str := range keywords {
			if strings.Index(output, str) >= 0 {
				return i
			}
		}
		time.Sleep(time.Second / time.Duration(10))
	}
	return -2
}

// Expect is the implement of the lua-function `expect`
func Expect(L *lua.LState) int {
	n := L.GetTop()
	keywords := make([]string, n)
	for i := 1; i <= n; i++ {
		keywords[i-1] = L.ToString(i)
	}

	var until time.Time
	if timeout, ok := L.GetGlobal("timeout").(lua.LNumber); ok {
		until = time.Now().Add(time.Duration(timeout) * time.Second)
	} else {
		until = time.Now().Add(time.Hour)
	}
	L.Push(lua.LNumber(expect(L.Context(), keywords, until)))
	return 1
}

var waitGroup sync.WaitGroup

func spawn(args []string) bool {
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
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}
	waitGroup.Add(1)
	go func() {
		cmd.Wait()
		waitGroup.Done()
	}()
	return true
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
	if spawn(args) {
		L.Push(lua.LTrue)
	} else {
		L.Push(lua.LFalse)
	}
	return 1
}

// DoFileExceptForAtmarkLines is the same (*lua.LState)DoFile
// but ignores lines starting with '@'
func DoFileExceptForAtmarkLines(L *lua.LState, fname string) (err error) {
	fd, err := os.Open(fname)
	if err != nil {
		return err
	}

	br := bufio.NewReader(fd)
	in := transform.NewTransformer(func() ([]byte, error) {
		bin, err := br.ReadBytes('\n')
		if err != nil {
			fd.Close()
			return nil, err
		}
		if len(bin) > 0 && bin[0] == '@' {
			rc := make([]byte, 0, len(bin)+2)
			rc = append(rc, '-')
			rc = append(rc, '-')
			rc = append(rc, bin...)
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
	if closer := colorable.EnableColorsStdout(nil) ; closer != nil {
		defer closer()
	}
	flag.Parse()

	if *eOption == "" && len(flag.Args()) < 1 {
		return fmt.Errorf("Usage: %s xxxx.lua", os.Args[0])
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

	table := L.NewTable()
	for i, s := range flag.Args() {
		L.SetTable(table, lua.LNumber(i), lua.LString(s))
	}
	L.SetGlobal("arg", table)

	end, ctx := interruptToCancel(context.Background(), func() {
		if trap, ok := L.GetGlobal("trap").(*lua.LFunction); ok {
			L.Push(trap)
			L.PCall(0, 0, nil)
		}
	})
	defer end()
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
