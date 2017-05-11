package honoc

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type telemetryControl struct {
	telemetry bool //indicates whether telemetry data should be sent
	noDelay   bool //whether telemetry data should be sent without any delay
	protocol  PROTOCOL
}

//checks the user provided device id and returns the same. If none,
//generates a new one and returns the same
func getDeviceId(deviceId int) int {
	if deviceId == 0 {
		return getRandomDeviceId()
	} else {
		return deviceId
	}
}

//generates a random device id
func getRandomDeviceId() int {
	return rand.Intn(time.Now().Nanosecond())
}

//generate random temperature value
func getRandomTemperature() string {
	return fmt.Sprintf("{\"temp\": %d}", rand.Int31n(100))
}

//registers a device and sends telemetry if registration was successful
func registerAndSendTelemetryViaRest(h HonoClient, tenant string, deviceId int, tc telemetryControl, registrationMetrics chan int64, telemetryMetrics chan int64) {
	resp, err := h.CreateDevice(tenant, deviceId, registrationMetrics)

	if err == nil && resp.StatusCode != http.StatusNotFound {
		sendTelemetryViaRest(h, tenant, deviceId, tc, telemetryMetrics)
	}
}

//gets the device from Hono, if the device is active, sends telemetry data
func sendTelemetryViaRest(h HonoClient, tenant string, deviceId int, tc telemetryControl, telemetryMetrics chan int64) {
	//get the registered device
	device, _, _ := h.GetDevice(tenant, deviceId)

	if tc.telemetry && device.DATA.ENABLED {
		//send telemetry data
		resp, err := h.SendTelemetry(tenant, deviceId, getRandomTemperature(), telemetryMetrics)
		//break out of loop if server doesn't accept any data
		for err == nil && resp.StatusCode == http.StatusAccepted {
			if tc.noDelay != true {
				time.Sleep(time.Second * 1)
			}
			resp, err = h.SendTelemetry(tenant, deviceId, getRandomTemperature(), telemetryMetrics)
		}
	}
}

func printMetrics(metricName string, metricsChannel chan int64) {
	fmt.Printf("%s - %d ms\n", metricName, <-metricsChannel)
}

func awaitTermination() {
	//infinite loop to prevent program from exiting
	var input string
	fmt.Scanln(&input)
}

func main() {

	//input parameters
	registerCommand := flag.NewFlagSet("register", flag.ExitOnError)
	noOfClients := registerCommand.Int("n", 1, "number of clients. This value takes effect only used in conjunction with -r. When the value is set to more than 1, then the input to -d is simply ignored")
	noDelay := registerCommand.Bool("c", false, "sends telemetry continously without any delay, default is false")
	r_tenant := registerCommand.String("t", "DEFAULT_TENANT", "default value is DEFAULT_TENANT")
	r_url := registerCommand.String("l", "http://localhost:8080", "hono rest adapter url. Default is http://localhost:8080")
	r_telemetry := registerCommand.Bool("tm", false, "sends random telemetry data when set to true, default is false")
	r_p := registerCommand.Int("tp", 1, "protocol for sending telemetry data [1 for http| 2 for mqtt| 3 for amqp]")
	r_inputDeviceId := registerCommand.Int("d", 0, "device id of the first client. Leave it, to generate a random device id")

	telemetryCommand := flag.NewFlagSet("telemetry", flag.ExitOnError)
	inputDeviceId := telemetryCommand.Int("d", 0, "test device id. Leave it to default 0 to generate a random device id")
	t_tenant := telemetryCommand.String("t", "DEFAULT_TENANT", "default value is DEFAULT_TENANT")
	t_url := telemetryCommand.String("l", "http://localhost:8080", "hono rest adapter url. Default is http://localhost:8080")
	t_p := telemetryCommand.Int("tp", 1, "protocol for sending telemetry data [1 for http| 2 for mqtt| 3 for amqp]")

	flag.Parse()

	register := false
	telemetry := false

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "register":
		registerCommand.Parse(os.Args[2:])
		register = true
		break
	case "telemetry":
		telemetryCommand.Parse(os.Args[2:])
		telemetry = true
		break
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		flag.PrintDefaults()
		os.Exit(2)
	}

	protocol := HTTP
	if *r_p != 1 || *t_p != 1 {
		fmt.Println("The telemetry protocol is not supported")
		os.Exit(1)
	}

	registrationMetrics := make(chan int64) //contains metrics for sending device registrations to Hono.
	telemetryMetrics := make(chan int64)    //contains metrics for sending telemetry to Hono.

	if register {

		tc := telemetryControl{telemetry: *r_telemetry, noDelay: *noDelay, protocol: protocol}

		httpClient := http.DefaultClient
		honoClient := NewHonoRestClient(httpClient, *r_url)

		deviceId := getDeviceId(*r_inputDeviceId)
		//register a new device
		for i := 0; i < *noOfClients; i++ {
			go registerAndSendTelemetryViaRest(honoClient, *r_tenant, deviceId, tc, registrationMetrics, telemetryMetrics)
			go printMetrics("registration", registrationMetrics)
			go printMetrics("telemetry", telemetryMetrics)
			deviceId = getRandomDeviceId()
		}
	} else if telemetry && *inputDeviceId != 0 {
		tc := telemetryControl{telemetry: true, noDelay: false, protocol: protocol}

		httpClient := http.DefaultClient
		honoClient := NewHonoRestClient(httpClient, *t_url)

		go sendTelemetryViaRest(honoClient, *t_tenant, *inputDeviceId, tc, telemetryMetrics)
		go printMetrics("telemetry", telemetryMetrics)

	} else {
		fmt.Println("Not a valid Device Id is provided to retrieve")
	}
	awaitTermination()
}
