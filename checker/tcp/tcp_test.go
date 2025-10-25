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
	"errors"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"wait4x.dev/v4/checker"
)

// TCPSuite is a test suite for TCP checker
type TCPSuite struct {
	suite.Suite

	// Shared resources for the test suite
	listener     net.Listener
	ipv6Listener net.Listener
	port         int
	unusedPort   int
	serverDone   chan struct{}
}

// SetupSuite sets up test suite resources
func (s *TCPSuite) SetupSuite() {
	// Set up a TCP server for tests that need an active connection
	var err error
	s.listener, err = net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)

	// Parse the port
	_, portStr, err := net.SplitHostPort(s.listener.Addr().String())
	s.Require().NoError(err)
	s.port, err = strconv.Atoi(portStr)
	s.Require().NoError(err)

	// Find an unused port for connection refused tests
	s.unusedPort = s.port + 1

	// Set up a channel to track server completion
	s.serverDone = make(chan struct{})

	// Handle connections in a goroutine
	go func() {
		defer close(s.serverDone)
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				return // listener closed
			}

			if conn != nil {
				s.Require().NoError(conn.Close())
			}
		}
	}()

	// Try to set up IPv6 listener if supported
	conn, err := net.Dial("udp", "[::1]:1")
	if err == nil {
		s.Require().NoError(conn.Close())
		s.ipv6Listener, err = net.Listen("tcp", "[::1]:0")
		if err != nil {
			s.T().Log("IPv6 listener setup failed:", err)
		} else {
			// Handle IPv6 connections
			go func() {
				for {
					conn, err := s.ipv6Listener.Accept()
					if err != nil {
						return // listener closed
					}
					if conn != nil {
						s.Require().NoError(conn.Close())
					}
				}
			}()
		}
	}
}

// TearDownSuite tears down test suite resources
func (s *TCPSuite) TearDownSuite() {
	// Close listeners
	if s.listener != nil {
		s.Require().NoError(s.listener.Close())
		<-s.serverDone // Wait for server goroutine to complete
	}

	if s.ipv6Listener != nil {
		s.Require().NoError(s.ipv6Listener.Close())
	}
}

// TestNew checks the constructor with default and custom options
func (s *TCPSuite) TestNew() {
	// Test default values
	tc := New("127.0.0.1:8080").(*TCP)
	s.Equal("127.0.0.1:8080", tc.address)
	s.Equal(DefaultConnectionTimeout, tc.timeout)

	// Test with options
	customTimeout := 5 * time.Second
	tc = New("127.0.0.1:8080", WithTimeout(customTimeout)).(*TCP)
	s.Equal("127.0.0.1:8080", tc.address)
	s.Equal(customTimeout, tc.timeout)
}

// TestWithTimeout tests the timeout option
func (s *TCPSuite) TestWithTimeout() {
	tc := &TCP{timeout: DefaultConnectionTimeout}
	opt := WithTimeout(10 * time.Second)
	opt(tc)
	s.Equal(10*time.Second, tc.timeout)
}

// TestIdentity tests the Identity method
func (s *TCPSuite) TestIdentity() {
	address := "127.0.0.1:8080"
	tc := New(address)
	identity, err := tc.Identity()
	s.NoError(err)
	s.Equal(address, identity)
}

// TestCheckSuccessful tests successful TCP connection
func (s *TCPSuite) TestCheckSuccessful() {
	tc := New(s.listener.Addr().String())
	err := tc.Check(context.Background())
	s.NoError(err)
}

// TestCheckConnectionRefused tests connection refused error
func (s *TCPSuite) TestCheckConnectionRefused() {
	address := fmt.Sprintf("127.0.0.1:%d", s.unusedPort)
	tc := New(address, WithTimeout(500*time.Millisecond))
	err := tc.Check(context.Background())

	// The error should be an ExpectedError
	s.Error(err)

	var expectedErr *checker.ExpectedError
	s.True(errors.As(err, &expectedErr))
	s.Contains(expectedErr.Error(), "failed to establish a tcp connection")
}

// TestCheckInvalidAddress tests invalid address format
func (s *TCPSuite) TestCheckInvalidAddress() {
	tc := New("invalid-address")
	err := tc.Check(context.Background())

	// This should be a generic error, not an ExpectedError
	s.Error(err)

	var expectedErr *checker.ExpectedError
	s.True(errors.As(err, &expectedErr))
}

// TestCheckTimeout tests timeout behavior
func (s *TCPSuite) TestCheckTimeout() {
	// Use a black-hole IP that will cause timeout
	tc := New("240.0.0.1:12345", WithTimeout(500*time.Millisecond))

	start := time.Now()
	err := tc.Check(context.Background())
	elapsed := time.Since(start)

	// Verify the timeout was respected
	s.Error(err)
	s.True(elapsed >= 500*time.Millisecond, "Timeout was not respected")

	// Check error type and details
	var expectedErr *checker.ExpectedError
	if s.True(errors.As(err, &expectedErr)) {
		s.Contains(expectedErr.Error(), "timed out while making a tcp call")

		details := expectedErr.Details()
		s.Equal("timeout", details[0])
		s.Equal(500*time.Millisecond, details[1])
	}
}

// TestCheckContextCancellation tests context cancellation
func (s *TCPSuite) TestCheckContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Use a black-hole IP to ensure the operation would take time
	tc := New("240.0.0.1:12345", WithTimeout(10*time.Second))

	// Cancel the context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := tc.Check(ctx)
	elapsed := time.Since(start)

	s.Error(err)
	s.True(elapsed < 5*time.Second, "Context cancellation was not respected")
	s.ErrorIs(err, context.Canceled)
}

// TestCheckNameResolution tests name resolution errors
func (s *TCPSuite) TestCheckNameResolution() {
	tc := New("non-existent-domain.example:12345", WithTimeout(500*time.Millisecond))
	err := tc.Check(context.Background())

	s.Error(err)

	var expectedErr *checker.ExpectedError
	s.True(errors.As(err, &expectedErr), "Name resolution error should be wrapped as an ExpectedError")
}

// TestCheckIPv6Address tests IPv6 support
func (s *TCPSuite) TestCheckIPv6Address() {
	if s.ipv6Listener == nil {
		s.T().Skip("IPv6 not available on this system")
	}

	tc := New(s.ipv6Listener.Addr().String())
	err := tc.Check(context.Background())
	s.NoError(err, "Should be able to connect to IPv6 address")
}

// TestTableDriven defines table-driven tests for various scenarios
func (s *TCPSuite) TestTableDriven() {
	// Create a context with a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Define test cases
	tests := []struct {
		name        string
		address     string
		timeout     time.Duration
		ctx         context.Context
		shouldError bool
		errorType   string // "expected", "other", or "" if no error
	}{
		{
			name:        "Valid Address",
			address:     fmt.Sprintf("127.0.0.1:%d", s.port),
			timeout:     1 * time.Second,
			ctx:         ctx,
			shouldError: false,
		},
		{
			name:        "Connection Refused",
			address:     fmt.Sprintf("127.0.0.1:%d", s.unusedPort),
			timeout:     1 * time.Second,
			ctx:         ctx,
			shouldError: true,
			errorType:   "expected", // ExpectedError
		},
		{
			name:        "Very Short Timeout",
			address:     "240.0.0.1:12345",    // non-routable IP, will time out
			timeout:     1 * time.Millisecond, // ultra short timeout
			ctx:         ctx,
			shouldError: true,
			errorType:   "expected", // ExpectedError for timeout
		},
		{
			name:        "Invalid Address Format",
			address:     "not-a-valid-address",
			timeout:     1 * time.Second,
			ctx:         ctx,
			shouldError: true,
			errorType:   "expected",
		},
	}

	// Run all test cases
	for _, tt := range tests {
		s.Run(tt.name, func() {
			tc := New(tt.address, WithTimeout(tt.timeout))
			err := tc.Check(tt.ctx)

			if tt.shouldError {
				s.Error(err)

				var expectedErr *checker.ExpectedError
				isExpectedErr := errors.As(err, &expectedErr)

				if tt.errorType == "expected" {
					s.True(isExpectedErr, "Expected an ExpectedError but got: %v", err)
				} else if tt.errorType == "other" {
					s.False(isExpectedErr, "Expected a non-ExpectedError")
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

// Helper method to get a cancelled context
func (s *TCPSuite) getCancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	return ctx
}

// TestTCPSuite runs the test suite
func TestTCPSuite(t *testing.T) {
	suite.Run(t, new(TCPSuite))
}
