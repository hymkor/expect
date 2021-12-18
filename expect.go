package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yuin/gopher-lua"

	"github.com/zetamatta/go-console/output"
)

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
