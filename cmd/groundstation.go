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

package main

import (
	api "github.com/infostellarinc/go-stellarstation/api/v1/groundstation"

	"github.com/vaelen/ssgs"

	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

/***************
 * Main Method *
 ***************/

func main() {
	var configFile, groundstation string

	flag.StringVar(&configFile, "c", "config.json", "config file")
	flag.Parse()

	config, err := ssgs.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config file: %v\n", err)
	}

	client := ssgs.NewClient()

	client.Connect(config.GroundStation)
	defer client.Stop()

	planWatcher := ssgs.NewPlanWatcher(client)

	dataServers := make([]*ssgs.DataServer, len(config.Data))

	for i, c := range config.Data {
		dataServers[i] = ssgs.NewDataServer(config.GroundStation, c, client)
		defer dataServers[i].Stop()
	}

	planStart := func(plan *api.Plan) {
		for _, s := range dataServers {
			go s.PlanStart(plan)
		}
	}

	planEnd := func(plan *api.Plan) {}

	planWatcher.Start(config.GroundStation.PlanUpdateInterval,
		planStart, planEnd)

	defer planWatcher.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		planWatcher.Stop()
		for _, s := range dataServers {
			s.Stop()
		}
		client.Stop()
	}()

	client.Wait()
}
