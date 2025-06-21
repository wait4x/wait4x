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
	dns "wait4x.dev/v3/checker/dns/ns"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewNSCommand creates a new DNS NS command
func NewNSCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "NS ADDRESS [-- command [args...]]",
		Aliases: []string{"ns"},
		Short:   "Check DNS NS records for a given domain",
		Long:    "Check DNS NS records for a given domain name and verify nameserver records",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the NS command")
			}

			return nil
		},
		Example: `
  # Check NS records existence
  wait4x dns NS wait4x.dev

  # Check NS records with expected nameservers
  wait4x dns NS wait4x.dev --expect-nameserver 'emma.ns.cloudflare.com'

  # Check NS records with multiple expected nameservers
  wait4x dns NS wait4x.dev --expect-nameserver 'emma.ns.cloudflare.com' --expect-nameserver 'gordon.ns.cloudflare.com'

  # Check NS records using a specific nameserver
  wait4x dns NS wait4x.dev --nameserver '8.8.8.8:53'

  # Check NS records with timeout and interval
  wait4x dns NS wait4x.dev --timeout 60s --interval 5s

  # Invert the check (wait until NS records don't match)
  wait4x dns NS wait4x.dev --expect-nameserver 'emma.ns.cloudflare.com' --invert-check`,
		RunE: runNS,
	}

	command.Flags().StringArray("expect-nameserver", nil, "Expect nameservers.")

	return command
}

// runNS is the command handler for the "dns NS" command
func runNS(cmd *cobra.Command, args []string) error {
	nameserver, err := cmd.Flags().GetString("nameserver")
	if err != nil {
		return fmt.Errorf("failed to parse --nameserver flag: %w", err)
	}

	expectNameservers, err := cmd.Flags().GetStringArray("expect-nameserver")
	if err != nil {
		return fmt.Errorf("failed to parse --expect-nameserver flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	dc := dns.New(
		args[0],
		dns.WithExpectedNameservers(expectNameservers),
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
