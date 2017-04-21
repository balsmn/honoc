package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type TelemetryControl struct {
	telemetry bool //indicates whether telemetry data should be sent
	noDelay   bool //whether telemetry data should be sent without any delay
	protocol  PROTOCOL
}

//checks the user provided device id and returns the same. If none,
//generates a new one and returns the same
func GetDeviceId(deviceId int) int {
	if deviceId == 0 {
		return GetRandomDeviceId()
	} else {
		return deviceId
	}
}

//generates a random device id
func GetRandomDeviceId() int {
	return rand.Intn(time.Now().Nanosecond())
}

//generate random temperature value
func GetRandomTemperature() string {
	return fmt.Sprintf("{\"temp\": %d}", rand.Int31n(100))
}

//registers a device and sends telemetry if registration was successful
func RegisterAndSendTelemetry(honoClient HonoClient, tenant string, deviceId int, tc TelemetryControl) {
	resp, err := CreateDevice(honoClient, tenant, deviceId)

	if err == nil && resp.StatusCode != http.StatusNotFound {
		GetDeviceAndSendTelemetry(honoClient, tenant, deviceId, tc)
	}
}

//gets the device from Hono, if the device is active, sends telemetry data
func GetDeviceAndSendTelemetry(honoClient HonoClient, tenant string, deviceId int, tc TelemetryControl) {
	//get the registered device
	device, _, _ := GetDevice(honoClient, tenant, deviceId)

	if tc.telemetry && device.DATA.ENABLED {
		//send telemetry data
		resp, err := SendTelemetry(honoClient, tenant, deviceId, GetRandomTemperature())
		for err == nil && resp.StatusCode == http.StatusAccepted {
			if tc.noDelay != true {
				time.Sleep(time.Second * 1)
			}
			resp, err = SendTelemetry(honoClient, tenant, deviceId, GetRandomTemperature())
		}
	}
}

func main() {

	//input parameters
	url := flag.String("l", "http://localhost:8080", "hono rest adapter url. Default is http://localhost:8080")
	tenant := flag.String("t", "DEFAULT_TENANT", "default value is DEFAULT_TENANT")
	inputDeviceId := flag.Int("d", 0, "test device id. Leave it to default 0 to generate a random device id")
	register := flag.Bool("r", false, "set to true to register a new device, default is false")
	noOfClients := flag.Int("n", 1, "number of clients. This value takes effect only used in conjunction with -r. When the value is set to more than 1, then the input to -d is simply ignored")
	telemetry := flag.Bool("tm", false, "sends random telemetry data when set to true, default is false")
	noDelay := flag.Bool("c", false, "sends telemetry continously without any delay, default is false")
	p := flag.Int("tp", 1, "protocol for sending telemetry data [1 for http| 2 for mqtt| 3 for amqp]")

	flag.Parse()

	protocol := HTTP
	if *p != 1 {
		fmt.Println("The telemetry protocol is not supported")
		os.Exit(1)
	}

	telemetryControl := &TelemetryControl{telemetry: *telemetry, noDelay: *noDelay, protocol: protocol}

	httpClient := http.DefaultClient
	honoClient := NewHonoRestClient(httpClient, *url)

	deviceId := GetDeviceId(*inputDeviceId)
	//register a new device
	if *register {
		for i := 0; i < *noOfClients; i++ {
			go RegisterAndSendTelemetry(*honoClient, *tenant, deviceId, *telemetryControl)
			deviceId = GetRandomDeviceId()
		}
		var input string
		fmt.Scanln(&input)
	} else if *inputDeviceId != 0 {
		GetDeviceAndSendTelemetry(*honoClient, *tenant, *inputDeviceId, *telemetryControl)
	} else {
		fmt.Println("Not a valid Device Id is provided to retrieve")
	}
}
