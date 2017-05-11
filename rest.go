package main

import (
	"fmt"
	s "github.com/dghubble/sling"
	"net/http"
	"strings"
	"time"
)

type protocol int

const (
	HTTP protocol = 1 + iota
	MQTT
	AMQP
)

//data structure representing a device data as returned by Hono Rest adapter
type data struct {
	enabled bool `json:enabled`
}

//GET device response from Hono Rest adapter
type device struct {
	id   string `json:"id"`
	data `json:"data"`
}

//input data structure used for registering a device.
type devRegParam struct {
	deviceId int `url:"device_id"`
}

//a wrapper struct around the Sling http client
type HonoClient struct {
	sling *s.Sling
}

//creates a new HonoClient with base url configured.
func NewHonoRestClient(httpClient *http.Client, baseUrl string) HonoClient {
	return HonoClient{sling: s.New().Client(httpClient).Base(baseUrl)}
}

//registers a new device using the Hono's REST API
func (h HonoClient) CreateDevice(tenant string, deviceId int, metricsChannel chan int64) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s", "registration", tenant)
	deviceBody := fmt.Sprintf("device_id=%d", deviceId)
	fmt.Printf("[%d] Device registration data : %s\n", deviceId, deviceBody)
	body := strings.NewReader(deviceBody)
	req, _ := h.sling.New().Post(path).Set("Content-Type", "application/x-www-form-urlencoded").Body(body).Request()

	start := time.Now().UnixNano()
	resp, err := h.sling.Do(req, nil, nil)
	end := time.Now().UnixNano()

	metricsChannel <- ((end - start) / 1000000) //in milliseconds

	if err != nil {
		fmt.Printf("[%d] Register device Error: %s\n", deviceId, err.Error())
	} else {
		fmt.Println("[%d] Register device Repsonse: %s\n", deviceId, resp.Status)
	}

	return resp, err
}

//retrieves the already registered device using Hono's REST API
func (h HonoClient) GetDevice(tenant string, deviceId int) (*device, *http.Response, error) {
	dev := new(device)
	path := fmt.Sprintf("%s/%s/%d", "registration", tenant, deviceId)
	req, _ := h.sling.New().Get(path).Request()
	resp, err := h.sling.Do(req, dev, nil)

	if err != nil {
		fmt.Printf("[%d] GET device Error: %s\n", deviceId, err.Error())
	} else {
		fmt.Printf("[%d] GET device Response: %s\n", deviceId, resp.Status)
		fmt.Printf("[%d] Device enabled[%t]\n", deviceId, dev.data.enabled)
	}

	return dev, resp, err
}

//sends the given telemetry data to the Hono's REST Adapter
func (h HonoClient) SendTelemetry(tenant string, deviceId int, data string, metricsChannel chan int64) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s/%d", "telemetry", tenant, deviceId)
	fmt.Printf("[%d] Sending telemetry data : %s\n", deviceId, data)
	body := strings.NewReader(data)
	req, _ := h.sling.New().Put(path).Set("Content-Type", "application/json").Body(body).Request()
	fmt.Println("Sending telemetry request :", req)

	start := time.Now().UnixNano()
	resp, err := h.sling.Do(req, nil, nil)
	end := time.Now().UnixNano()

	metricsChannel <- ((end - start) / 1000000) //in milliseconds

	if err != nil {
		fmt.Printf("[%d] Telemetry Error: %s\n", deviceId, err.Error())
	} else {
		fmt.Printf("[%d] Telemetry Response: %d\n", deviceId, resp)
	}
	return resp, err
}
