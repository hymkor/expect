package main

import (
	"context"
	"os"
	"os/signal"
)

func interruptToCancel(ctx context.Context) (context.Context, func()) {
	newctx, cancel := context.WithCancel(ctx)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		select {
		case <-sigch:
			cancel()
		case <-newctx.Done():
		}
	}()
	return newctx, func() {
		cancel()
		signal.Stop(sigch)
		close(sigch)
	}
}
