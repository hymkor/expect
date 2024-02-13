package typekeyas

import (
	"github.com/hymkor/expect/internal/go-console/input"
)

func SendKeyEvent(handle consoleinput.Handle, events ...*consoleinput.KeyEventRecord) uint32 {
	records := make([]consoleinput.InputRecord, len(events))
	for i, e := range events {
		records[i].EventType = consoleinput.KEY_EVENT
		keyEvent := records[i].KeyEvent()
		*keyEvent = *e
		if keyEvent.RepeatCount <= 0 {
			keyEvent.RepeatCount = 1
		}
	}
	return handle.Write(records[:])
}

func Rune(handle consoleinput.Handle, c rune) uint32 {
	return SendKeyEvent(handle,
		&consoleinput.KeyEventRecord{UnicodeChar: uint16(c), KeyDown: 1},
		&consoleinput.KeyEventRecord{UnicodeChar: uint16(c)},
	)
}

func String(handle consoleinput.Handle, s string) {
	for _, c := range s {
		Rune(handle, c)
	}
}
