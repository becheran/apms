package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/becheran/apms/internal/config"
	"github.com/becheran/apms/internal/eap"
	"github.com/becheran/apms/internal/gpio"
	"github.com/becheran/apms/internal/timehelper"
	"github.com/stianeikeland/go-rpio"
)

func Manage(ip string, ledPin int, buttonPin int) {
	led := gpio.NewLed(ledPin)

	eap3 := eap.NewEAP(ip, config.User, config.Password)
	isEnabledEap3 := eap3.IsEnabled()
	if isEnabledEap3 {
		led.On()
	} else {
		led.Off()
	}

	go gpio.WaitForButton(buttonPin, func() {
		fmt.Println("Button pressed")
		cancel := make(chan bool)
		led.Blink(cancel)
		if isEnabledEap3 {
			isEnabledEap3 = eap3.Disable()
		} else {
			isEnabledEap3 = eap3.Enable()
		}
		cancel <- true
		if isEnabledEap3 {
			led.On()
		} else {
			led.Off()
		}
	})
}

func main() {
	fmt.Println("Eap Manager")
	timehelper.RetryTillNill(rpio.Open)
	defer rpio.Close()

	// Pinout https://de.pinout.xyz/pinout/pin15_gpio22
	go Manage("192.168.0.100", 4, 17)
	go Manage("192.168.0.101", 2, 3)
	go Manage("192.168.0.102", 18, 23)
	go Manage("192.168.0.103", 14, 15)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Terminated")
}
