package main

import (
	"fmt"
	s "github.com/dghubble/sling"
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
	sling *s.Sling
}

//creates a new HonoClient with base url configured.
func NewHonoRestClient(httpClient *http.Client, baseUrl string) *HonoClient {
	return &HonoClient{
		sling: s.New().Client(httpClient).Base(baseUrl),
	}
}

//registers a new device using the Hono's REST API
func CreateDevice(h HonoClient, tenant string, deviceId int) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s", "registration", tenant)
	deviceBody := fmt.Sprintf("device_id=%d", deviceId)
	fmt.Println("Device registration data : ", deviceBody)
	body := strings.NewReader(deviceBody)
	req, _ := h.sling.New().Post(path).Set("Content-Type", "application/x-www-form-urlencoded").Body(body).Request()

	resp, err := h.sling.Do(req, nil, nil)

	if err != nil {
		fmt.Println("Register device Error: ", err)
	} else {
		fmt.Println("Register device Repsonse: ", resp.Status)
	}

	return resp, err
}

//retrieves the already registered device using Hono's REST API
func GetDevice(h HonoClient, tenant string, deviceId int) (*DEVICE, *http.Response, error) {
	device := new(DEVICE)
	path := fmt.Sprintf("%s/%s/%d", "registration", tenant, deviceId)
	req, _ := h.sling.New().Get(path).Request()
	resp, err := h.sling.Do(req, device, nil)

	if err != nil {
		fmt.Println("GET device Error: ", err.Error())
	} else {
		fmt.Println("GET device Response: ", resp.Status)
		fmt.Printf("Retrieved device: DeviceId[%s] enabled[%t]\n", device.ID, device.DATA.ENABLED)
	}

	return device, resp, err
}

//Sends the given telemetry data to the Hono's REST Adapter
func SendTelemetry(h HonoClient, tenant string, deviceId int, data string) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s/%d", "telemetry", tenant, deviceId)
	fmt.Println("Sending telemetry data : ", data)
	body := strings.NewReader(data)
	req, _ := h.sling.New().Put(path).Set("Content-Type", "application/json").Body(body).Request()

	resp, err := h.sling.Do(req, nil, nil)
	if err != nil {
		fmt.Println("Telemetry Error: ", err)
	} else {
		fmt.Println("Telemetry Response: ", resp.StatusCode)
	}
	return resp, err
}
