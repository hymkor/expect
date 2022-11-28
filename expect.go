package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yuin/gopher-lua"

	"github.com/hymkor/expect/internal/go-console/output"
)

var useStderrOnGetRecentOutput = false

func getRecentOutputByStdoutOrStderr() ([]string, error) {
	for {
		if useStderrOnGetRecentOutput {
			result, err := consoleoutput.GetRecentOutputByStderr(2)
			return result, err
		}
		result, err := consoleoutput.GetRecentOutputByStdout(2)
		if err == nil {
			return result, nil
		}
		useStderrOnGetRecentOutput = true
	}
}

type Matching struct {
	Position  int
	Line      string
	Match     string
	PreMatch  string
	PostMatch string
}

func expect(ctx context.Context, keywords []string, timeover time.Duration) (int, *Matching, error) {
	tick := time.NewTicker(time.Second / 10)
	defer tick.Stop()
	timer := time.NewTimer(timeover)
	defer timer.Stop()
	for {
		outputs, err := getRecentOutputByStdoutOrStderr()
		if err != nil {
			return -1, nil, fmt.Errorf("expect: %w", err)
		}
		for _, output := range outputs {
			for i, str := range keywords {
				if pos := strings.Index(output, str); pos >= 0 {
					return i, &Matching{
						Position:  pos,
						Line:      output,
						Match:     output[pos : pos+len(str)],
						PreMatch:  output[:pos],
						PostMatch: output[pos+len(str):],
					}, nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return -1, nil, fmt.Errorf("expect: %w", ctx.Err())
		case <-timer.C:
			return -1, nil, fmt.Errorf("expect: %w", context.DeadlineExceeded)
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
	rc, info, err := expect(L.Context(), keywords, timeout)
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
	if info != nil {
		L.SetGlobal("_MATCHPOSITION", lua.LNumber(info.Position))
		L.SetGlobal("_MATCHLINE", lua.LString(info.Line))
		L.SetGlobal("_MATCH", lua.LString(info.Match))
		L.SetGlobal("_PREMATCH", lua.LString(info.PreMatch))
		L.SetGlobal("_POSTMATCH", lua.LString(info.PostMatch))
	} else {
		L.SetGlobal("_MATCHPOSITION", lua.LNil)
		L.SetGlobal("_MATCHLINE", lua.LNil)
		L.SetGlobal("_MATCH", lua.LNil)
		L.SetGlobal("_PREMATCH", lua.LNil)
		L.SetGlobal("_POSTMATCH", lua.LNil)
	}
	L.Push(lua.LNumber(rc))
	return 1
}
