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

// Package cname provides the CNAME checker for the Wait4X application.
package cname

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/miekg/dns"
	"wait4x.dev/v4/checker"
	dns2 "wait4x.dev/v4/checker/dns"
)

// Option configures an DNS CNAME record
type Option func(d *CNAME)

// CNAME is a DNS CNAME checker
type CNAME struct {
	nameserver      string
	address         string
	expectedDomains []string
}

// New creates a new DNS CNAME checker
func New(address string, opts ...Option) checker.Checker {
	d := &CNAME{
		address: address,
	}

	// apply the list of options to CNAME
	for _, opt := range opts {
		opt(d)
	}

	return d
}

// WithNameServer overrides the default nameserver for the DNS CNAME checker
func WithNameServer(nameserver string) Option {
	return func(d *CNAME) {
		d.nameserver = nameserver
	}
}

// WithExpectedDomains sets expected domains for the DNS CNAME checker
func WithExpectedDomains(doamins []string) Option {
	return func(d *CNAME) {
		d.expectedDomains = doamins
	}
}

// Identity returns the identity of the DNS CNAME checker
func (d *CNAME) Identity() (string, error) {
	return d.address, nil
}

// Check checks the DNS CNAME records
func (d *CNAME) Check(ctx context.Context) (err error) {
	c := new(dns.Client)
	c.Timeout = dns2.DefaultTimeout

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(d.address), dns.TypeCNAME)
	m.RecursionDesired = true

	r, _, err := c.ExchangeContext(ctx, m, dns2.RR(d.nameserver))
	if err != nil {
		return err
	}

	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("response code is not successful, %d", r.Rcode)
	}

	if len(r.Answer) == 0 {
		return checker.NewExpectedError("no CNAME record found", nil)
	}

	if len(d.expectedDomains) == 0 {
		return nil
	}

	actualRecords := make([]string, 0)
	for _, answer := range r.Answer {
		if cname, ok := answer.(*dns.CNAME); ok {
			actualRecord := strings.TrimSuffix(cname.Target, ".")
			actualRecords = append(actualRecords, actualRecord)

			for _, expectedDomain := range d.expectedDomains {
				matched, _ := regexp.MatchString(expectedDomain, actualRecord)
				if matched {
					return nil
				}
			}
		}
	}

	return checker.NewExpectedError(
		"the CNAME record value doesn't match expected",
		nil,
		"actual", actualRecords, "expect", d.expectedDomains,
	)
}
