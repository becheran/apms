package eap

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/becheran/apms/internal/helper"
)

type EAP struct {
	url    string
	creds  string
	client *http.Client
}

// NewEAP create a new Access Point Object
func NewEAP(ip, user, pw string) *EAP {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err.Error())
	}

	hashedPassword := md5.Sum([]byte(pw))

	eap := EAP{
		url: fmt.Sprintf("https://%s", ip),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar: jar,
		},
		creds: fmt.Sprintf("username=%s&password=%X", user, hashedPassword),
	}

	return &eap
}

func (eap *EAP) post(url string, payload string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(payload)))
	if err != nil {
		panic(err.Error())
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Referer", eap.url)
	return eap.client.Do(req)
}

func (eap *EAP) login() {
	helper.Assert(func() error {
		_, err := eap.post(eap.url, eap.creds)
		return err
	})

	var resp *http.Response
	helper.Assert(func() (err error) {
		loginURL := fmt.Sprintf("%s/data/login.json ", eap.url)
		resp, err = eap.post(loginURL, "operation=read")
		return err
	})
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

type loginData struct {
	// Result of led change
	Enabled *bool `json:"enable"`
	// Result of wifi change (either "on" or "off")
	Status *string `json:"status"`
}

func (ld *loginData) UnmarshalJSON(b []byte) error {
	type loginDataOrig struct {
		Enabled string `json:"enable"`
	}
	var value loginDataOrig
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}
	var result bool
	if value.Enabled == "on" {
		result = true
		ld.Enabled = &result
	} else if value.Enabled == "off" {
		result = false
		ld.Enabled = &result
	}
	return nil
}

type resultPayload struct {
	Error   int       `json:"error"`
	Success bool      `json:"success"`
	Data    loginData `json:"data"`
}

type mode int

const (
	enable mode = iota
	disable
	read
)

func (eap *EAP) postWithLoginRetry(url string, payload string) (body string) {
	req := func() (result string, success bool) {
		resp, err := eap.post(url, payload)
		if err != nil {
			fmt.Printf("Failed to send %s %s. Err: %s", url, payload, err)
			return "", false
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if !strings.Contains(string(b), `"data"`) {
			fmt.Printf("Expected body %s to contain 'data' field.\n", b)
			return "", false
		}
		return string(b), true
	}

	if res, ok := req(); ok {
		return res
	}
	eap.login()
	// Try another x times
	const repeat = 5
	for i := 0; i < repeat; i++ {
		if res, ok := req(); ok {
			return res
		}
		time.Sleep(time.Second / 2)
	}
	fmt.Printf("Unexpected result after %d requests\n", repeat)
	return ""
}

func (eap *EAP) mode(mode mode) (isEnabled bool) {
	if mode == disable || mode == enable {
		var statusMode string
		switch mode {
		case disable:
			statusMode = "off"
		case enable:
			statusMode = "on"
		}
		wifiURL := fmt.Sprintf("%s/data/wireless.basic.json", eap.url)
		// ID 0: 2.4 GHz
		// ID 1: 5.0 GHz
		for radioID := 0; radioID < 2; radioID++ {
			request := func() (success bool) {
				payload := fmt.Sprintf("operation=write&wireless-bset-status=%s&radioID=%d", statusMode, radioID)
				loginRes := eap.postWithLoginRetry(wifiURL, payload)
				var result loginData
				json.Unmarshal([]byte(loginRes), &result)
				if result.Status == nil {
					return false
				}
				return *result.Status == statusMode
			}
			for tries := 0; tries < 5; tries++ {
				fmt.Printf("Send req %d/5\n", tries+1)
				success := request()
				if success {
					break
				}
			}
		}
	}

	ledURL := fmt.Sprintf("%s/data/ledctrl.json", eap.url)
	var payload string
	switch mode {
	case enable:
		payload = "operation=write&enable=on"
	case disable:
		payload = "operation=write&enable=off"
	case read:
		payload = "operation=read"
	}
	body := eap.postWithLoginRetry(ledURL, payload)
	fmt.Println(body)

	var result resultPayload
	json.Unmarshal([]byte(body), &result)

	if result.Data.Enabled == nil {
		fmt.Println("Unexpected result. Maybe need another login.")
	}
	isEnabled = *result.Data.Enabled

	return
}

func (eap *EAP) IsEnabled() (isEnabled bool) {
	return eap.mode(read)
}

func (eap *EAP) Enable() (isEnabled bool) {
	return eap.mode(enable)
}

func (eap *EAP) Disable() (isEnabled bool) {
	return eap.mode(disable)
}
