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

// Package txt provides the TXT checker for the Wait4X application.
package txt

import (
	"context"
	"net"
	"regexp"

	"wait4x.dev/v4/checker"
)

// Option configures an DNS TXT records
type Option func(d *TXT)

// TXT is a DNS TXT checker
type TXT struct {
	nameserver     string
	address        string
	expectedValues []string
	resolver       *net.Resolver
}

// New creates a new DNS TXT checker
func New(address string, opts ...Option) checker.Checker {
	d := &TXT{
		address:  address,
		resolver: net.DefaultResolver,
	}

	// apply the list of options to TXT
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

// WithNameServer overrides the default nameserver for the DNS TXT checker
func WithNameServer(nameserver string) Option {
	return func(d *TXT) {
		d.nameserver = nameserver
	}
}

// WithExpectedValues sets expected values for the DNS TXT checker
func WithExpectedValues(values []string) Option {
	return func(d *TXT) {
		d.expectedValues = values
	}
}

// Identity returns the identity of the DNS TXT checker
func (d *TXT) Identity() (string, error) {
	return d.address, nil
}

// Check checks the DNS TXT records
func (d *TXT) Check(ctx context.Context) (err error) {
	values, err := d.resolver.LookupTXT(ctx, d.address)
	if err != nil {
		return err
	}

	for _, txt := range values {
		if len(d.expectedValues) == 0 {
			return nil
		}
		for _, expectedValue := range d.expectedValues {
			matched, _ := regexp.MatchString(expectedValue, txt)
			if matched {
				return nil
			}
		}
	}

	return checker.NewExpectedError(
		"the TXT record value doesn't expect", nil,
		"actual", values, "expect", d.expectedValues,
	)
}
