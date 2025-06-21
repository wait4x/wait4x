// Copyright 2019-2025 The Wait4X Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package influxdb provides the InfluxDB checker for the Wait4X application.
package influxdb

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"wait4x.dev/v3/checker"
)

// InfluxDB is an InfluxDB checker
type InfluxDB struct {
	serverURL string
}

// New creates a new InfluxDB checker
func New(serverURL string) checker.Checker {
	i := &InfluxDB{
		serverURL: serverURL,
	}

	return i
}

// Identity returns the identity of the InfluxDB checker
func (i *InfluxDB) Identity() (string, error) {
	return i.serverURL, nil
}

// Check checks the InfluxDB connection
func (i *InfluxDB) Check(ctx context.Context) error {
	// InfluxDB doesn't validate authentication params on Ping and Health requests.
	ic := influxdb2.NewClient(i.serverURL, "")
	defer ic.Close()

	res, err := ic.Ping(ctx)
	if !res {
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the influxdb server", err,
				"address", i.serverURL,
			)
		}

		return err
	}

	return nil
}
