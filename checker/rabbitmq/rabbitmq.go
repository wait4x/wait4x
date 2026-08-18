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

// Package rabbitmq provides the RabbitMQ checker for the Wait4X application.
package rabbitmq

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"wait4x.dev/v4/checker"
)

var hidePasswordRegexp = regexp.MustCompile(`(amqp://[^/:]+):[^:@]+@`)

// Option configures a RabbitMQ.
type Option func(r *RabbitMQ)

const (
	// DefaultHeartbeat is the default heartbeat duration
	DefaultHeartbeat = 10 * time.Second
	// DefaultConnectionTimeout is the default connection timeout duration
	DefaultConnectionTimeout = 3 * time.Second
	// DefaultLocale is the default locale
	DefaultLocale = "en_US"
	// DefaultInsecureSkipTLSVerify is the default value for whether to skip tls verify
	DefaultInsecureSkipTLSVerify = false
)

// RabbitMQ is a RabbitMQ checker
type RabbitMQ struct {
	dsn                   string
	timeout               time.Duration
	insecureSkipTLSVerify bool
}

// New creates a new RabbitMQ checker
func New(dsn string, opts ...Option) checker.Checker {
	t := &RabbitMQ{
		dsn:                   dsn,
		timeout:               DefaultConnectionTimeout,
		insecureSkipTLSVerify: DefaultInsecureSkipTLSVerify,
	}

	// Apply options to RabbitMQ checker
	for _, opt := range opts {
		opt(t)
	}

	return t
}

// WithTimeout configures a timeout for establishing new connections
func WithTimeout(timeout time.Duration) Option {
	return func(r *RabbitMQ) {
		r.timeout = timeout
	}
}

// WithInsecureSkipTLSVerify configures whether to skip tls verify
func WithInsecureSkipTLSVerify(insecureSkipTLSVerify bool) Option {
	return func(r *RabbitMQ) {
		r.insecureSkipTLSVerify = insecureSkipTLSVerify
	}
}

// Identity returns the identity of the RabbitMQ checker
func (r *RabbitMQ) Identity() (string, error) {
	u, err := amqp.ParseURI(r.dsn)
	if err != nil {
		return "", fmt.Errorf("can't retrieve the checker identity: %w", err)
	}

	return fmt.Sprintf("%s:%d", u.Host, u.Port), nil
}

// Check checks the RabbitMQ connection
func (r *RabbitMQ) Check(ctx context.Context) (err error) {
	conn, err := amqp.DialConfig(
		r.dsn,
		amqp.Config{
			Heartbeat: DefaultHeartbeat,
			Locale:    DefaultLocale,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: r.insecureSkipTLSVerify,
			},
			Dial: func(network, addr string) (net.Conn, error) {
				d := net.Dialer{Timeout: r.timeout}
				conn, err := d.DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}

				// Heartbeating hasn't started yet, don't stall forever on a dead server.
				// A deadline is set for TLS and AMQP handshaking. After AMQP is established,
				// the deadline is cleared in openComplete.
				if err := conn.SetDeadline(time.Now().Add(r.timeout)); err != nil {
					return nil, err
				}

				return conn, nil
			},
		},
	)

	if err != nil {
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the rabbitmq server", err,
				"dsn", hidePasswordRegexp.ReplaceAllString(r.dsn, `$1:***@`),
			)
		}

		return err
	}

	defer func(conn *amqp.Connection) {
		if connerr := conn.Close(); connerr != nil {
			err = connerr
		}
	}(conn)

	// Open a channel to check the connection.
	_, err = conn.Channel()
	if err != nil {
		return err
	}

	return nil
}
