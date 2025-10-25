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
	"github.com/spf13/cobra"

	"wait4x.dev/v4/checker/temporal"
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
