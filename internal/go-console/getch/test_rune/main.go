package main

import (
	"fmt"
	"time"

	"github.com/hymkor/expect/internal/go-console/getch"
)

func main() {
	time.Sleep(time.Second / 10)
	getch.Flush()

	ch := getch.Rune()
	fmt.Printf("%08X\n", ch)
}
