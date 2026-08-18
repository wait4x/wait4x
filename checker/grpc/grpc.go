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

// Package grpc provides the gRPC health checker for the Wait4X application.
package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"wait4x.dev/v4/checker"
)

// Option configures a GRPC checker
type Option func(*GRPC)

const (
	// DefaultConnectionTimeout is the default connection timeout duration
	DefaultConnectionTimeout = 3 * time.Second

	// DefaultInsecureTransport is the default insecure transport security
	DefaultInsecureTransport = false

	// DefaultInsecureSkipTLSVerify is the default insecure skip tls verify
	DefaultInsecureSkipTLSVerify = false

	// DefaultServiceName is the default service name for health check
	// Empty string checks the overall server health
	DefaultServiceName = ""
)

// GRPC is a gRPC health checker
type GRPC struct {
	address               string
	timeout               time.Duration
	insecureTransport     bool
	insecureSkipTLSVerify bool
	serviceName           string
}

// New creates a new gRPC health checker
func New(address string, opts ...Option) checker.Checker {
	g := &GRPC{
		address:               address,
		timeout:               DefaultConnectionTimeout,
		insecureTransport:     DefaultInsecureTransport,
		insecureSkipTLSVerify: DefaultInsecureSkipTLSVerify,
		serviceName:           DefaultServiceName,
	}

	// Apply the list of options to GRPC
	for _, opt := range opts {
		opt(g)
	}

	return g
}

// WithTimeout configures a timeout for maximum amount of time a dial will wait for a gRPC connection to complete
func WithTimeout(timeout time.Duration) Option {
	return func(g *GRPC) {
		g.timeout = timeout
	}
}

// WithInsecureTransport disables transport security for the gRPC connection
func WithInsecureTransport(insecureTransport bool) Option {
	return func(g *GRPC) {
		g.insecureTransport = insecureTransport
	}
}

// WithInsecureSkipTLSVerify configures insecure skip tls verify
func WithInsecureSkipTLSVerify(insecureSkipTLSVerify bool) Option {
	return func(g *GRPC) {
		g.insecureSkipTLSVerify = insecureSkipTLSVerify
	}
}

// WithServiceName configures the service name to check
// Empty string (default) checks the overall server health
func WithServiceName(serviceName string) Option {
	return func(g *GRPC) {
		g.serviceName = serviceName
	}
}

// Identity returns the identity of the gRPC health checker
func (g *GRPC) Identity() (string, error) {
	if g.serviceName != "" {
		return fmt.Sprintf("%s (service: %s)", g.address, g.serviceName), nil
	}
	return g.address, nil
}

// Check checks the gRPC health
func (g *GRPC) Check(ctx context.Context) error {
	conn, err := g.getGRPCConn()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	return g.checkHealth(ctx, conn)
}

// getGRPCConn creates a gRPC connection
func (g *GRPC) getGRPCConn() (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: g.timeout}
			return d.DialContext(ctx, "tcp", addr)
		}),
	}

	if g.insecureTransport {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(
			opts,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: g.insecureSkipTLSVerify,
			})),
		)
	}

	conn, err := grpc.NewClient(g.address, opts...)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, checker.NewExpectedError(
				"timed out while making a grpc call",
				err,
				"timeout", g.timeout,
			)
		} else if checker.IsConnectionRefused(err) {
			return nil, checker.NewExpectedError(
				"failed to establish a grpc connection",
				err,
				"address", g.address,
			)
		}

		return nil, err
	}

	return conn, nil
}

// checkHealth performs the gRPC health check
func (g *GRPC) checkHealth(ctx context.Context, conn grpc.ClientConnInterface) error {
	healthClient := grpc_health_v1.NewHealthClient(conn)
	req := &grpc_health_v1.HealthCheckRequest{
		Service: g.serviceName,
	}

	resp, err := healthClient.Check(ctx, req)
	if err != nil {
		// Check for common error types and wrap appropriately
		if ctx.Err() == context.Canceled {
			return ctx.Err()
		}
		if ctx.Err() == context.DeadlineExceeded {
			return ctx.Err()
		}
		if os.IsTimeout(err) {
			return checker.NewExpectedError(
				"timed out while performing health check",
				err,
				"service", g.serviceName,
			)
		}
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish a grpc connection",
				err,
				"address", g.address,
				"service", g.serviceName,
			)
		}

		return checker.NewExpectedError(
			"health check failed",
			err,
			"service", g.serviceName,
		)
	}

	status := resp.GetStatus()
	if status != grpc_health_v1.HealthCheckResponse_SERVING {
		return checker.NewExpectedError(
			"service is not serving",
			nil,
			"status", status.String(),
			"expected", grpc_health_v1.HealthCheckResponse_SERVING.String(),
			"service", g.serviceName,
		)
	}

	return nil
}
