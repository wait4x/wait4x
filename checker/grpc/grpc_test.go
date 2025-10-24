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
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"wait4x.dev/v4/checker"
)

// GRPCSuite is a test suite for gRPC health checker
type GRPCSuite struct {
	suite.Suite

	// Shared resources for the test suite
	server       *grpc.Server
	healthServer *health.Server
	listener     net.Listener
	serverAddr   string
}

// SetupSuite sets up test suite resources
func (s *GRPCSuite) SetupSuite() {
	// Create a listener on a random port
	var err error
	s.listener, err = net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	s.serverAddr = s.listener.Addr().String()

	// Create gRPC server with health check
	s.server = grpc.NewServer()
	s.healthServer = health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.server, s.healthServer)

	// Start server in background
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			s.T().Logf("Server stopped: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
}

// TearDownSuite tears down test suite resources
func (s *GRPCSuite) TearDownSuite() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// TestNew checks the constructor with default and custom options
func (s *GRPCSuite) TestNew() {
	// Test default values
	gc := New("localhost:50051").(*GRPC)
	s.Equal("localhost:50051", gc.address)
	s.Equal(DefaultConnectionTimeout, gc.timeout)
	s.Equal(DefaultInsecureTransport, gc.insecureTransport)
	s.Equal(DefaultInsecureSkipTLSVerify, gc.insecureSkipTLSVerify)
	s.Equal(DefaultServiceName, gc.serviceName)

	// Test with options
	customTimeout := 5 * time.Second
	gc = New(
		"localhost:50051",
		WithTimeout(customTimeout),
		WithInsecureTransport(true),
		WithServiceName("myservice"),
	).(*GRPC)
	s.Equal("localhost:50051", gc.address)
	s.Equal(customTimeout, gc.timeout)
	s.True(gc.insecureTransport)
	s.Equal("myservice", gc.serviceName)
}

// TestIdentity tests the Identity method
func (s *GRPCSuite) TestIdentity() {
	// Without service name
	gc := New("localhost:50051")
	identity, err := gc.Identity()
	s.NoError(err)
	s.Equal("localhost:50051", identity)

	// With service name
	gc = New("localhost:50051", WithServiceName("myservice"))
	identity, err = gc.Identity()
	s.NoError(err)
	s.Equal("localhost:50051 (service: myservice)", identity)
}

// TestCheckSuccessful tests successful health check
func (s *GRPCSuite) TestCheckSuccessful() {
	// Set server as serving
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	gc := New(s.serverAddr, WithInsecureTransport(true))
	err := gc.Check(context.Background())
	s.NoError(err)
}

// TestCheckWithServiceName tests health check with specific service
func (s *GRPCSuite) TestCheckWithServiceName() {
	serviceName := "test.Service"

	// Set specific service as serving
	s.healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	gc := New(
		s.serverAddr,
		WithInsecureTransport(true),
		WithServiceName(serviceName),
	)
	err := gc.Check(context.Background())
	s.NoError(err)
}

// TestCheckNotServing tests health check when service is not serving
func (s *GRPCSuite) TestCheckNotServing() {
	serviceName := "test.NotServing"

	// Set service as not serving
	s.healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	gc := New(
		s.serverAddr,
		WithInsecureTransport(true),
		WithServiceName(serviceName),
	)

	err := gc.Check(context.Background())
	s.Error(err)

	var expectedErr *checker.ExpectedError
	s.True(errors.As(err, &expectedErr))
	s.Contains(err.Error(), "service is not serving")
}

// TestCheckConnectionRefused tests connection refused error
func (s *GRPCSuite) TestCheckConnectionRefused() {
	gc := New(
		"127.0.0.1:19999", // Unlikely to be in use
		WithInsecureTransport(true),
		WithTimeout(500*time.Millisecond),
	)

	err := gc.Check(context.Background())
	s.Error(err)

	var expectedErr *checker.ExpectedError
	s.True(errors.As(err, &expectedErr))
	s.Contains(err.Error(), "connection refused")
}

// TestCheckTimeout tests timeout behavior
func (s *GRPCSuite) TestCheckTimeout() {
	// Use a black-hole IP that will cause timeout
	gc := New(
		"240.0.0.1:50051",
		WithInsecureTransport(true),
		WithTimeout(500*time.Millisecond),
	)

	start := time.Now()
	err := gc.Check(context.Background())
	elapsed := time.Since(start)

	s.Error(err)
	s.True(elapsed >= 500*time.Millisecond, "Timeout was not respected")

	var expectedErr *checker.ExpectedError
	if s.True(errors.As(err, &expectedErr)) {
		s.Contains(err.Error(), "timeout")
	}
}

// TestCheckContextCancellation tests context cancellation
func (s *GRPCSuite) TestCheckContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Use a black-hole IP to ensure the operation would take time
	gc := New(
		"240.0.0.1:50051",
		WithInsecureTransport(true),
		WithTimeout(10*time.Second),
	)

	// Cancel the context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := gc.Check(ctx)
	elapsed := time.Since(start)

	s.Error(err)
	s.True(elapsed < 5*time.Second, "Context cancellation was not respected")
	// Context cancellation should propagate through
	s.Contains(err.Error(), "canceled")
}

// TestTableDriven defines table-driven tests for various scenarios
func (s *GRPCSuite) TestTableDriven() {
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func() string // Returns address to test
		options     []Option
		shouldError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name: "Valid Address with Serving Status",
			setup: func() string {
				s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
				return s.serverAddr
			},
			options:     []Option{WithInsecureTransport(true)},
			shouldError: false,
		},
		{
			name: "Service Not Found",
			setup: func() string {
				return s.serverAddr
			},
			options: []Option{
				WithInsecureTransport(true),
				WithServiceName("nonexistent.Service"),
			},
			shouldError: true,
			errorCheck: func(t *testing.T, err error) {
				var expectedErr *checker.ExpectedError
				s.True(errors.As(err, &expectedErr))
			},
		},
		{
			name: "Connection Refused",
			setup: func() string {
				return "127.0.0.1:19999"
			},
			options: []Option{
				WithInsecureTransport(true),
				WithTimeout(500 * time.Millisecond),
			},
			shouldError: true,
			errorCheck: func(t *testing.T, err error) {
				var expectedErr *checker.ExpectedError
				s.True(errors.As(err, &expectedErr))
				s.Contains(err.Error(), "connection refused")
			},
		},
		{
			name: "Service Unknown Status",
			setup: func() string {
				s.healthServer.SetServingStatus("test.Unknown", grpc_health_v1.HealthCheckResponse_UNKNOWN)
				return s.serverAddr
			},
			options: []Option{
				WithInsecureTransport(true),
				WithServiceName("test.Unknown"),
			},
			shouldError: true,
			errorCheck: func(t *testing.T, err error) {
				var expectedErr *checker.ExpectedError
				s.True(errors.As(err, &expectedErr))
				s.Contains(err.Error(), "service is not serving")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			addr := tt.setup()
			gc := New(addr, tt.options...)
			err := gc.Check(ctx)

			if tt.shouldError {
				s.Error(err)
				if tt.errorCheck != nil {
					tt.errorCheck(s.T(), err)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestGRPCSuite runs the test suite
func TestGRPCSuite(t *testing.T) {
	suite.Run(t, new(GRPCSuite))
}
