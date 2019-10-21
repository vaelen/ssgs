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
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

/*****************
 * Configuration *
 *****************/

var DefaultPlanUpdateInterval = time.Minute * 5

// DataServerConfig represents the configuration for a single data server.
type DataServerConfig struct {
	// Name should be a descriptive name.
	// This is mainly used for logging purposes. (required)
	Name string
	// Address is the TCP address and port number to listen to.
	// (required)
	Address string
	// Framing should match the framing type of the data.
	// Valid values are: BITSTREAM, AX.25, WATERFALL, IQ
	// (required)
	Framing string
	// Satellite should match the satellite ID provided
	// by the StellarStation platform. (optional)
	Satellite string
	// Channel should match the communications channel
	// ID provided by the StellarStation platform.
	// (optional)
	Channel string
}

type SchedulerConfig struct {
	// Name should be a descriptive name.
	// This is mainly used for logging purposes. (required)
	Name string
	// Address is the TCP address and port number to listen to.
	// (required)
	Address string
	// Satellite should match the satellite ID provided
	// by the StellarStation platform. (optional)
	Satellite string
	// Channel should match the communications channel
	// ID provided by the StellarStation platform.
	// (optional)
	Channel string
}

// GroundStationConfig defines the configuration for
// the ground station.
type GroundStationConfig struct {
	// Name should be a descriptive name.
	Name string
	// ID should match the ground station ID
	// provided by StellarStation.
	ID string
	// Address is the URL of the ground station.
	Address string
	// Key should be the filename of the API key to use.
	Key string
	// PlanUpdateInterval is the time between plan update checks.
	PlanUpdateInterval time.Duration
}

// Config contains all of the configuration for the application.
type Config struct {
	// GroundStation contains the groundstation configuration.
	GroundStation GroundStationConfig
	// Data contains the data server configurations
	Data []DataServerConfig
	// Schedulers contains the schedular configurations
	Schedulers []SchedulerConfig
}

// LoadConfig loads a configuration file.
func LoadConfig(configFile string) (config *Config, err error) {
	log.Printf("Loading config file: %s\n", configFile)

	var data []byte

	data, err = ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config = &Config{}

	ext := strings.ToLower(path.Ext(configFile))

	switch ext {
	case ".json":
		err = json.Unmarshal(data, config)
	case ".yaml":
		err = yaml.Unmarshal(data, config)
	default:
		log.Fatalf("Unsupported config file type: %s\n", ext)
	}

	if config.GroundStation.PlanUpdateInterval == 0 {
		config.GroundStation.PlanUpdateInterval = DefaultPlanUpdateInterval
	}

	return config, err
}
