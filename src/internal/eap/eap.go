package eap

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

type EAP struct {
	url    string
	creds  string
	client *http.Client
}

// NewEAP create a new Access Point Object
func NewEAP(ip, user, pw string) *EAP {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: config}

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err.Error())
	}

	hashedPassword := md5.Sum([]byte(pw))

	eap := EAP{
		url: fmt.Sprintf("https://%s", ip),
		client: &http.Client{
			Transport: tr,
			Jar:       jar,
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
	_, err := eap.post(eap.url, eap.creds)
	if err != nil {
		panic(err.Error())
	}

	loginUrl := fmt.Sprintf("%s/data/login.json ", eap.url)
	resp, err := eap.post(loginUrl, "operation=read")
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

type loginData struct {
	Enabled *bool `json:"enable"`
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

type loginResult struct {
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
	req := func() (string, error) {
		resp, err := eap.post(url, payload)
		if err == nil {
			defer resp.Body.Close()
			b, _ := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(b), `"data"`) {
				return string(b), nil
			}
		}
		return "", fmt.Errorf("Invalid response")
	}

	if res, err := req(); err == nil {
		return res
	}
	eap.login()
	// Try another x times
	for i := 0; i < 5; i++ {
		if res, err := req(); err == nil {
			return res
		}
		time.Sleep(time.Second / 2)
	}
	fmt.Println("Unexpected result after n requests")
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
		wifiUrl := fmt.Sprintf("%s/data/wireless.basic.json", eap.url)
		// ID 0: 2.4 GHz
		// ID 1: 5.0 GHz
		for radioId := 0; radioId < 2; radioId++ {
			payload := fmt.Sprintf("operation=write&wireless-bset-status=%s&radioID=%d", statusMode, radioId)
			eap.postWithLoginRetry(wifiUrl, payload)
		}
	}

	ledUrl := fmt.Sprintf("%s/data/ledctrl.json", eap.url)
	var payload string
	switch mode {
	case enable:
		payload = "operation=write&enable=on"
	case disable:
		payload = "operation=write&enable=off"
	case read:
		payload = "operation=read"
	}
	body := eap.postWithLoginRetry(ledUrl, payload)
	fmt.Println(body)

	var result loginResult
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
