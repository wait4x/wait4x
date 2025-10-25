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

// Package mx provides the MX checker for the Wait4X application.
package mx

import (
	"context"
	"fmt"
	"net"
	"regexp"

	"wait4x.dev/v4/checker"
)

// Option configures an DNS MX records
type Option func(d *MX)

// MX is a DNS MX checker
type MX struct {
	nameserver      string
	address         string
	expectedDomains []string
	resolver        *net.Resolver
}

// New creates a new DNS MX checker
func New(address string, opts ...Option) checker.Checker {
	d := &MX{
		address:  address,
		resolver: net.DefaultResolver,
	}

	// apply the list of options to MX
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

// WithNameServer overrides the default nameserver for the DNS MX checker
func WithNameServer(nameserver string) Option {
	return func(d *MX) {
		d.nameserver = nameserver
	}
}

// WithExpectedDomains sets expected domains for the DNS MX checker
func WithExpectedDomains(domains []string) Option {
	return func(d *MX) {
		d.expectedDomains = domains
	}
}

// Identity returns the identity of the DNS MX checker
func (d *MX) Identity() (string, error) {
	return fmt.Sprintf("MX %s %s", d.address, d.expectedDomains), nil
}

// Check checks the DNS MX records
func (d *MX) Check(ctx context.Context) (err error) {
	values, err := d.resolver.LookupMX(ctx, d.address)
	if err != nil {
		return err
	}

	for _, mx := range values {
		if len(d.expectedDomains) == 0 {
			return nil
		}
		for _, expectedDomain := range d.expectedDomains {
			matched, _ := regexp.MatchString(expectedDomain, mx.Host)
			if matched {
				return nil
			}
		}
	}

	return checker.NewExpectedError(
		"the MX record value doesn't expect", nil,
		"actual", values, "expect", d.expectedDomains,
	)
}
