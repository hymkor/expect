package getch

import (
	"os"
	"os/signal"
)

func (h *Handle) ctrlCHandler(ch chan os.Signal) {
	for _ = range ch {
		event1 := Event{Key: &keyEvent{3, 0, LEFT_CTRL_PRESSED}}
		if h.eventBuffer == nil {
			h.eventBuffer = []Event{event1}
			h.eventBufferRead = 0
		} else {
			h.eventBuffer = append(h.eventBuffer, event1)
		}
	}
}

func (h *Handle) IsCtrlCPressed() bool {
	if h.eventBuffer != nil {
		for _, p := range h.eventBuffer[h.eventBufferRead:] {
			if p.Key != nil && p.Key.Rune == rune(3) {
				return true
			}
		}
	}
	return false
}

func (h *Handle) DisableCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go h.ctrlCHandler(ch)
}
