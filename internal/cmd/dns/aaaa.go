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
	dns "wait4x.dev/v4/checker/dns/aaaa"
	"wait4x.dev/v4/internal/contextutil"
	"wait4x.dev/v4/waiter"
)

// NewAAAACommand creates a new DNS AAAA command
func NewAAAACommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "AAAA ADDRESS [-- command [args...]]",
		Aliases: []string{"aaaa"},
		Short:   "Check DNS AAAA (IPv6) records for a given domain",
		Long:    "Check for the existence and validity of DNS AAAA (IPv6) records for a specified domain. Supports verification against expected IPv6 addresses and custom nameserver configuration.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the AAAA command")
			}

			return nil
		},
		Example: `
  # Check AAAA records existence
  wait4x dns AAAA wait4x.dev

  # Check AAAA records with expected IPv6 addresses
  wait4x dns AAAA wait4x.dev --expect-ip '2606:4700:3033::ac43:9ab4'

  # Check AAAA records with multiple expected IPv6 addresses
  wait4x dns AAAA wait4x.dev --expect-ip '2606:4700:3033::ac43:9ab4' --expect-ip '2606:4700:3034::ac43:9ab4'

  # Check AAAA records using a specific nameserver
  wait4x dns AAAA wait4x.dev --expect-ip '2606:4700:3033::ac43:9ab4' --nameserver gordon.ns.cloudflare.com

  # Check AAAA records with custom interval and timeout
  wait4x dns AAAA wait4x.dev --interval 5s --timeout 60s`,
		RunE: runAAAA,
	}

	command.Flags().StringArray("expect-ip", nil, "Expect ipv6s.")

	return command
}

// runAAAA is the command handler for the "dns AAAA" command
func runAAAA(cmd *cobra.Command, args []string) error {
	nameserver, err := cmd.Flags().GetString("nameserver")
	if err != nil {
		return fmt.Errorf("failed to parse --nameserver flag: %w", err)
	}

	expectIPs, err := cmd.Flags().GetStringArray("expect-ip")
	if err != nil {
		return fmt.Errorf("failed to parse --expect-ip flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	dc := dns.New(
		args[0],
		dns.WithExpectedIPV6s(expectIPs),
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
