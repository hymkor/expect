package console

import (
	"sync"

	"golang.org/x/sys/windows"
)

// CoordT exists for compatible. You should use windows.Coord
type CoordT struct {
	x int16
	y int16
}

// X is getter of X-coordination.
func (c CoordT) X() int { return int(c.x) }

// Y is getter of Y-coordination.
func (c CoordT) Y() int { return int(c.y) }

// XY is getter of all field of the coordination.
func (c CoordT) XY() (int, int) { return int(c.x), int(c.y) }

// SmallRectT exists for compatible. You should use windows.SmallRect
type SmallRectT struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

// LeftTopRightBottom is the constructor for SmallRectT
func LeftTopRightBottom(L, T, R, B int) *SmallRectT {
	return &SmallRectT{
		left:   int16(L),
		top:    int16(T),
		right:  int16(R),
		bottom: int16(B),
	}
}

// Left is the getter of `left` field.
func (s SmallRectT) Left() int { return int(s.left) }

// Top is the getter of `top` field.
func (s SmallRectT) Top() int { return int(s.top) }

// Right is the getter of `right` field.
func (s SmallRectT) Right() int { return int(s.right) }

// Bottom is the getter of `bottom` field.
func (s SmallRectT) Bottom() int { return int(s.bottom) }

// Handle is the alias of windows.Handle
type Handle = windows.Handle

// Kernel32 is the instance of kernel32.dll
var Kernel32 = windows.NewLazyDLL("kernel32")

var out Handle
var outOnce sync.Once

// Out returns the handle for Console-Output
func Out() Handle {
	return windows.Stdout
}

var in Handle
var inOnce sync.Once

// In returns the handle for Console-Input
func In() Handle {
	return windows.Stdin
}
