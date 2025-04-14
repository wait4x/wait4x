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

//go:build !disable_redis

package cmd

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/redis"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewRedisCommand creates the redis sub-command
func NewRedisCommand() *cobra.Command {
	redisCommand := &cobra.Command{
		Use:   "redis ADDRESS... [flags] [-- command [args...]]",
		Short: "Check Redis connection or key existence",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the redis command")
			}

			return nil
		},
		Example: `
  # Checking Redis connection
  wait4x redis redis://127.0.0.1:6379

  # Specify username, password and db
  wait4x redis redis://user:password@localhost:6379/1

  # Checking Redis connection over unix socket
  wait4x redis unix://user:password@/path/to/redis.sock?db=1

  # Checking a key existence
  wait4x redis redis://127.0.0.1:6379 --expect-key FOO

  # Checking a key existence and matching the value
  wait4x redis redis://127.0.0.1:6379 --expect-key "FOO=^b[A-Z]r$"
`,
		RunE: runRedis,
	}

	redisCommand.Flags().Duration("connection-timeout", redis.DefaultConnectionTimeout, "Dial timeout for establishing new connections.")
	redisCommand.Flags().String("expect-key", "", "Checking key existence.")

	return redisCommand
}

func runRedis(cmd *cobra.Command, args []string) error {
	conTimeout, err := cmd.Flags().GetDuration("connection-timeout")
	if err != nil {
		return fmt.Errorf("failed to parse --connection-timeout flag: %w", err)
	}

	expectKey, err := cmd.Flags().GetString("expect-key")
	if err != nil {
		return fmt.Errorf("failed to parse --expect-key flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	checkers := make([]checker.Checker, len(args))
	for i, arg := range args {
		checkers[i] = redis.New(
			arg,
			redis.WithExpectKey(expectKey),
			redis.WithTimeout(conTimeout),
		)
	}

	return waiter.WaitParallelContext(
		cmd.Context(),
		checkers,
		waiter.WithTimeout(contextutil.GetTimeout(cmd.Context())),
		waiter.WithInterval(contextutil.GetInterval(cmd.Context())),
		waiter.WithInvertCheck(contextutil.GetInvertCheck(cmd.Context())),
		waiter.WithBackoffPolicy(contextutil.GetBackoffPolicy(cmd.Context())),
		waiter.WithBackoffCoefficient(contextutil.GetBackoffCoefficient(cmd.Context())),
		waiter.WithBackoffExponentialMaxInterval(contextutil.GetBackoffExponentialMaxInterval(cmd.Context())),
		waiter.WithLogger(logger),
	)
}
