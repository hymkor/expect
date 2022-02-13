package main

import (
	"fmt"
	"os"

	"github.com/zetamatta/go-console/input"
	"github.com/zetamatta/go-console/typekeyas"
)

func main() {
	console := consoleinput.New()
	for _, s := range os.Args[1:] {
		typekeyas.String(console, s)
		typekeyas.Rune(console, '\r')
	}
	if err := console.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
