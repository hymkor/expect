package consoleinput

import (
	"unsafe"

	"github.com/hymkor/expect/internal/go-console"
)

var writeConsoleInput = console.Kernel32.NewProc("WriteConsoleInputW")

func (handle Handle) Write(events []InputRecord) uint32 {
	var count uint32
	writeConsoleInput.Call(uintptr(handle), uintptr(unsafe.Pointer(&events[0])), uintptr(len(events)), uintptr(unsafe.Pointer(&count)))

	return count
}
