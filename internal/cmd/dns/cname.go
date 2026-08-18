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
	dns "wait4x.dev/v4/checker/dns/cname"
	"wait4x.dev/v4/internal/contextutil"
	"wait4x.dev/v4/waiter"
)

// NewCNAMECommand creates a new DNS CNAME command
func NewCNAMECommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "CNAME ADDRESS [-- command [args...]]",
		Aliases: []string{"cname"},
		Short:   "Check DNS CNAME records for a given domain",
		Long:    "Check DNS CNAME records and optionally verify expected domain names",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the CNAME command")
			}
			return nil
		},
		Example: `
  # Check CNAME record existence
  wait4x dns CNAME example.com

  # Check CNAME records with expected domains
  wait4x dns CNAME example.com --expected-domain target.example.com

  # Check CNAME record using a specific nameserver
  wait4x dns CNAME example.com --expected-domain target.example.com -n 8.8.8.8

  # Check CNAME record with custom timeout and interval
  wait4x dns CNAME example.com --timeout 30s --interval 5s
`,
		RunE: runCNAME,
	}

	command.Flags().StringArray("expect-domain", nil, "Expect domains.")

	return command
}

// runCNAME is the command handler for the "dns CNAME" command
func runCNAME(cmd *cobra.Command, args []string) error {
	nameserver, err := cmd.Flags().GetString("nameserver")
	if err != nil {
		return fmt.Errorf("failed to parse --nameserver flag: %w", err)
	}

	expectDomains, err := cmd.Flags().GetStringArray("expect-domain")
	if err != nil {
		return fmt.Errorf("failed to parse --expect-domain flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
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
