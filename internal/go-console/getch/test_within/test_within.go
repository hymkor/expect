package main

import (
	"fmt"

	"github.com/zetamatta/go-console/getch"
)

func main() {
	for i := 0; i < 5; i++ {
		e, err := getch.Within(1000)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(e.String())
		}
	}
}
