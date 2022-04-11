package consoleoutput

import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hymkor/expect/internal/go-console"
)

type Coord = console.CoordT
type SmallRect = console.SmallRectT

type CharInfoT struct {
	UnicodeChar uint16
	Attributes  uint16
}

const (
	COMMON_LVB_LEADING_BYTE  = 0x0100
	COMMON_LVB_TRAILING_BYTE = 0x0200
)

var procReadConsoleOutput = console.Kernel32.NewProc("ReadConsoleOutputW")

func readConsoleOutput(buffer []CharInfoT, size windows.Coord, coord windows.Coord, read_region *windows.SmallRect) error {

	sizeValue := *(*uintptr)(unsafe.Pointer(&size))
	coordValue := *(*uintptr)(unsafe.Pointer(&coord))

	status, _, err := procReadConsoleOutput.Call(
		uintptr(windows.Stdout),
		uintptr(unsafe.Pointer(&buffer[0])),
		sizeValue,
		coordValue,
		uintptr(unsafe.Pointer(read_region)))
	if status == 0 {
		return fmt.Errorf("ReadConsoleOutputW: %w", err)
	}
	return nil
}

// for compatible
func ReadConsoleOutput(buffer []CharInfoT, size Coord, coord Coord, read_region *SmallRect) error {
	return readConsoleOutput(
		buffer,
		windows.Coord{int16(size.X()), int16(size.Y())},
		windows.Coord{int16(coord.X()), int16(coord.Y())},
		(*windows.SmallRect)(unsafe.Pointer(read_region)))
}

func GetRecentOutput() (string, error) {
	var screen windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(windows.Stdout, &screen)
	if err != nil {
		return "", fmt.Errorf("GetConsoleScreenBufferInfo: %w", err)
	}

	y := 0
	h := 1
	if screen.CursorPosition.Y >= 1 {
		y = int(screen.CursorPosition.Y - 1)
		h++
	}

	region := &windows.SmallRect{
		Left:   0,
		Top:    int16(y),
		Right:  int16(screen.Size.X),
		Bottom: int16(y + h - 1),
	}

	home := &windows.Coord{}
	charinfo := make([]CharInfoT, int(screen.Size.X)*int(screen.Size.Y))
	err = readConsoleOutput(charinfo, screen.Size, *home, region)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	for i := 0; i < int(screen.Size.Y); i++ {
		for j := 0; j < int(screen.Size.X); j++ {
			p := &charinfo[i*int(screen.Size.X)+j]
			if (p.Attributes & COMMON_LVB_TRAILING_BYTE) != 0 {
				// right side of wide charactor

			} else if (p.Attributes & COMMON_LVB_LEADING_BYTE) != 0 {
				// left side of wide charactor
				if p.UnicodeChar != 0 {
					buffer.WriteRune(rune(p.UnicodeChar))
				}
			} else {
				// narrow charactor
				if p.UnicodeChar != 0 {
					buffer.WriteRune(rune(p.UnicodeChar & 0xFF))
				}
			}
		}
	}
	return strings.TrimSpace(buffer.String()), nil
}
