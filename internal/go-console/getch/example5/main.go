package main

import (
	"fmt"

	"github.com/zetamatta/go-console/getch"
)

func main() {
	getch.Flush()

	// get console event
	fmt.Println("Type ESC-Key to program shutdown.")
	for {
		e := getch.All()
		if k := e.Key; k != nil {
			fmt.Printf("key down: code=%04X scan=%04X shift=%04X\n",
				k.Rune, k.Scan, k.Shift)
		}
		if k := e.KeyUp; k != nil {
			fmt.Printf("key up: code=%04X scan=%04X shift=%04X\n",
				k.Rune, k.Scan, k.Shift)
			if k.Rune == 0x1B {
				break
			}
		}
	}
}
