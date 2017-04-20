package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

//checks the user provided device id and returns the same. If none,
//generates a new one and returns the same
func GetDeviceId(deviceId int) int {
	if deviceId == 0 {
		return rand.Intn(time.Now().Nanosecond())
	} else {
		return deviceId
	}
}

func GetRandomTemperature() string {
	//random temperature value
	return fmt.Sprintf("{\"temp\": %d}", rand.Int31n(100))
}

func main() {

	//input parameters
	url := flag.String("l", "http://localhost:8080", "hono rest adapter url. Default is http://localhost:8080")
	tenant := flag.String("t", "DEFAULT_TENANT", "default value is DEFAULT_TENANT")
	inputDeviceId := flag.Int("d", 0, "test device id. Leave it to default 0 to generate a random device id")
	register := flag.Bool("r", false, "set to true to register a new device, default is false")
	telemetry := flag.Bool("tm", false, "sends random telemetry data when set to true, default is false")
	noDelay := flag.Bool("c", false, "sends telemetry continously without any delay, default is false")

	flag.Parse()

	httpClient := http.DefaultClient
	honoClient := NewHonoRestClient(httpClient, *url)
	var device = new(DEVICE)
	deviceId := GetDeviceId(*inputDeviceId)

	//Register a new device
	if *register {
		resp, err := CreateDevice(*honoClient, *tenant, deviceId)

		if err == nil && resp.StatusCode != http.StatusNotFound {
			//Get the registered device
			device, _, _ = GetDevice(*honoClient, *tenant, deviceId)
		}
	} else if *inputDeviceId != 0 {
		//Get the registered device
		device, _, _ = GetDevice(*honoClient, *tenant, *inputDeviceId)
	} else {
		fmt.Println("Not a valid Device Id is provided to retrieve")
	}

	if *telemetry && device.DATA.ENABLED {
		//send telemetry data
		resp, err := SendTelemetry(*honoClient, *tenant, deviceId, GetRandomTemperature())
		for err == nil && resp.StatusCode == http.StatusAccepted {
			if *noDelay != true {
				time.Sleep(time.Second * 1)
			}
			resp, err = SendTelemetry(*honoClient, *tenant, deviceId, GetRandomTemperature())
		}

	}
}
