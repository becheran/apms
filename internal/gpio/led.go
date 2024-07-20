package gpio

import (
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio"
)

type Led struct {
	pin rpio.Pin
}

func NewLed(pinID int) *Led {
	pin := rpio.Pin(pinID)
	pin.Output()
	return &Led{
		pin: pin,
	}
}

func (led *Led) On() {
	fmt.Println("LED on")
	led.pin.High()
}

func (led *Led) Off() {
	fmt.Println("LED off")
	led.pin.Low()
}

func (led *Led) Blink(cancel <-chan bool) {
	fmt.Println("LED blink")
	ticker := time.NewTicker(time.Second / 4)
	led.pin.Toggle()
	go func() {
		for {
			select {
			case <-ticker.C:
				led.pin.Toggle()
			case <-cancel:
				fmt.Println("LED blink terminated")
				return
			}
		}
	}()
}
