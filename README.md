# honoc 
honoc is a simple command line interface (CLI) written in Go for creating a device and sending telemetry data to an Eclipse Hono instance.

The simple super fast CLI can be used for testing and perform stress/load test your Eclipse Hono installation.

honoc simplifies the registration process and can be used to send telemetry data using different protocols like HTTP or MQTT. Current implementation uses the Hono's REST interface to register a new device and supports only HTTP for telemetry data. MQTT is planned to be available soon.

### Features

* Can talk to any Eclipse Hono instance reachable in internet
* Create a device in Hono: Registers a new device with a randomly generated device id and verifies the availability of such a device in Hono.
* Sends telemetry data (randomly generated temperature value) using the selected protocol

## Install

    go get github.com/balsmn/honoc
    go install github.com/balsmn/honoc

## Usage

	honoc -?

## Reference

For more details about Eclipse Hono read [here](https://github.com/eclipse/hono)