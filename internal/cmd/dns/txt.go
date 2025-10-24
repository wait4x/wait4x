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

// Package dns provides the DNS command-line interface for the Wait4X application.
package dns

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	dns "wait4x.dev/v4/checker/dns/txt"
	"wait4x.dev/v4/internal/contextutil"
	"wait4x.dev/v4/waiter"
)

// NewTXTCommand creates a new DNS TXT command
func NewTXTCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "TXT ADDRESS [-- command [args...]]",
		Aliases: []string{"txt"},
		Short:   "Check DNS TXT records for a given domain",
		Long:    "Check DNS TXT records for a given domain name and verify TXT records",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the TXT command")
			}

			return nil
		},
		Example: `
  # Check TXT records existence
  wait4x dns TXT wait4x.dev

  # Check TXT records with expected values
  wait4x dns TXT wait4x.dev --expect-value 'include:_spf.mx.cloudflare.net'

  # Check TXT records by defined nameserver
  wait4x dns TXT wait4x.dev --expect-value 'include:_spf.mx.cloudflare.net' --nameserver gordon.ns.cloudflare.com

  # Check TXT records with multiple expected values
  wait4x dns TXT wait4x.dev --expect-value 'v=spf1' --expect-value 'include:_spf.mx.cloudflare.net'

  # Check TXT records with timeout and interval
  wait4x dns TXT wait4x.dev --timeout 60s --interval 5s`,
		RunE: runTXT,
	}

	command.Flags().StringArray("expect-value", nil, "Expected TXT record values")

	return command
}

// runTXT is the command handler for the "dns TXT" command
func runTXT(cmd *cobra.Command, args []string) error {
	nameserver, err := cmd.Flags().GetString("nameserver")
	if err != nil {
		return fmt.Errorf("failed to parse --nameserver flag: %w", err)
	}

	expectValues, err := cmd.Flags().GetStringArray("expect-value")
	if err != nil {
		return fmt.Errorf("failed to parse --expect-value flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("unable to get logger from context: %w", err)
	}

	dc := dns.New(
		args[0],
		dns.WithExpectedValues(expectValues),
		dns.WithNameServer(nameserver),
	)

	return waiter.WaitContext(cmd.Context(),
		dc,
		waiter.WithTimeout(contextutil.GetTimeout(cmd.Context())),
		waiter.WithInterval(contextutil.GetInterval(cmd.Context())),
		waiter.WithInvertCheck(contextutil.GetInvertCheck(cmd.Context())),
		waiter.WithLogger(logger),
	)
}
