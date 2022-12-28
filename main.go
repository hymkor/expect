package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/yuin/gopher-lua"

	"github.com/hymkor/expect/internal/filter"
	"github.com/hymkor/expect/internal/go-console/input"
)

var (
	escEcho  = "\x1B[49;31;1m"
	escSend  = "\x1B[49;35;1m"
	escSpawn = "\x1B[49;32;1m"
	escEnd   = "\x1B[49;39;1m"
)

var (
	eOption      = flag.String("e", "", "execute string")
	xOption      = flag.Bool("x", false, "obsoluted option. Lines startings with '@' are always skipped.")
	colorOption  = flag.String("color", "always", "colorize the output; can be 'always' (default if omitted), 'auto', or 'never'.")
	noLogoOption = flag.Bool("nologo", false, "do not show logo")
)

var conIn consoleinput.Handle
var output = colorable.NewColorableStdout()
var echo = io.Discard

func Sleep(L *lua.LState) int {
	value, ok := L.Get(-1).(lua.LNumber)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("Expect a number as the first argument"))
		return 2
	}
	time.Sleep(time.Second * time.Duration(value))
	L.Push(lua.LTrue)
	return 1
}

func USleep(L *lua.LState) int {
	value, ok := L.Get(-1).(lua.LNumber)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("Expect a number as the first argument"))
		return 2
	}
	time.Sleep(time.Microsecond * time.Duration(value))
	L.Push(lua.LTrue)
	return 1
}

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

// DoFileExceptForAtmarkLines is the same (*lua.LState)DoFile
// but ignores lines starting with '@'
func DoFileExceptForAtmarkLines(L *lua.LState, fname string) (err error) {
	fd, err := os.Open(fname)
	if err != nil {
		return err
	}

	br := bufio.NewReader(fd)
	keepComment := false
	in := &filter.Reader{
		In: func() ([]byte, error) {
			bin, err := br.ReadBytes('\n')
			if err != nil {
				fd.Close()
				if err != io.EOF {
					return nil, err
				}
			}
			if keepComment || (len(bin) > 0 && bin[0] == '@') {
				rc := make([]byte, 0, len(bin)+2)
				rc = append(rc, '-')
				rc = append(rc, '-')
				rc = append(rc, bin...)

				trim := bytes.TrimRight(bin, "\r\n")
				keepComment = len(trim) > 0 && bin[len(trim)-1] == '^'
				return rc, err
			}
			return bin, err
		},
	}

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

	if !*noLogoOption {
		fmt.Fprintf(os.Stderr, "Expect-lua %s-windows-%s\n", version, runtime.GOARCH)
	}

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
	L.SetGlobal("wait", L.NewFunction(Wait))
	L.SetGlobal("shot", L.NewFunction(Shot))
	L.SetGlobal("sleep", L.NewFunction(Sleep))
	L.SetGlobal("usleep", L.NewFunction(USleep))

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

var version = "snapshot"

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
