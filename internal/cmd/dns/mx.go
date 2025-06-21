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
	dns "wait4x.dev/v3/checker/dns/mx"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewMXCommand creates a new DNS MX command
func NewMXCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "MX ADDRESS [-- command [args...]]",
		Aliases: []string{"mx"},
		Short:   "Check DNS MX (mail exchanger) records for a given domain",
		Long:    "Check for the existence and validity of DNS MX (mail exchanger) records for a specified domain. MX records specify the mail servers responsible for receiving email on behalf of the domain.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the MX command")
			}

			return nil
		},
		Example: `
  # Check MX records existence
  wait4x dns MX wait4x.dev

  # Check MX records with expected domains
  wait4x dns MX wait4x.dev --expect-domain 'route1.mx.cloudflare.net'

  # Check MX records with multiple expected domains
  wait4x dns MX wait4x.dev --expect-domain 'route1.mx.cloudflare.net' --expect-domain 'route2.mx.cloudflare.net'

  # Check MX records by defined nameserver
  wait4x dns MX wait4x.dev --expect-domain 'route1.mx.cloudflare.net' --nameserver 'gordon.ns.cloudflare.com'

  # Check MX records with custom timeout and interval
  wait4x dns MX wait4x.dev --timeout 30s --interval 5s`,
		RunE: runMX,
	}

	command.Flags().StringArray("expect-domain", nil, "Expected domain names in MX records. Can be specified multiple times for multiple domains.")

	return command
}

// runMX is the command handler for the "dns MX" command
func runMX(cmd *cobra.Command, args []string) error {
	nameserver, err := cmd.Flags().GetString("nameserver")
	if err != nil {
		return fmt.Errorf("unable to parse --nameserver flag: %w", err)
	}

	expectDomains, err := cmd.Flags().GetStringArray("expect-domain")
	if err != nil {
		return fmt.Errorf("unable to parse --expect-domain flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("unable to get logger from context: %w", err)
	}

	dc := dns.New(
		args[0],
		dns.WithExpectedDomains(expectDomains),
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
