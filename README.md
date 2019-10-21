# Groundstation - Add your ground station to the StellarStation platform

This application provides a simple way to add your ground station to the
StellarStation network without the need to implement all of the APIs yourself.

It includes the following components:
  1. An interface to the StellarStation scheduling APIs
  2. A server for sending satellite data and receiving commands

To build:

``` shell
$ go build cmd/groundstation.go
```

Usage:

``` shell
$ groundstation [-c config.yaml]
```

Example configuration:

``` yaml
GroundStation:
  Name: "Test"
  ID: "1"
  Address: "test:443"
  Key: "api-key.json"
  PlanUpdateInterval: "5m"
Data:
  - Name: "AX.25"
    Address: ":5555"
    Framing: "AX.25"
  - Name: "Bitstream"
    Address: ":5556"
    Framing: "BITSTREAM"
Schedulers:
  - Name: "Everything"
    Address: ":5554"
```

Notes on the GroundStation configuration:
- The `Name` value is used for logging purposes only.
- The `ID` value should be either the ground station ID as provided by StellarStation.
- The `Address` field should be the hostname (or IP address) and port number of the StarPass device.
- The `Key` field should be the file name of your API key.
- The `PlanUpdateInterval` fields sets the interval for updating plan data.

Notes on the Data configuration:
- Each entry represents a data port.
- Data sent to the data port will be forwarded on to StellarStation.
- Commands received from StellarStation will be sent to connected clients.
- The format for the either sending or receiving data is:
  - 4 bytes : Data length (LSB)
  - N bytes : Data
- The `Name` value is used for logging purposes only.
- The `Address` field tells the server which address and port to listen to.
- The `Framing` field tells the server which data framing to use.
  - Valid values are: `AX.25`, `BITSTREAM` (default), `WATERFALL`, and `IQ`
- The `Satellite` field can be set to limit the data sent and received to a particular satellite.

Notes on the Schedulers configuration:
- Each entry represents a scheduling port.
- New plans will be send to connected clients.
- NOTE: This feature is not yet implemented.
- The `Name` value is used for logging purposes only.
- The `Address` field tells the server which address and port to listen to.
