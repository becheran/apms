package eap_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/becheran/apms/internal/config"
	"github.com/becheran/apms/internal/eap"
)

func TestConnect(t *testing.T) {
	eap_3 := eap.NewEAP("192.168.0.100", config.User, config.Password)
	eap_3.Disable()
	fmt.Println(eap_3.IsEnabled())
	time.Sleep(time.Second * 5)
	eap_3.Enable()
	fmt.Println(eap_3.IsEnabled())
}
