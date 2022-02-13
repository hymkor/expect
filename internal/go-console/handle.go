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

func (c CoordT) X() int         { return int(c.x) }
func (c CoordT) Y() int         { return int(c.y) }
func (c CoordT) XY() (int, int) { return int(c.x), int(c.y) }

// SmallRectT exists for compatible. You should use windows.SmallRect
type SmallRectT struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

func LeftTopRightBottom(L, T, R, B int) *SmallRectT {
	return &SmallRectT{
		left:   int16(L),
		top:    int16(T),
		right:  int16(R),
		bottom: int16(B),
	}
}
func (s SmallRectT) Left() int   { return int(s.left) }
func (s SmallRectT) Top() int    { return int(s.top) }
func (s SmallRectT) Right() int  { return int(s.right) }
func (s SmallRectT) Bottom() int { return int(s.bottom) }

// Handle is the alias of windows.Handle
type Handle = windows.Handle

var Kernel32 = windows.NewLazyDLL("kernel32")

var out Handle
var outOnce sync.Once

// ConOut returns the handle for Console-Output
func Out() Handle {
	return windows.Stdout
}

var in Handle
var inOnce sync.Once

func In() Handle {
	return windows.Stdin
}
