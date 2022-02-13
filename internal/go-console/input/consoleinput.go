package consoleinput

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/zetamatta/go-console"
)

type Handle windows.Handle

func New() Handle {
	return Handle(console.In())
}

func (handle Handle) Close() error {
	return nil
}

func (handle Handle) GetConsoleMode() uint32 {
	var mode uint32
	windows.GetConsoleMode(windows.Handle(handle), &mode)
	return mode
}

func (handle Handle) SetConsoleMode(flag uint32) {
	windows.SetConsoleMode(windows.Handle(handle), flag)
}

var flushConsoleInputBuffer = console.Kernel32.NewProc("FlushConsoleInputBuffer")

func (handle Handle) FlushConsoleInputBuffer() error {
	status, _, err := flushConsoleInputBuffer.Call(uintptr(handle))
	if status != 0 {
		return nil
	} else {
		return err
	}
}

var getNumberOfConsoleInputEvents = console.Kernel32.NewProc("GetNumberOfConsoleInputEvents")

func (handle Handle) GetNumberOfEvent() (int, error) {
	var count uint32 = 0
	status, _, err := getNumberOfConsoleInputEvents.Call(uintptr(handle),
		uintptr(unsafe.Pointer(&count)))
	if status != 0 {
		return int(count), nil
	} else {
		return 0, err
	}
}

var waitForSingleObject = console.Kernel32.NewProc("WaitForSingleObject")

func (handle Handle) WaitForSingleObject(msec uintptr) (uintptr, error) {
	status, _, err := waitForSingleObject.Call(uintptr(handle), msec)
	return status, err
}
