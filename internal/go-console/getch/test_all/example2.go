package main

import (
	"fmt"

	"github.com/hymkor/expect/internal/go-console/getch"
)

const COUNT = 5

func main() {
	getch.Flush()
	for i := 0; i < COUNT; i++ {
		fmt.Printf("[%d/%d] ", i+1, COUNT)
		e := getch.All()
		fmt.Println(e.String())
	}
}
