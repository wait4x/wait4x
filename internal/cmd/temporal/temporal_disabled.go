// Copyright 2018-2025 The Wait4X Authors
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

//go:build disable_temporal

// Package temporal provides the Temporal command-line interface for the Wait4X application.
package temporal

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"wait4x.dev/v3/checker"
)

// NewTemporalCommand creates a new temporal sub-command
func NewTemporalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "temporal",
		Short: "Check Temporal connection - this feature is disabled",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("Temporal feature disabled in this build.")
		},
	}
}

// TemporalServiceConfig contains configuration options for temporal checker
type TemporalServiceConfig struct {
	Mode                      string
	Namespace                 string
	TaskQueue                 string
	ExpectWorkerIdentityRegex string
	ConnectionTimeout         *time.Duration
	InsecureTransport         *bool
	InsecureSkipTLSVerify     *bool
}

// CreateTemporalChecker creates a new Temporal checker when feature is disabled
func CreateTemporalChecker(_ string, _ TemporalServiceConfig) (checker.Checker, error) {
	return nil, errors.New("temporal feature disabled in this build")
}
