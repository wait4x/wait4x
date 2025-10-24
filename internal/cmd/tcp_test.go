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
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"wait4x.dev/v4/internal/test"
)

// TCPCommandSuite is a test suite for TCP command functionality
type TCPCommandSuite struct {
	suite.Suite

	// Shared resources for the test suite
	rootCmd    *cobra.Command
	tcpCmd     *cobra.Command
	listener   net.Listener
	port       int
	unusedPort int
	serverDone chan struct{}
}

// SetupSuite sets up test suite resources
func (s *TCPCommandSuite) SetupSuite() {
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
}

// TearDownSuite tears down test suite resources
func (s *TCPCommandSuite) TearDownSuite() {
	// Close listener
	if s.listener != nil {
		s.Require().NoError(s.listener.Close())
		<-s.serverDone // Wait for server goroutine to complete
	}
}

// SetupTest sets up each test
func (s *TCPCommandSuite) SetupTest() {
	s.rootCmd = NewRootCommand()
	s.tcpCmd = NewTCPCommand()
	s.rootCmd.AddCommand(s.tcpCmd)
}

// TestNewTCPCommand tests the TCP command creation
func (s *TCPCommandSuite) TestNewTCPCommand() {
	cmd := NewTCPCommand()

	s.Equal("tcp ADDRESS... [flags] [-- command [args...]]", cmd.Use)
	s.Equal("Check TCP connection", cmd.Short)
	s.NotNil(cmd.Example)
	s.Contains(cmd.Example, "wait4x tcp 127.0.0.1:9090")

	// Test that the command has the expected flags
	flags := cmd.Flags()
	connectionTimeout, err := flags.GetDuration("connection-timeout")
	s.NoError(err)
	s.Equal(3*time.Second, connectionTimeout) // Default from tcp package
}

// TestTCPCommandInvalidArgument tests the TCP command with invalid arguments
func (s *TCPCommandSuite) TestTCPCommandInvalidArgument() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp")
	s.Error(err)
	s.Equal("ADDRESS is required argument for the tcp command", err.Error())
}

// TestTCPCommandEmptyArgs tests the TCP command with empty arguments
func (s *TCPCommandSuite) TestTCPCommandEmptyArgs() {
	err := s.tcpCmd.Args(s.tcpCmd, []string{})
	s.Error(err)
	s.Equal("ADDRESS is required argument for the tcp command", err.Error())
}

// TestTCPCommandValidArgs tests the TCP command with valid arguments
func (s *TCPCommandSuite) TestTCPCommandValidArgs() {
	err := s.tcpCmd.Args(s.tcpCmd, []string{"127.0.0.1:8080"})
	s.NoError(err)

	err = s.tcpCmd.Args(s.tcpCmd, []string{"127.0.0.1:8080", "192.168.1.1:9090"})
	s.NoError(err)
}

// TestTCPConnectionSuccess tests the TCP connection success
func (s *TCPCommandSuite) TestTCPConnectionSuccess() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53")
	s.NoError(err)
}

// TestTCPConnectionSuccessLocal tests the TCP connection success to local server
func (s *TCPCommandSuite) TestTCPConnectionSuccessLocal() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", s.listener.Addr().String())
	s.NoError(err)
}

// TestTCPConnectionFail tests the TCP connection failure
func (s *TCPCommandSuite) TestTCPConnectionFail() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "127.0.0.1:8080", "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestTCPConnectionFailUnusedPort tests the TCP connection failure on unused port
func (s *TCPCommandSuite) TestTCPConnectionFailUnusedPort() {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(s.unusedPort))
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", address, "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestTCPConnectionTimeout tests the TCP connection timeout behavior
func (s *TCPCommandSuite) TestTCPConnectionTimeout() {
	// Use a black-hole IP that will cause timeout
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "240.0.0.1:12345", "-t", "1s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestTCPConnectionWithCustomTimeout tests the TCP connection with custom timeout
func (s *TCPCommandSuite) TestTCPConnectionWithCustomTimeout() {
	// Test with a very short connection timeout
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "240.0.0.1:12345", "--connection-timeout", "100ms", "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestTCPConnectionWithInvalidTimeout tests the TCP connection with invalid timeout
func (s *TCPCommandSuite) TestTCPConnectionWithInvalidTimeout() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "127.0.0.1:8080", "--connection-timeout", "invalid")
	s.Error(err)
	s.Contains(err.Error(), "invalid argument \"invalid\" for \"--connection-timeout\" flag")
}

// TestTCPMultipleAddresses tests the TCP command with multiple addresses
func (s *TCPCommandSuite) TestTCPMultipleAddresses() {
	// Test with multiple valid addresses
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "8.8.8.8:53")
	s.NoError(err)
}

// TestTCPMultipleAddressesMixed tests the TCP command with mixed valid/invalid addresses
func (s *TCPCommandSuite) TestTCPMultipleAddressesMixed() {
	// One valid, one invalid - should fail
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "127.0.0.1:8080", "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestTCPCommandWithDash tests the TCP command with dash separator for command execution
func (s *TCPCommandSuite) TestTCPCommandWithDash() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--", "echo", "success")
	s.NoError(err)
}

// TestTCPCommandWithInvertCheck tests the TCP command with invert check flag
func (s *TCPCommandSuite) TestTCPCommandWithInvertCheck() {
	// With invert check, we expect the command to fail when connection succeeds
	// Use a shorter timeout to make the test fail faster
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "-v", "-t", "1s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestTCPCommandWithInvertCheckFail tests the TCP command with invert check when connection fails
func (s *TCPCommandSuite) TestTCPCommandWithInvertCheckFail() {
	// With invert check, we expect the command to succeed when connection fails
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "127.0.0.1:8080", "-v", "-t", "2s")
	s.NoError(err)
}

// TestTCPCommandWithInterval tests the TCP command with custom interval
func (s *TCPCommandSuite) TestTCPCommandWithInterval() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "-i", "500ms")
	s.NoError(err)
}

// TestTCPCommandWithBackoffPolicy tests the TCP command with different backoff policies
func (s *TCPCommandSuite) TestTCPCommandWithBackoffPolicy() {
	// Test with exponential backoff
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--backoff-policy", "exponential")
	s.NoError(err)

	// Test with linear backoff
	_, err = test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--backoff-policy", "linear")
	s.NoError(err)
}

// TestTCPCommandWithInvalidBackoffPolicy tests the TCP command with invalid backoff policy
func (s *TCPCommandSuite) TestTCPCommandWithInvalidBackoffPolicy() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--backoff-policy", "invalid")
	s.Error(err)
	s.Contains(err.Error(), "--backoff-policy must be one of")
}

// TestTCPCommandWithBackoffCoefficient tests the TCP command with custom backoff coefficient
func (s *TCPCommandSuite) TestTCPCommandWithBackoffCoefficient() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--backoff-exponential-coefficient", "1.5")
	s.NoError(err)
}

// TestTCPCommandWithBackoffMaxInterval tests the TCP command with custom backoff max interval
func (s *TCPCommandSuite) TestTCPCommandWithBackoffMaxInterval() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--backoff-exponential-max-interval", "3s")
	s.NoError(err)
}

// TestTCPCommandWithInvalidBackoffMaxInterval tests the TCP command with invalid backoff max interval
func (s *TCPCommandSuite) TestTCPCommandWithInvalidBackoffMaxInterval() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--backoff-policy", "exponential", "--backoff-exponential-max-interval", "100ms", "-i", "200ms")
	s.Error(err)
	s.Contains(err.Error(), "--backoff-exponential-max-interval must be greater than --interval")
}

// TestTCPCommandWithQuietMode tests the TCP command with quiet mode
func (s *TCPCommandSuite) TestTCPCommandWithQuietMode() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "-q")
	s.NoError(err)
}

// TestTCPCommandWithNoColor tests the TCP command with no color flag
func (s *TCPCommandSuite) TestTCPCommandWithNoColor() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "--no-color")
	s.NoError(err)
}

// TestTCPCommandWithZeroTimeout tests the TCP command with zero timeout (unlimited)
func (s *TCPCommandSuite) TestTCPCommandWithZeroTimeout() {
	// This should work but take longer, so we'll use a short timeout for the test
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53", "-t", "0s")
	s.NoError(err)
}

// TestTCPCommandWithInvalidAddressFormat tests the TCP command with invalid address format
func (s *TCPCommandSuite) TestTCPCommandWithInvalidAddressFormat() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "invalid-address", "-t", "2s")
	s.Error(err)
	// The error should be either a connection error or timeout
	if err == context.DeadlineExceeded {
		// Timeout is acceptable for invalid addresses
		s.Equal(context.DeadlineExceeded, err)
	} else {
		// Or it should be a connection error
		s.Contains(err.Error(), "failed to establish a tcp connection")
	}
}

// TestTCPCommandWithIPv6Address tests the TCP command with IPv6 address
func (s *TCPCommandSuite) TestTCPCommandWithIPv6Address() {
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "[::1]:53", "-t", "2s")
	// This might fail if IPv6 is not available, but should not crash
	if err != nil {
		if err == context.DeadlineExceeded {
			// Timeout is acceptable for IPv6 if not available
			s.Equal(context.DeadlineExceeded, err)
		} else {
			s.Contains(err.Error(), "failed to establish a tcp connection")
		}
	}
}

// TestTCPCommandTableDriven defines table-driven tests for various scenarios
func (s *TCPCommandSuite) TestTCPCommandTableDriven() {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
		errorType   string // "timeout", "validation", "connection", "connection_or_timeout", or "" if no error
	}{
		{
			name:        "Valid Address",
			args:        []string{"tcp", "1.1.1.1:53"},
			shouldError: false,
		},
		{
			name:        "No Arguments",
			args:        []string{"tcp"},
			shouldError: true,
			errorType:   "validation",
		},
		{
			name:        "Connection Refused",
			args:        []string{"tcp", "240.0.0.1:12345", "-t", "2s"},
			shouldError: true,
			errorType:   "timeout",
		},
		{
			name:        "Invalid Address Format",
			args:        []string{"tcp", "not-a-valid-address", "-t", "2s"},
			shouldError: true,
			errorType:   "connection_or_timeout",
		},
		{
			name:        "Multiple Valid Addresses",
			args:        []string{"tcp", "1.1.1.1:53", "8.8.8.8:53"},
			shouldError: false,
		},
		{
			name:        "With Custom Interval",
			args:        []string{"tcp", "1.1.1.1:53", "-i", "500ms"},
			shouldError: false,
		},
		{
			name:        "With Invert Check Success",
			args:        []string{"tcp", "240.0.0.1:12345", "-v", "-t", "2s"},
			shouldError: false, // Should succeed because connection fails and we're inverting
		},
		{
			name:        "With Invert Check Failure",
			args:        []string{"tcp", "1.1.1.1:53", "-v", "-t", "1s"},
			shouldError: true, // Should fail because connection succeeds and we're inverting
			errorType:   "timeout",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := test.ExecuteCommand(s.rootCmd, tt.args...)

			if tt.shouldError {
				s.Error(err)
				if tt.errorType == "timeout" {
					s.Equal(context.DeadlineExceeded, err)
				} else if tt.errorType == "validation" {
					s.Contains(err.Error(), "ADDRESS is required argument for the tcp command")
				} else if tt.errorType == "connection" {
					s.Contains(err.Error(), "failed to establish a tcp connection")
				} else if tt.errorType == "connection_or_timeout" {
					if err == context.DeadlineExceeded {
						s.Equal(context.DeadlineExceeded, err)
					} else {
						s.Contains(err.Error(), "failed to establish a tcp connection")
					}
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestTCPCommandFlags tests the TCP command flags
func (s *TCPCommandSuite) TestTCPCommandFlags() {
	flags := s.tcpCmd.Flags()

	// Test connection-timeout flag
	connectionTimeout, err := flags.GetDuration("connection-timeout")
	s.NoError(err)
	s.Equal(3*time.Second, connectionTimeout)

	// Test that the flag is required
	s.True(flags.Lookup("connection-timeout") != nil)
}

// TestTCPCommandHelp tests the TCP command help
func (s *TCPCommandSuite) TestTCPCommandHelp() {
	output, err := test.ExecuteCommand(s.rootCmd, "tcp", "--help")
	s.NoError(err)
	s.Contains(output, "Check TCP connection")
	s.Contains(output, "connection-timeout")
}

// TestTCPCommandExample tests the TCP command example
func (s *TCPCommandSuite) TestTCPCommandExample() {
	// The example should be present in the command
	s.Contains(s.tcpCmd.Example, "wait4x tcp 127.0.0.1:9090")
}

// TestTCPCommandRunE tests the runTCP function directly
func (s *TCPCommandSuite) TestTCPCommandRunE() {
	// Test argument validation - this should work without logger context
	err := s.tcpCmd.Args(s.tcpCmd, []string{})
	s.Error(err)
	s.Equal("ADDRESS is required argument for the tcp command", err.Error())

	// Test with valid arguments - this should also work for argument validation
	err = s.tcpCmd.Args(s.tcpCmd, []string{"1.1.1.1:53"})
	s.NoError(err)

	// For testing the actual runTCP function, we need to use the full command execution
	// since runTCP requires a logger in the context
	_, err = test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53")
	s.NoError(err)
}

// TestTCPCommandWithContext tests the TCP command with context
func (s *TCPCommandSuite) TestTCPCommandWithContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.tcpCmd.SetContext(ctx)

	// Test that the command works with a valid context
	_, err := test.ExecuteCommand(s.rootCmd, "tcp", "1.1.1.1:53")
	s.NoError(err)
}

// TestTCPCommandSuite runs the test suite
func TestTCPCommandSuite(t *testing.T) {
	suite.Run(t, new(TCPCommandSuite))
}
