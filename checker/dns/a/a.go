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

// Package a provides the A checker for the Wait4X application.
package a

import (
	"context"
	"net"

	"wait4x.dev/v4/checker"
)

// Option configures an DNS A records
type Option func(d *A)

// A is a DNS A checker
type A struct {
	nameserver  string
	address     string
	expectedIPs []string
	resolver    *net.Resolver
}

// New creates a new DNS A checker for the given address
func New(address string, opts ...Option) checker.Checker {
	d := &A{
		address:  address,
		resolver: net.DefaultResolver,
	}

	// Apply the list of options to A
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

// WithNameServer overrides the default nameserver for the DNS A checker
func WithNameServer(nameserver string) Option {
	return func(d *A) {
		d.nameserver = nameserver
	}
}

// WithExpectedIPV4s sets expected IPv4s for the DNS A checker
func WithExpectedIPV4s(ips []string) Option {
	return func(d *A) {
		d.expectedIPs = ips
	}
}

// Identity returns the identity of the DNS A checker
func (d *A) Identity() (string, error) {
	return d.address, nil
}

// Check checks the DNS A records
func (d *A) Check(ctx context.Context) (err error) {
	ips, err := d.resolver.LookupIP(ctx, "ip4", d.address)
	if err != nil {
		return err
	}

	for _, ip := range ips {
		if len(d.expectedIPs) == 0 {
			return nil
		}
		for _, expectedIP := range d.expectedIPs {
			if expectedIP == ip.String() {
				return nil
			}
		}
	}

	return checker.NewExpectedError(
		"the A record value doesn't expect", nil,
		"actual", ips, "expect", d.expectedIPs,
	)
}
