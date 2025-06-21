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

//go:build !disable_rabbitmq

package cmd

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/rabbitmq"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewRabbitMQCommand creates a new rabbitmq sub-command
func NewRabbitMQCommand() *cobra.Command {
	rabbitmqCommand := &cobra.Command{
		Use:   "rabbitmq DSN... [flags] [-- command [args...]]",
		Short: "Check RabbitMQ connection",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("DSN is required argument for the rabbitmq sub-command")
			}

			return nil
		},
		Example: `
  # Checking RabbitMQ connection
  wait4x rabbitmq 'amqp://127.0.0.1:5672'

  # Checking RabbitMQ connection with credentials and vhost
  wait4x rabbitmq 'amqp://guest:guest@127.0.0.1:5672/vhost'
`,
		RunE: runRabbitMQ,
	}

	rabbitmqCommand.Flags().Duration("connection-timeout", rabbitmq.DefaultConnectionTimeout, "Timeout is the maximum amount of time a dial will wait for a connection to complete.")
	rabbitmqCommand.Flags().Bool("insecure-skip-tls-verify", rabbitmq.DefaultInsecureSkipTLSVerify, "InsecureSkipTLSVerify controls whether a client verifies the server's certificate chain and hostname.")

	return rabbitmqCommand
}

func runRabbitMQ(cmd *cobra.Command, args []string) error {
	conTimeout, err := cmd.Flags().GetDuration("connection-timeout")
	if err != nil {
		return fmt.Errorf("unable to parse --connection-timeout flag: %w", err)
	}

	insecureSkipTLSVerify, err := cmd.Flags().GetBool("insecure-skip-tls-verify")
	if err != nil {
		return fmt.Errorf("unable to parse --insecure-skip-tls-verify flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("unable to get logger from context: %w", err)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	checkers := make([]checker.Checker, len(args))
	for i, arg := range args {
		checkers[i] = rabbitmq.New(
			arg,
			rabbitmq.WithTimeout(conTimeout),
			rabbitmq.WithInsecureSkipTLSVerify(insecureSkipTLSVerify),
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
