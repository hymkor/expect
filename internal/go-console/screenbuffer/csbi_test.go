package csbi

import (
	"fmt"
	"os"
	"testing"
)

func TestGetLocate(t *testing.T) {
	for i := 0; i < 40; i++ {
		fmt.Println()
	}
	os.Stdout.Sync()
	x, y := GetConsoleScreenBufferInfo().CursorPosition.XY()
	fmt.Print(" ")
	x1 := GetConsoleScreenBufferInfo().CursorPosition.X()
	if x1 != x+1 {
		t.Fatal("GetLocate(x) failed")
		return
	}
	fmt.Print("\n")
	os.Stdout.Sync()
	x1, y1 := GetConsoleScreenBufferInfo().CursorPosition.XY()
	if x1 != 0 || (y1 != y && y1 != y+1) {
		t.Fatal("GetLocate(y) failed")
		return
	}
}
