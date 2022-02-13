package getch

var hconin *Handle

func lazyinit() {
	if hconin != nil {
		return
	}
	hconin = New()
}

// Get all console-event (keyboard,resize,...)
func All() Event {
	lazyinit()
	return hconin.All()
}

// Get character as a Rune
func Rune() rune {
	lazyinit()
	return hconin.Rune()
}

func Count() (int, error) {
	lazyinit()
	return hconin.GetNumberOfEvent()
}

func Flush() error {
	lazyinit()
	return hconin.Flush()
}

// wait for keyboard event
func Wait(timeout_msec uintptr) (bool, error) {
	lazyinit()
	return hconin.Wait(timeout_msec)
}

func Within(msec uintptr) (Event, error) {
	lazyinit()
	return hconin.Within(msec)
}

func RuneWithin(msec uintptr) (rune, error) {
	lazyinit()
	return hconin.RuneWithin(msec)
}

func IsCtrlCPressed() bool {
	lazyinit()
	return hconin.IsCtrlCPressed()
}

func DisableCtrlC() {
	lazyinit()
	hconin.DisableCtrlC()
}
