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

// Package redis provides the Redis checker for the Wait4X application.
package redis

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"wait4x.dev/v3/checker"
)

var hidePasswordRegexp = regexp.MustCompile(`([^/]+//[^/:]+):[^:@]+@`)

// Option configures a Redis.
type Option func(r *Redis)

const (
	// DefaultConnectionTimeout is the default connection timeout duration
	DefaultConnectionTimeout = 3 * time.Second
)

// Redis is a Redis checker
type Redis struct {
	address   string
	expectKey string
	timeout   time.Duration
}

// New creates a new Redis checker
func New(address string, opts ...Option) checker.Checker {
	r := &Redis{
		address: address,
		timeout: DefaultConnectionTimeout,
	}

	// apply the list of options to Redis
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithTimeout configures a timeout for establishing new connections
func WithTimeout(timeout time.Duration) Option {
	return func(r *Redis) {
		r.timeout = timeout
	}
}

// WithExpectKey configures a key expectation
func WithExpectKey(key string) Option {
	return func(r *Redis) {
		r.expectKey = key
	}
}

// Identity returns the identity of the Redis checker
func (r *Redis) Identity() (string, error) {
	opts, err := redis.ParseURL(r.address)
	if err != nil {
		return "", fmt.Errorf("can't retrieve the checker identity: %w", err)
	}

	return opts.Addr, nil
}

// Check checks the Redis connection
func (r *Redis) Check(ctx context.Context) error {
	opts, err := redis.ParseURL(r.address)
	if err != nil {
		return err
	}
	opts.DialTimeout = r.timeout

	client := redis.NewClient(opts)

	// Check Redis connection
	_, err = client.Ping(ctx).Result()
	if err != nil {
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the redis server", err,
				"dsn", hidePasswordRegexp.ReplaceAllString(r.address, `$1:***@`),
			)
		}

		return err
	}

	// It can connect to Redis successfully
	if r.expectKey == "" {
		return nil
	}

	splittedKey := strings.Split(r.expectKey, "=")
	keyHasValue := len(splittedKey) == 2

	val, err := client.Get(ctx, splittedKey[0]).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Redis key does not exist.
			return checker.NewExpectedError("the key doesn't exist", nil, "key", splittedKey[0])
		}

		// Error occurred on get Redis key
		return err
	}

	// The Redis key exists and user doesn't want to match value
	if !keyHasValue {
		return nil
	}

	// When the user expect a key with value
	matched, _ := regexp.MatchString(splittedKey[1], val)
	if matched {
		return nil
	}

	return checker.NewExpectedError(
		"the key and desired value doesn't exist", nil,
		"key", splittedKey[0], "actual", val, "expect", splittedKey[1],
	)
}
