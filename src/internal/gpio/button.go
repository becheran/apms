package gpio

import (
	"time"

	"github.com/stianeikeland/go-rpio"
)

// WaitForButton wait infinetly long for button rise
func WaitForButton(pinID int, toggleCallback func()) {
	pin := rpio.Pin(pinID)
	pin.Input()
	pin.PullUp()
	// Edge Detection does not work on Pi 1
	//buttonPin.Detect(rpio.FallEdge)
	lastState := pin.Read()
	ticker := time.NewTicker(time.Millisecond * 100) // TODO google time?
	for {
		<-ticker.C
		state := pin.Read()
		if state < lastState {
			toggleCallback()
		}
		lastState = state

	}
}
