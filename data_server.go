/*
 * Copyright 2019 Andrew C. Young
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ssgs

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/infostellarinc/go-stellarstation/api/v1"
	api "github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
)

/*********
 * DataServer *
 *********/

var ByteOrder = binary.LittleEndian

// DataServer does all of the actual work.
type DataServer struct {
	gsConfig GroundStationConfig
	config   DataServerConfig
	send     <-chan []byte
	recv     chan<- []byte
	client   *Client
	framing  v1.Framing
	ctx      context.Context
	cancel   func()
}

// NewDataServer creates a new DataServer instance
func NewDataServer(gsConfig GroundStationConfig,
	config DataServerConfig,
	client *Client) *DataServer {

	var framing v1.Framing

	switch strings.ToUpper(config.Framing) {
	case "AX.25":
		framing = v1.Framing_AX25
	case "IQ":
		framing = v1.Framing_IQ
	case "WATERFALL":
		framing = v1.Framing_WATERFALL
	default:
		framing = v1.Framing_BITSTREAM
	}

	ctx, cancel := context.WithCancel(context.Background())

	dataServer := &DataServer{
		gsConfig: gsConfig,
		config:   config,
		client:   client,
		framing:  framing,
		send:     make(<-chan []byte),
		recv:     make(chan<- []byte),
		ctx:      ctx,
		cancel:   cancel,
	}

	go dataServer.listen()

	return dataServer
}

func (s *DataServer) listen() {
	addr, err := net.ResolveTCPAddr("tcp", s.config.Address)
	if err != nil {
		log.Fatalf("!!!!! (%v) Couldn't parse TCP address: %v, error: %v\n",
			s.config.Name, s.config.Address, err)
	}

	listener, err := net.ListenTCP("tcp", addr)

	if err != nil {
		log.Fatalf("!!!!! (%v) Couldn't start TCP listener: %v\n", s.config.Name, err)
	}
	defer listener.Close()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			listener.SetDeadline(time.Now().Add(time.Millisecond * 100))
			conn, err := listener.AcceptTCP()
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			}
			if err != nil {
				log.Fatalf("!!!!! (%v) Error listening for connection. Local Address: %v, Error: %v\n",
					s.config.Name, s.config.Address, err)
			}
			if conn != nil {
				go s.serve(conn)
			}
		}
	}

}

func logNetworkError(name string, err error, conn *net.TCPConn) bool {
	if err != nil {
		if err != io.EOF {
			log.Printf("!!!!! (%v) Connection error.  Address: %v, Error: %v\n",
				name, conn.RemoteAddr(), err)
		}
		return true
	}
	return false
}

func (s *DataServer) serve(conn *net.TCPConn) {
	log.Printf(".:.:. (%v) Connection Opened.  Address:%v\n",
		s.config.Name, conn.RemoteAddr())
	defer func() {
		log.Printf(":.:.: (%v) Connection Closed.  Address:%v\n",
			s.config.Name, conn.RemoteAddr())
	}()

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	go func() {
		defer conn.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-s.send:
				var l uint32
				l = uint32(len(data))

				err := binary.Write(conn, ByteOrder, l)
				if logNetworkError(s.config.Name, err, conn) {
					return
				}

				err = binary.Write(conn, ByteOrder, data)
				if logNetworkError(s.config.Name, err, conn) {
					return
				}
			}
		}
	}()

	in := bufio.NewReader(conn)

	for {
		var l uint32
		err := binary.Read(in, ByteOrder, &l)
		if logNetworkError(s.config.Name, err, conn) {
			return
		}

		data := make([]byte, l)
		err = binary.Read(in, ByteOrder, data)
		if logNetworkError(s.config.Name, err, conn) {
			return
		}

		s.recv <- data
	}
}

// Stop will stop the dataServer
func (s *DataServer) Stop() {
	s.cancel()
}

// Wait will wait for the dataServer to stop before returning
func (s *DataServer) Wait() {
	<-s.ctx.Done()
}

// PlanStart is executed when a plan starts
func (s *DataServer) PlanStart(plan *api.Plan) {
	_, aos, los, _, err := parseTimestamps(plan)
	if err != nil {
		log.Printf("!!!!! Couldn't parse plan timestamps. Plan ID: %v, Error: %v\n", plan.PlanId, err)
		return
	}

	log.Printf(">>>>> Plan started. %v\n", shortPlanData(plan))
	defer func() {
		log.Printf("<<<<< Plan ended. %v\n", shortPlanData(plan))
	}()

	// Wait for AOS to start streaming
	time.Sleep(time.Until(aos))

	log.Printf("***** Plan AOS reached. %v\n", shortPlanData(plan))

	losTimer := time.NewTimer(time.Until(los))
	defer losTimer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := s.client.OpenGroundStationStream(ctx)
	if err != nil {
		log.Printf("!!!!! Couldn't open stream. Plan ID: %v, Error: %v\n", plan.PlanId, err)
		return
	}

	log.Printf("||||| Stream opened.  Plan ID: %v\n", plan.PlanId)

	for {
		select {
		case <-losTimer.C:
			log.Printf("..... Plan LOS reached. %v\n", shortPlanData(plan))
			return
		case data := <-s.send:
			request := s.TelemetryRequest(plan, s.framing, data)
			err = stream.Send(request)
			if err != nil {
				log.Printf("!!!!! Couldn't send telemetry request. Plan ID: %v, Error: %v\n", plan.PlanId, err)
			}
			log.Printf("^^^^^ Sent %v bytes. Plan ID: %v\n", len(data), plan.PlanId)
		}
	}
}

// TelemetryRequest creates a telemetry request objet for the given plan and data.
func (s *DataServer) TelemetryRequest(plan *api.Plan, framing v1.Framing, data []byte) *api.GroundStationStreamRequest {
	now := ptypes.TimestampNow()
	freq := plan.DownlinkRadioDevice.CenterFrequencyHz

	telemetry := &api.SatelliteTelemetry{
		PlanId: plan.PlanId,
		Telemetry: &v1.Telemetry{
			Data:                  data,
			Framing:               framing,
			TimeFirstByteReceived: now,
			TimeLastByteReceived:  now,
			DownlinkFrequencyHz:   freq,
		},
	}
	return s.client.TelemetryRequest(telemetry)
}
