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

//go:build !disable_temporal

// Package temporal provides the Temporal command-line interface for the Wait4X application.
package temporal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/temporal"
)

// NewTemporalCommand creates a new temporal sub-command
func NewTemporalCommand() *cobra.Command {
	temporalCommand := &cobra.Command{
		Use:   "temporal",
		Short: "Check Temporal server & worker",
	}

	temporalCommand.PersistentFlags().Duration("connection-timeout", temporal.DefaultConnectionTimeout, "Timeout is the maximum amount of time a dial will wait for a GRPC connection to complete.")
	temporalCommand.PersistentFlags().Bool("insecure-transport", temporal.DefaultInsecureTransport, "Skips GRPC transport security.")
	temporalCommand.PersistentFlags().Bool("insecure-skip-tls-verify", temporal.DefaultInsecureSkipTLSVerify, "Skips tls certificate checks for the GRPC request.")

	temporalCommand.AddCommand(NewServerCommand())
	temporalCommand.AddCommand(NewWorkerCommand())

	return temporalCommand
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

// CreateTemporalChecker creates a new Temporal checker from configuration
func CreateTemporalChecker(target string, sc TemporalServiceConfig) (checker.Checker, error) {
	var mode temporal.CheckMode = temporal.CheckModeServer
	if sc.Mode != "" {
		switch strings.ToLower(sc.Mode) {
		case "server":
			mode = temporal.CheckModeServer
		case "worker":
			mode = temporal.CheckModeWorker
		default:
			return nil, fmt.Errorf("invalid temporal mode: %s (must be server or worker)", sc.Mode)
		}
	}

	var opts []temporal.Option
	if sc.ConnectionTimeout != nil {
		opts = append(opts, temporal.WithTimeout(*sc.ConnectionTimeout))
	}
	if sc.InsecureTransport != nil {
		opts = append(opts, temporal.WithInsecureTransport(*sc.InsecureTransport))
	}
	if sc.InsecureSkipTLSVerify != nil {
		opts = append(opts, temporal.WithInsecureSkipTLSVerify(*sc.InsecureSkipTLSVerify))
	}

	if mode == temporal.CheckModeWorker {
		if sc.Namespace == "" {
			return nil, errors.New("temporal worker mode requires namespace")
		}
		if sc.TaskQueue == "" {
			return nil, errors.New("temporal worker mode requires task-queue")
		}
		opts = append(opts, temporal.WithNamespace(sc.Namespace))
		opts = append(opts, temporal.WithTaskQueue(sc.TaskQueue))
		if sc.ExpectWorkerIdentityRegex != "" {
			opts = append(opts, temporal.WithExpectWorkerIdentityRegex(sc.ExpectWorkerIdentityRegex))
		}
	}

	return temporal.New(mode, target, opts...), nil
}
