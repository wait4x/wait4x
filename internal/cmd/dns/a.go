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
	"wait4x.dev/v3/internal/contextutil"

	dns "wait4x.dev/v3/checker/dns/a"
	"wait4x.dev/v3/waiter"
)

// NewACommand creates a new DNS A command
func NewACommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "A ADDRESS [-- command [args...]]",
		Aliases: []string{"a"},
		Short:   "Check DNS A records for a given domain",
		Long:    "Check DNS A records to verify domain name resolution to IPv4 addresses. Supports checking against expected IPs and custom nameservers.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the A command")
			}

			return nil
		},
		Example: `
  # Check A records existence for a domain
  wait4x dns A wait4x.dev

  # Check A records with specific expected IPv4 addresses
  wait4x dns A wait4x.dev --expect-ip 172.67.154.180
  wait4x dns A wait4x.dev --expect-ip 172.67.154.180 --expect-ip 104.21.60.85

  # Check A records using a custom nameserver
  wait4x dns A wait4x.dev --expect-ip 172.67.154.180 -n gordon.ns.cloudflare.com

  # Check A records with timeout and interval settings
  wait4x dns A wait4x.dev --timeout 30s --interval 5s

  # Invert the check (wait until records don't match)
  wait4x dns A wait4x.dev --expect-ip 172.67.154.180 --invert-check`,
		RunE: runA,
	}

	command.Flags().StringArray("expect-ip", []string{}, "Expected IPv4 addresses to match against A records")

	return command
}

// runA is the command handler for the "dns A" command
func runA(cmd *cobra.Command, args []string) error {
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
		dns.WithExpectedIPV4s(expectIPs),
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
