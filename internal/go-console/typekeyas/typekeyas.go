package typekeyas

import (
	"github.com/zetamatta/go-console/input"
)

func Rune(handle consoleinput.Handle, c rune) uint32 {
	records := []consoleinput.InputRecord{
		consoleinput.InputRecord{EventType: consoleinput.KEY_EVENT},
		consoleinput.InputRecord{EventType: consoleinput.KEY_EVENT},
	}
	keydown := records[0].KeyEvent()
	keydown.KeyDown = 1
	keydown.RepeatCount = 1
	keydown.UnicodeChar = uint16(c)

	keyup := records[1].KeyEvent()
	keyup.RepeatCount = 1
	keyup.UnicodeChar = uint16(c)

	return handle.Write(records[:])
}

func String(handle consoleinput.Handle, s string) {
	for _, c := range s {
		Rune(handle, c)
	}
}
