package main

import (
	"fmt"
	"time"

	"github.com/hymkor/expect/internal/go-console/getch"
)

func main() {
	time.Sleep(time.Second / 10)
	getch.Flush()

	// wait keyboard-event (timeout: 10-seconds)
	ok, err := getch.Wait(10000)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if !ok {
		fmt.Println("Time-out")
		return
	}

	// get console event
	e := getch.All()
	if k := e.Key; k != nil {
		fmt.Printf("key down: code=%04X scan=%04X shift=%04X\n",
			k.Rune, k.Scan, k.Shift)
	}
	if k := e.KeyUp; k != nil {
		fmt.Printf("key up: code=%04X scan=%04X shift=%04X\n",
			k.Rune, k.Scan, k.Shift)
	}
	if r := e.Resize; r != nil {
		fmt.Printf("window resize: width=%d height=%d\n",
			r.Width, r.Height)
	}
	if e.Mouse != nil {
		fmt.Println("mouse event")
	}
	if e.Menu != nil {
		fmt.Println("menu event")
	}
	if e.Focus != nil {
		fmt.Println("focus event")
	}
}
