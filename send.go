package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/yuin/gopher-lua"

	"github.com/zetamatta/go-console/typekeyas"
)

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
