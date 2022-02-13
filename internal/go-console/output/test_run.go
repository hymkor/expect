// +build run

package main

import (
	"fmt"
	"os"

	"github.com/zetamatta/go-console/output"
)

func Main() error {
	for _, arg1 := range os.Args {
		fmt.Println(arg1)
		output, err := consoleoutput.GetRecentOutput()
		if err != nil {
			return err
		}
		fmt.Printf("-->[%s]\n", output)
		fmt.Print(arg1)
		output, err = consoleoutput.GetRecentOutput()
		if err != nil {
			return err
		}
		fmt.Printf("\n-->[%s]\n", output)
	}
	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
