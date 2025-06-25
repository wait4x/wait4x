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

// Package kafka provides Kafka checker.
package kafka

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
	"wait4x.dev/v3/checker"
)

const (
	// DefaultConnectionTimeout is the default connection timeout duration
	DefaultConnectionTimeout = 100 * time.Millisecond
)

// Kafka represents Kafka checker
type Kafka struct {
	dsn string
}

// New creates the Kafka checker
func New(dsn string) checker.Checker {
	i := &Kafka{
		dsn: dsn,
	}

	return i
}

// Identity returns the identity of the checker
func (k *Kafka) Identity() (string, error) {
	_, _, _, broker, err := parseDSN(k.dsn)
	if err != nil {
		return "", fmt.Errorf("failed to parse DSN: %w", err)
	}

	return broker, nil
}

// parseDSN parses the DSN and returns the authentication mechanism, password, username, and broker address.
// The DSN format is expected to be:
// kafka://username:password@broker:port?authMechanism=scram-sha-512
func parseDSN(dsn string) (authMechanism, password, username, broker string, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse DSN %q: %w", dsn, err)
	}

	if u.Scheme != "kafka" {
		return "", "", "", "", fmt.Errorf("invalid DSN scheme %q, expected 'kafka'", u.Scheme)
	}

	broker = u.Host
	if broker == "" {
		return "", "", "", "", fmt.Errorf("broker address is required in DSN %q", dsn)
	}

	username = u.User.Username()
	password, _ = u.User.Password()
	authMechanism = u.Query().Get("authMechanism")

	return
}

// Check checks Kafka connection
func (k *Kafka) Check(ctx context.Context) (err error) {
	authMechanism, password, username, broker, err := parseDSN(k.dsn)
	if err != nil {
		return fmt.Errorf("failed to parse DSN %q: %w", k.dsn, err)
	}

	var saslMechanism sasl.Mechanism

	switch strings.ToUpper(authMechanism) {
	case scram.SHA256.Name():
		saslMechanism, err = scram.Mechanism(scram.SHA256, username, password)
	case scram.SHA512.Name():
		saslMechanism, err = scram.Mechanism(scram.SHA512, username, password)
	case "":
		saslMechanism = nil
	default:
		err = fmt.Errorf("unknown auth mechanism %q", authMechanism)
	}

	if err != nil {
		return fmt.Errorf("failed to create SASL mechanism: %w", err)
	}

	dialer := &kafka.Dialer{
		SASLMechanism: saslMechanism,
		ClientID:      "wait4x-kafka-checker",
		Timeout:       DefaultConnectionTimeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", broker)
	if err != nil {
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the Kafka server",
				err,
				"broker", broker,
				"authMechanism", authMechanism,
				"username", username,
				"password", trimPassword(password),
			)
		}

		return fmt.Errorf("failed to connect to Kafka broker %s: %w", broker, err)
	}

	defer conn.Close()

	// Use it as alternative to ping the broker
	_, err = conn.Brokers()
	if err != nil {
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to get Kafka broker list",
				err,
				"broker", broker,
			)
		}

		return fmt.Errorf("failed to get Kafka broker list %s: %w", broker, err)
	}

	return nil
}

func trimPassword(password string) string {
	if len(password) > 4 {
		return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
	}

	return strings.Repeat("*", len(password))
}
