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

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker/temporal"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewServerCommand creates a new server sub-command
func NewServerCommand() *cobra.Command {
	serverCommand := &cobra.Command{
		Use:   "server TARGET [flags] [-- command [args...]]",
		Short: "Check Temporal server health check",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("TARGET is required argument for the server command")
			}

			return nil
		},
		Example: `
  # Checking just Temporal server health check
  wait4x temporal server 127.0.0.1:7233

  # Checking insecure Temporal server (no TLS)
  wait4x temporal server 127.0.0.1:7233 --insecure-transport
`,
		RunE: runServer,
	}

	return serverCommand
}

// runServer runs the server command
func runServer(cmd *cobra.Command, args []string) error {
	conTimeout, err := cmd.Flags().GetDuration("connection-timeout")
	if err != nil {
		return fmt.Errorf("failed to parse connection-timeout flag: %w", err)
	}

	insecureTransport, err := cmd.Flags().GetBool("insecure-transport")
	if err != nil {
		return fmt.Errorf("failed to parse insecure-transport flag: %w", err)
	}

	insecureSkipTLSVerify, err := cmd.Flags().GetBool("insecure-skip-tls-verify")
	if err != nil {
		return fmt.Errorf("failed to parse insecure-skip-tls-verify flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	tc := temporal.New(
		temporal.CheckModeServer,
		args[0],
		temporal.WithTimeout(conTimeout),
		temporal.WithInsecureTransport(insecureTransport),
		temporal.WithInsecureSkipTLSVerify(insecureSkipTLSVerify),
	)

	return waiter.WaitContext(
		cmd.Context(),
		tc,
		waiter.WithTimeout(contextutil.GetTimeout(cmd.Context())),
		waiter.WithInterval(contextutil.GetInterval(cmd.Context())),
		waiter.WithInvertCheck(contextutil.GetInvertCheck(cmd.Context())),
		waiter.WithLogger(logger),
	)
}
