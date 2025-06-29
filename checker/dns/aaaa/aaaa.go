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

// Package aaaa provides functionality for checking the AAAA records of a domain.
package aaaa

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	dns2 "wait4x.dev/v3/checker/dns"

	"wait4x.dev/v3/checker"
)

// Option configures an DNS AAAA records
type Option func(d *AAAA)

// AAAA represents DNS AAAA data structure
type AAAA struct {
	nameserver  string
	address     string
	expectedIPs []string
}

// New creates a new AAAA checker with the given address and optional configuration options.
func New(address string, opts ...Option) checker.Checker {
	d := &AAAA{
		address: address,
	}

	// apply the list of options to AAAA
	for _, opt := range opts {
		opt(d)
	}

	return d
}

// WithNameServer overrides the default nameserver
func WithNameServer(nameserver string) Option {
	return func(d *AAAA) {
		d.nameserver = nameserver
	}
}

// WithExpectedIPV6s sets expected IPv6s
func WithExpectedIPV6s(ips []string) Option {
	return func(d *AAAA) {
		d.expectedIPs = ips
	}
}

// Identity returns the identity of the checker
func (d *AAAA) Identity() (string, error) {
	return d.address, nil
}

// Check checks DNS records
func (d *AAAA) Check(ctx context.Context) (err error) {
	c := new(dns.Client)
	c.Timeout = dns2.DefaultTimeout

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(d.address), dns.TypeAAAA)
	m.RecursionDesired = true

	r, _, err := c.ExchangeContext(ctx, m, dns2.RR(d.nameserver))
	if err != nil {
		return err
	}

	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("response code is not successful, %d", r.Rcode)
	}

	if len(r.Answer) == 0 {
		return checker.NewExpectedError("no AAAA record found", nil)
	}

	if len(d.expectedIPs) == 0 {
		return nil
	}

	actualRecords := make([]string, 0)
	for _, answer := range r.Answer {
		if aaaa, ok := answer.(*dns.AAAA); ok {
			actualRecord := aaaa.AAAA.String()
			actualRecords = append(actualRecords, actualRecord)

			for _, expectedIP := range d.expectedIPs {
				if expectedIP == actualRecord {
					return nil
				}
			}
		}
	}

	return checker.NewExpectedError(
		"the AAAA record value doesn't match expected",
		nil,
		"actual", actualRecords, "expect", d.expectedIPs,
	)
}
