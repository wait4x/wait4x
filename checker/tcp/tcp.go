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

// Package tcp provides the TCP checker for the Wait4X application.
package tcp

import (
	"context"
	"net"
	"os"
	"time"

	"wait4x.dev/v3/checker"
)

// Option configures a TCP checker
type Option func(t *TCP)

const (
	// DefaultConnectionTimeout is the default connection timeout duration
	DefaultConnectionTimeout = 3 * time.Second
)

// TCP is a TCP checker
type TCP struct {
	address string
	timeout time.Duration
}

// New creates a new TCP checker
func New(address string, opts ...Option) checker.Checker {
	t := &TCP{
		address: address,
		timeout: DefaultConnectionTimeout,
	}

	// Apply the list of options to TCP
	for _, opt := range opts {
		opt(t)
	}

	return t
}

// WithTimeout configures a timeout for maximum amount of time a dial will wait for a connection to complete
func WithTimeout(timeout time.Duration) Option {
	return func(t *TCP) {
		t.timeout = timeout
	}
}

// Identity returns the identity of the TCP checker
func (t *TCP) Identity() (string, error) {
	return t.address, nil
}

// Check checks the TCP connection
func (t *TCP) Check(ctx context.Context) error {
	d := net.Dialer{Timeout: t.timeout}

	_, err := d.DialContext(ctx, "tcp", t.address)
	if err != nil {
		if os.IsTimeout(err) {
			return checker.NewExpectedError("timed out while making a tcp call", err, "timeout", t.timeout)
		} else if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError("failed to establish a tcp connection", err)
		}

		return err
	}

	return nil
}
