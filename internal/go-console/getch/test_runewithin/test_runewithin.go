package main

import (
	"fmt"

	"github.com/hymkor/expect/internal/go-console/getch"
)

func main() {
	for i := 0; i < 5; i++ {
		r, err := getch.RuneWithin(1000)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Printf("typed = '%v'\n", r)
		}
	}
}
