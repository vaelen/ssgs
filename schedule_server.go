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
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
)

/******************
 * ScheduleServer *
 ******************/

// ScheduleServer provides a RESTful API for scheduling information
type ScheduleServer struct {
	gsConfig GroundStationConfig
	config   SchedulerConfig
	client   *Client
	router   *mux.Router
	server   *http.Server
}

// NewScheduleServer creates a new ScheduleServer instance
func NewScheduleServer(gsConfig GroundStationConfig,
	config SchedulerConfig,
	client *Client) *ScheduleServer {

	server := &http.Server{
		Addr: config.Address,
	}

	scheduleServer := &ScheduleServer{
		gsConfig: gsConfig,
		config:   config,
		client:   client,
		server:   server,
	}

	scheduleServer.configureRouter()

	go func() {
		log.Printf("(%v) Starting scheduler. Port: %v\n",
			config.Name, config.Address)
		defer log.Printf("(%v) Scheduler stopped. Port: %v\n",
			config.Name, config.Address)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("(%v) Couldn't start schedule server. Error: %v\n",
				config.Name, err)
		}
	}()

	return scheduleServer
}

func (s *ScheduleServer) configureRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.home)
	s.server.Handler = r
}

func (s *ScheduleServer) Stop() {
	s.server.Close()
}

func (s *ScheduleServer) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		now := time.Now()
		start := now.Add(-time.Hour)
		end := now.Add(24 * time.Hour)

		plans, err := s.client.ListPlans(start, end)
		if err != nil {
			log.Printf("!!!!! (%v) Failed to list plans: %v\n",
				s.config.Name, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": %q}", err)))
			return
		}

		m := jsonpb.Marshaler{
			EmitDefaults: true,
			Indent:       "  ",
		}

		buffer := &bytes.Buffer{}

		for _, plan := range plans {
			err = m.Marshal(buffer, plan)
			if err != nil {
				log.Printf("!!!!! (%v) Failed to convert plans to JSON: %v\n",
					s.config.Name, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\": %q}", err)))
				return
			}
			buffer.Write([]byte(",\n"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"plans": [`))
		w.Write([]byte{'\n'})
		w.Write(buffer.Bytes())
		w.Write([]byte(`]}`))
		w.Write([]byte{'\n'})
	case "POST":
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "post called"}`))
	case "PUT":
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message": "put called"}`))
	case "DELETE":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "delete called"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
	}
}
