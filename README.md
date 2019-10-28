# SSGS - Add your ground station to the StellarStation platform

This application provides a simple way to add your ground station to the
StellarStation network without the need to implement all of the APIs yourself.

It includes the following components:
  1. An interface to the StellarStation scheduling APIs
  2. A server for sending satellite data and receiving commands

To build:

``` shell
$ go build cmd/ssgs/ssgs.go
$ go build cmd/sender/sender.go
```

Usage:

``` shell
$ ssgs [-c config.yaml]
```

Example configuration:

``` yaml
station:
  name: "Test"
  id: "1"
  address: "test:443"
  key: "api-key.json"
  plan-update-interval: "5m"
ports:
  - name: "AX.25"
    address: ":5555"
    framing: "AX.25"
  - name: "Bitstream"
    address: ":5556"
    framing: "BITSTREAM"
schedulers:
  - name: "Everything"
    address: ":5554"
```

Notes on the station configuration:
- The `name` value is used for logging purposes only.
- The `id` value should be either the ground station ID as provided by StellarStation.
- The `address` field should be the hostname (or IP address) and port number of the StarPass device.
- The `key` field should be the file name of your API key.
- The `plan-update-interval` field sets the interval for updating plan data.

Notes on the port configuration:
- Each entry represents a data port.
- Data sent to the data port will be forwarded on to StellarStation.
- Commands received from StellarStation will be sent to connected clients.
- The format for the either sending or receiving data is:
  - 4 bytes : Data length (LSB uint32)
  - N bytes : Data
- The `name` value is used for logging purposes only.
- The `address` field tells the server which address and port to listen to.
- The `framing` field tells the server which data framing to use.
  - Valid values are: `AX.25`, `BITSTREAM` (default), `WATERFALL`, and `IQ`
- The `satellite` field can be set to limit the data sent and received to a particular satellite.

Notes on the schedulers configuration:
- Each entry represents a scheduling port.
- Each scheduling port implements a REST API.
- The next 24 hours worth of passes will be returned when `GET /` is called.
- The `name` value is used for logging purposes only.
- The `address` field tells the server which address and port to listen to.
