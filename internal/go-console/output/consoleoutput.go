package consoleoutput

import (
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

func readConsoleOutput(handle windows.Handle, buffer []CharInfoT, size windows.Coord, coord windows.Coord, read_region *windows.SmallRect) error {

	sizeValue := *(*uintptr)(unsafe.Pointer(&size))
	coordValue := *(*uintptr)(unsafe.Pointer(&coord))

	status, _, err := procReadConsoleOutput.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&buffer[0])),
		sizeValue,
		coordValue,
		uintptr(unsafe.Pointer(read_region)))
	if status == 0 {
		return fmt.Errorf("ReadConsoleOutputW: %w", err)
	}
	return nil
}

func GetRecentOutputByHandle(handle windows.Handle, height int) ([]string, error) {
	var screen windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(handle, &screen)
	if err != nil {
		return nil, fmt.Errorf("GetConsoleScreenBufferInfo: %w", err)
	}

	top := int(screen.CursorPosition.Y) - (height - 1)
	if top < 0 {
		top = 0
	}

	region := &windows.SmallRect{
		Left:   0,
		Top:    int16(top),
		Right:  int16(screen.Size.X),
		Bottom: int16(screen.CursorPosition.Y),
	}

	home := &windows.Coord{}
	charinfo := make([]CharInfoT, int(screen.Size.X)*int(screen.Size.Y))
	err = readConsoleOutput(handle, charinfo, screen.Size, *home, region)
	if err != nil {
		return nil, err
	}

	lines := []string{}
	for i := 0; i < int(screen.Size.Y); i++ {
		var buffer strings.Builder
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
		lines = append(lines, strings.TrimSpace(buffer.String()))
	}
	return lines, nil
}

func GetRecentOutputByStdout(height int) ([]string, error) {
	return GetRecentOutputByHandle(windows.Stdout, height)
}

func GetRecentOutputByStderr(height int) ([]string, error) {
	return GetRecentOutputByHandle(windows.Stderr, height)
}
