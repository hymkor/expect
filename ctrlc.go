package main

import (
	"context"
	"os"
	"os/signal"
)

func IsContextCanceled(ctx context.Context) bool {
	done := ctx.Done()
	if done != nil {
		select {
		case <-done:
			return true
		default:
		}
	}
	return false
}

func interruptToCancel(ctx context.Context, on func()) (func(), context.Context) {
	newctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		select {
		case <-sigChan:
			if on != nil {
				on()
			}
			cancel()
		}
	}()
	return func() {
		signal.Stop(sigChan)
		close(sigChan)
		cancel()
	}, newctx
}
