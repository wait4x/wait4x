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

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"wait4x.dev/v4/checker"
	"wait4x.dev/v4/checker/grpc"
	"wait4x.dev/v4/internal/contextutil"
	"wait4x.dev/v4/waiter"
)

// NewGRPCCommand creates a new gRPC command
func NewGRPCCommand() *cobra.Command {
	grpcCommand := &cobra.Command{
		Use:   "grpc ADDRESS... [flags] [-- command [args...]]",
		Short: "Check gRPC server health",
		Long: `Check gRPC server health using the standard gRPC Health Checking Protocol.

The gRPC health checker connects to a gRPC server and performs a health check
using the standard grpc.health.v1.Health service. This is compatible with any
gRPC server that implements the standard health checking protocol.

The protocol is defined at: https://github.com/grpc/grpc/blob/master/doc/health-checking.md

Examples:
  # Check if gRPC server is healthy (overall server health)
  wait4x grpc localhost:50051 --insecure-transport

  # Check specific service health
  wait4x grpc localhost:50051 --service myapp.UserService --insecure-transport

  # Check with TLS
  wait4x grpc api.example.com:443

  # Check with TLS but skip certificate verification
  wait4x grpc localhost:50051 --insecure-skip-tls-verify

  # Check with custom timeout
  wait4x grpc localhost:50051 --connection-timeout 10s --insecure-transport

  # Wait for gRPC service and then run a command
  wait4x grpc localhost:50051 --insecure-transport -- ./start-client.sh

  # Check multiple gRPC servers in parallel
  wait4x grpc localhost:50051 localhost:50052 localhost:50053 --insecure-transport`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the grpc command")
			}
			return nil
		},
		RunE: runGRPC,
	}

	// Add flags
	grpcCommand.Flags().Duration("connection-timeout", grpc.DefaultConnectionTimeout, "Connection timeout for the gRPC call")
	grpcCommand.Flags().Bool("insecure-transport", grpc.DefaultInsecureTransport, "Disable transport security (no TLS)")
	grpcCommand.Flags().Bool("insecure-skip-tls-verify", grpc.DefaultInsecureSkipTLSVerify, "Skip TLS certificate verification")
	grpcCommand.Flags().String("service", "", "Service name to check (empty for overall server health)")

	return grpcCommand
}

func runGRPC(cmd *cobra.Command, args []string) error {
	connectionTimeout, err := cmd.Flags().GetDuration("connection-timeout")
	if err != nil {
		return fmt.Errorf("failed to parse --connection-timeout flag: %w", err)
	}

	insecureTransport, err := cmd.Flags().GetBool("insecure-transport")
	if err != nil {
		return fmt.Errorf("failed to parse --insecure-transport flag: %w", err)
	}

	insecureSkipTLSVerify, err := cmd.Flags().GetBool("insecure-skip-tls-verify")
	if err != nil {
		return fmt.Errorf("failed to parse --insecure-skip-tls-verify flag: %w", err)
	}

	serviceName, err := cmd.Flags().GetString("service")
	if err != nil {
		return fmt.Errorf("failed to parse --service flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	// Create options
	opts := []grpc.Option{
		grpc.WithTimeout(connectionTimeout),
		grpc.WithInsecureTransport(insecureTransport),
		grpc.WithInsecureSkipTLSVerify(insecureSkipTLSVerify),
	}

	if serviceName != "" {
		opts = append(opts, grpc.WithServiceName(serviceName))
	}

	// Create checkers for all addresses
	checkers := make([]checker.Checker, len(args))
	for i, address := range args {
		checkers[i] = grpc.New(address, opts...)
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
