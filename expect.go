package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/yuin/gopher-lua"
	"github.com/zetamatta/go-console/input"
	"github.com/zetamatta/go-console/output"
	"github.com/zetamatta/go-console/typekeyas"
	"github.com/zetamatta/go-texts"
	"github.com/zetamatta/go-texts/filter"
)

var eOption = flag.String("e", "", "execute string")
var xOption = flag.Bool("x", false, "obsoluted option. Lines startings with '@' are always skipped.")

var conIn consoleinput.Handle

var output = colorable.NewColorableStdout()
var echo = ioutil.Discard

const escEcho = "\x1B[40;31;1m"
const escSend = "\x1B[40;35;1m"
const escSpawn = "\x1B[40;32;1m"
const escEnd = "\x1B[37;1m"

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

// Send is the implement of the lua-function `send`
func Send(L *lua.LState) int {
	wait := 0
	if val, ok := L.Get(2).(lua.LNumber); ok {
		wait = int(val)
	}
	send(L.ToString(1), wait)
	L.Push(lua.LTrue)
	return 1
}

func expect(keywords []string) int {
	for {
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
}

// Expect is the implement of the lua-function `expect`
func Expect(L *lua.LState) int {
	n := L.GetTop()
	keywords := make([]string, n)
	for i := 1; i <= n; i++ {
		keywords[i-1] = L.ToString(i)
	}
	L.Push(lua.LNumber(expect(keywords)))
	return 1
}

var waitProcess = []*exec.Cmd{}

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
	waitProcess = append(waitProcess, cmd)
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
func DoFileExceptForAtmarkLines(L *lua.LState, fname string) error {
	fd, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer fd.Close()

	reader := filter.New(fd, func(line []byte) ([]byte, error) {
		line = bytes.Replace(line, texts.ByteOrderMark, []byte{}, -1)
		if len(line) > 0 && line[0] == '@' {
			line = []byte{'\n'}
		}
		return line, nil
	})

	f, err := L.Load(reader, fname)
	if err != nil {
		return err
	}
	L.Push(f)
	return L.PCall(0, 0, nil)
}

func main1() error {
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
	L.SetGlobal("expect", L.NewFunction(Expect))
	L.SetGlobal("spawn", L.NewFunction(Spawn))

	table := L.NewTable()
	for i, s := range flag.Args() {
		L.SetTable(table, lua.LNumber(i), lua.LString(s))
	}
	L.SetGlobal("arg", table)

	if *eOption != "" {
		err = L.DoString(*eOption)
	} else {
		err = DoFileExceptForAtmarkLines(L, flag.Arg(0))
	}
	for _, c := range waitProcess {
		c.Wait()
	}
	return err
}

func main() {
	if err := main1(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
