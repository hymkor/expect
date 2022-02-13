package getch

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/zetamatta/go-console/input"
)

const (
	RIGHT_ALT_PRESSED  = 1
	LEFT_ALT_PRESSED   = 2
	RIGHT_CTRL_PRESSED = 4
	LEFT_CTRL_PRESSED  = 8
	CTRL_PRESSED       = RIGHT_CTRL_PRESSED | LEFT_CTRL_PRESSED
	ALT_PRESSED        = RIGHT_ALT_PRESSED | LEFT_ALT_PRESSED
)

type keyEvent struct {
	Rune  rune
	Scan  uint16
	Shift uint32
}

func (k keyEvent) String() string {
	return fmt.Sprintf("Rune:%v,Scan=%d,Shift=%d", k.Rune, k.Scan, k.Shift)
}

type resizeEvent struct {
	Width  uint
	Height uint
}

func (r resizeEvent) String() string {
	return fmt.Sprintf("Width:%d,Height:%d", r.Width, r.Height)
}

const ( // Button
	FROM_LEFT_1ST_BUTTON_PRESSED = 0x0001
	FROM_LEFT_2ND_BUTTON_PRESSED = 0x0004
	FROM_LEFT_3RD_BUTTON_PRESSED = 0x0008
	FROM_LEFT_4TH_BUTTON_PRESSED = 0x0010
	RIGHTMOST_BUTTON_PRESSED     = 0x0002
)

type Event struct {
	Focus   *struct{} // MS says it should be ignored
	Key     *keyEvent // == KeyDown
	KeyDown *keyEvent
	KeyUp   *keyEvent
	Menu    *struct{}                      // MS says it should be ignored
	Mouse   *consoleinput.MouseEventRecord // not supported,yet
	Resize  *resizeEvent
}

func (e Event) String() string {
	event := make([]string, 0, 7)
	if e.Focus != nil {
		event = append(event, "Focus")
	}
	if e.KeyDown != nil {
		event = append(event, "KeyDown("+e.KeyDown.String()+")")
	}
	if e.KeyUp != nil {
		event = append(event, "KeyUp("+e.KeyUp.String()+")")
	}
	if e.Menu != nil {
		event = append(event, "Menu")
	}
	if e.Mouse != nil {
		event = append(event, "Mouse")
	}
	if e.Resize != nil {
		event = append(event, "Resize("+e.Resize.String()+")")
	}
	if len(event) > 0 {
		return strings.Join(event, ",")
	} else {
		return "no events"
	}
}

type Handle struct {
	consoleinput.Handle
	lastkey         *keyEvent
	eventBuffer     []Event
	eventBufferRead int
}

func New() *Handle {
	return &Handle{Handle: consoleinput.New()}
}

func (h *Handle) Close() {
	h.Handle.Close()
}

func (h *Handle) readEvents(flag uint32) []Event {
	orgConMode := h.GetConsoleMode()
	h.SetConsoleMode(flag)
	defer h.SetConsoleMode(orgConMode)

	result := make([]Event, 0, 2)

	for len(result) <= 0 {
		var events [10]consoleinput.InputRecord
		numberOfEventsRead := h.Read(events[:])

		for i := uint32(0); i < numberOfEventsRead; i++ {
			e := events[i]
			var r Event
			switch e.EventType {
			case consoleinput.FOCUS_EVENT:
				r = Event{Focus: &struct{}{}}
			case consoleinput.KEY_EVENT:
				p := e.KeyEvent()
				k := &keyEvent{
					Rune:  rune(p.UnicodeChar),
					Scan:  p.VirtualKeyCode,
					Shift: p.ControlKeyState,
				}
				if p.KeyDown != 0 {
					r = Event{Key: k, KeyDown: k}
				} else {
					r = Event{KeyUp: k}
				}
			case consoleinput.MENU_EVENT:
				r = Event{Menu: &struct{}{}}
			case consoleinput.MOUSE_EVENT:
				p := e.MouseEvent()
				r = Event{
					Mouse: &consoleinput.MouseEventRecord{
						X:          p.X,
						Y:          p.Y,
						Button:     p.Button,
						ControlKey: p.ControlKey,
						Event:      p.Event,
					},
				}
			case consoleinput.WINDOW_BUFFER_SIZE_EVENT:
				width, height := e.ResizeEvent()
				r = Event{
					Resize: &resizeEvent{
						Width:  uint(width),
						Height: uint(height),
					},
				}
			default:
				continue
			}
			result = append(result, r)
		}
	}
	return result
}

func (h *Handle) bufReadEvent(flag uint32) Event {
	for h.eventBuffer == nil || h.eventBufferRead >= len(h.eventBuffer) {
		h.eventBuffer = h.readEvents(flag)
		h.eventBufferRead = 0
	}
	h.eventBufferRead++
	return h.eventBuffer[h.eventBufferRead-1]
}

// Get a event with concatinating a surrogate-pair of keyevents.
func (h *Handle) getEvent(flag uint32) Event {
	for {
		event1 := h.bufReadEvent(flag)
		if k := event1.Key; k != nil {
			println(k.Rune)
			if h.lastkey != nil {
				k.Rune = utf16.DecodeRune(h.lastkey.Rune, k.Rune)
				h.lastkey = nil
			} else if utf16.IsSurrogate(k.Rune) {
				h.lastkey = k
				continue
			}
		}
		return event1
	}
}

const ALL_EVENTS = consoleinput.ENABLE_WINDOW_INPUT | consoleinput.ENABLE_MOUSE_INPUT

// Get all console-event (keyboard,resize,...)
func (h *Handle) All() Event {
	return h.getEvent(ALL_EVENTS)
}

const IGNORE_RESIZE_EVENT uint32 = 0

// Get character as a Rune
func (h *Handle) Rune() rune {
	for {
		e := h.getEvent(IGNORE_RESIZE_EVENT)
		if e.Key != nil && e.Key.Rune != 0 {
			return e.Key.Rune
		}
	}
}

func (h *Handle) Flush() error {
	org := h.GetConsoleMode()
	h.SetConsoleMode(ALL_EVENTS)
	defer h.SetConsoleMode(org)

	h.eventBuffer = nil
	return h.FlushConsoleInputBuffer()
}

// wait for keyboard event
func (h *Handle) Wait(timeout_msec uintptr) (bool, error) {
	status, err := h.WaitForSingleObject(timeout_msec)
	switch status {
	case consoleinput.WAIT_OBJECT_0:
		return true, nil
	case consoleinput.WAIT_TIMEOUT:
		return false, nil
	case consoleinput.WAIT_ABANDONED:
		return false, errors.New("WAIT_ABANDONED")
	default: // including WAIT_FAILED:
		if err != nil {
			return false, err
		} else {
			return false, errors.New("WAIT_FAILED")
		}
	}
}

func (h *Handle) Within(msec uintptr) (Event, error) {
	orgConMode := h.GetConsoleMode()
	h.SetConsoleMode(ALL_EVENTS)
	defer h.SetConsoleMode(orgConMode)

	if ok, err := h.Wait(msec); err != nil || !ok {
		return Event{}, err
	}
	return h.All(), nil
}

const NUL = '\000'

func (h *Handle) RuneWithin(msec uintptr) (rune, error) {
	orgConMode := h.GetConsoleMode()
	h.SetConsoleMode(IGNORE_RESIZE_EVENT)
	defer h.SetConsoleMode(orgConMode)

	if ok, err := h.Wait(msec); err != nil || !ok {
		return NUL, err
	}
	e := h.getEvent(IGNORE_RESIZE_EVENT)
	if e.Key != nil {
		return e.Key.Rune, nil
	}
	return NUL, nil
}
