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

// Package ns provides the NS checker for the Wait4X application.
package ns

import (
	"context"
	"net"
	"regexp"

	"wait4x.dev/v4/checker"
)

// Option configures an DNS NS records
type Option func(d *NS)

// NS is a DNS NS checker
type NS struct {
	nameserver          string
	address             string
	expectedNameservers []string
	resolver            *net.Resolver
}

// New creates a new DNS NS checker
func New(address string, opts ...Option) checker.Checker {
	d := &NS{
		address:  address,
		resolver: net.DefaultResolver,
	}

	// apply the list of options to NS
	for _, opt := range opts {
		opt(d)
	}

	// Nameserver settings.
	if d.nameserver != "" {
		d.resolver = &net.Resolver{
			Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, network, d.nameserver)
			},
		}
	}

	return d
}

// WithNameServer overrides the default nameserver for the DNS NS checker
func WithNameServer(nameserver string) Option {
	return func(d *NS) {
		d.nameserver = nameserver
	}
}

// WithExpectedNameservers sets expected nameservers for the DNS NS checker
func WithExpectedNameservers(nameservers []string) Option {
	return func(d *NS) {
		d.expectedNameservers = nameservers
	}
}

// Identity returns the identity of the DNS NS checker
func (d *NS) Identity() (string, error) {
	return d.address, nil
}

// Check checks the DNS NS records
func (d *NS) Check(ctx context.Context) (err error) {
	values, err := d.resolver.LookupNS(ctx, d.address)
	if err != nil {
		return err
	}

	for _, ns := range values {
		if len(d.expectedNameservers) == 0 {
			return nil
		}
		for _, expectedNameserver := range d.expectedNameservers {
			matched, _ := regexp.MatchString(expectedNameserver, ns.Host)
			if matched {
				return nil
			}
		}
	}

	return checker.NewExpectedError(
		"the NS record value doesn't expect", nil,
		"actual", values, "expect", d.expectedNameservers,
	)
}
