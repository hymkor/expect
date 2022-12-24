package main

import (
	"github.com/yuin/gopher-lua"

	"github.com/hymkor/expect/internal/go-console/output"
)

func shot(n int) ([]string, error) {
	if !useStderrOnGetRecentOutput {
		result, err := consoleoutput.GetRecentOutputByStdout(n)
		if err == nil {
			return result, nil
		}
		useStderrOnGetRecentOutput = true
	}
	return consoleoutput.GetRecentOutputByStderr(n)
}

func Shot(L *lua.LState) int {
	n, ok := L.Get(-1).(lua.LNumber)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("Expected a number"))
		return 2
	}
	result, err := shot(int(n))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	table := L.NewTable()
	for i, line := range result {
		L.SetTable(table, lua.LNumber(i+1), lua.LString(line))
	}
	L.Push(table)
	return 1
}
