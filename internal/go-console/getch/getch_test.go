package getch

import (
	"fmt"
	"testing"
	"time"
)

func reportEvent(e Event) {
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

func TestAll(t *testing.T) {
	if err := Flush(); err != nil {
		t.Error(err.Error())
		return
	}
	for i := 0; i < 3; i++ {
		fmt.Printf("[%d/3] ", i+1)
		reportEvent(All())
	}
}

func TestCount(t *testing.T) {
	var err error
	if err = Flush(); err != nil {
		t.Error(err.Error())
		return
	}
	var n int
	for {
		n, err = Count()
		if err != nil {
			t.Error(err.Error())
			return
		}
		if n > 0 {
			break
		}
		fmt.Println("sleep")
		time.Sleep(time.Second)
	}
	fmt.Printf("break(n=%d)\n", n)
	reportEvent(All())
}

func TestWait(t *testing.T) {
	var err error
	if err = Flush(); err != nil {
		t.Error(err.Error())
		return
	}
	for {
		hit, hit_err := Wait(1000)
		if hit_err != nil {
			t.Error(hit_err.Error())
			return
		}
		if hit {
			break
		}
		fmt.Println("Nothing")
	}
	reportEvent(All())
}
