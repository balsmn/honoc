package main

import (
	"fmt"
	. "github.com/dghubble/sling"
	"math/rand"
	"net/http"
	"strings"
)

//data structure representing a device data as returned by Hono Rest adapter
type DATA struct {
	ENABLED bool `json:enabled`
}

//GET device response from Hono Rest adapter
type DEVICE struct {
	ID   string `json:"id"`
	DATA `json:"data"`
}

//input data structure used for registering a device.
type DEVICE_REG_PARAMS struct {
	DEVICE_ID int `url:"device_id"`
}

//A wrapper struct around the Sling http client
type HonoClient struct {
	sling *Sling
}

func NewHonoRestClient(httpClient *http.Client, baseUrl string) *HonoClient {
	return &HonoClient{
		sling: New().Client(httpClient).Base(baseUrl),
	}
}

func CreateDevice(h *HonoClient, tenant string, deviceId int) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s", "registration", tenant)
	deviceBody := fmt.Sprintf("device_id=%d", deviceId)
	fmt.Println("Device registration data : ", deviceBody)
	body := strings.NewReader(deviceBody)
	req, _ := h.sling.Post(path).Set("Content-Type", "application/x-www-form-urlencoded").Body(body).Request()

	resp, err := h.sling.Do(req, nil, nil)

	return resp, err
}

func GetDevice(h *HonoClient, tenant string, deviceId int) (*DEVICE, *http.Response, error) {
	device := new(DEVICE)
	path := fmt.Sprintf("%s/%s/%d", "registration", tenant, deviceId)
	req, _ := h.sling.Get(path).Request()
	resp, err := h.sling.Do(req, device, nil)

	if err != nil {
		fmt.Println("GET Error: ", err.Error())
	} else {
		fmt.Println("GET Response: ", resp.Status)
		fmt.Printf("Received device: DeviceId[%s] enabled[%t]\n", device.ID, device.DATA.ENABLED)
	}

	return device, resp, err
}

func SendTelemetry(h *HonoClient, tenant string, deviceId int) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s/%d", "telemetry", tenant, deviceId)
	//random temperature value
	telemetryData := fmt.Sprintf("{\"temp\": %d}", rand.Int31n(100))
	fmt.Println("Sending telemetry data : ", telemetryData)
	body := strings.NewReader(telemetryData)
	req, _ := h.sling.Put(path).Set("Content-Type", "application/json").Body(body).Request()

	resp, err := h.sling.Do(req, nil, nil)
	if err != nil {
		fmt.Println("Telemetry Error: ", err)
	} else {
		fmt.Println("Telemetry Response: ", resp.StatusCode)
	}
	return resp, err
}
