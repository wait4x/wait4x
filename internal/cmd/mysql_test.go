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
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"wait4x.dev/v4/internal/test"
)

// MySQLCommandSuite is a test suite for MySQL command functionality
type MySQLCommandSuite struct {
	suite.Suite

	// Shared resources for the test suite
	rootCmd          *cobra.Command
	mysqlCmd         *cobra.Command
	container        *mysql.MySQLContainer
	connectionString string
}

// SetupSuite sets up test suite resources
func (s *MySQLCommandSuite) SetupSuite() {
	// Set up a MySQL container for tests that need an active connection
	var err error
	s.container, err = mysql.Run(
		context.Background(),
		"mysql:8.0.36",
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)
	s.Require().NoError(err)

	// Get the connection string for testing
	s.connectionString, err = s.container.ConnectionString(context.Background())
	s.Require().NoError(err)
}

// TearDownSuite tears down test suite resources
func (s *MySQLCommandSuite) TearDownSuite() {
	// Close container
	if s.container != nil {
		err := s.container.Terminate(context.Background())
		s.Require().NoError(err)
	}
}

// SetupTest sets up each test
func (s *MySQLCommandSuite) SetupTest() {
	s.rootCmd = NewRootCommand()
	s.mysqlCmd = NewMysqlCommand()
	s.rootCmd.AddCommand(s.mysqlCmd)
}

// TestNewMysqlCommand tests the MySQL command creation
func (s *MySQLCommandSuite) TestNewMysqlCommand() {
	cmd := NewMysqlCommand()

	s.Equal("mysql DSN... [flags] [-- command [args...]]", cmd.Use)
	s.Equal("Check MySQL connection", cmd.Short)
	s.NotNil(cmd.Example)
	s.Contains(cmd.Example, "wait4x mysql user:password@tcp(localhost:5555)/dbname?tls=skip-verify")
	s.Contains(cmd.Example, "wait4x mysql username:password@unix(/tmp/mysql.sock)/myDatabase")
}

// TestMysqlCommandValidation tests argument validation
func (s *MySQLCommandSuite) TestMysqlCommandValidation() {
	// Test missing arguments
	_, err := test.ExecuteCommand(s.rootCmd, "mysql")
	s.Require().Error(err)
	s.Equal("DSN is required argument for the mysql command", err.Error())

	// Test empty arguments
	err = s.mysqlCmd.Args(s.mysqlCmd, []string{})
	s.Require().Error(err)
	s.Equal("DSN is required argument for the mysql command", err.Error())

	// Test valid arguments
	err = s.mysqlCmd.Args(s.mysqlCmd, []string{"user:password@tcp(localhost:3306)/dbname"})
	s.NoError(err)

	err = s.mysqlCmd.Args(s.mysqlCmd, []string{
		"user:password@tcp(localhost:3306)/dbname",
		"user2:password2@tcp(localhost:3307)/dbname2",
	})
	s.NoError(err)
}

// TestMysqlConnectionScenarios tests various connection scenarios
func (s *MySQLCommandSuite) TestMysqlConnectionScenarios() {
	// Test successful connection
	_, err := test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString)
	s.NoError(err)

	// Test connection failure (timeout)
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", "user:password@tcp(localhost:8080)/dbname", "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)

	// Test connection timeout with black-hole IP
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", "user:password@tcp(240.0.0.1:3306)/dbname", "-t", "1s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)

	// Test multiple DSNs (mixed valid/invalid)
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "user:password@tcp(localhost:8080)/dbname", "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestMysqlCommandFlags tests all command flags
func (s *MySQLCommandSuite) TestMysqlCommandFlags() {
	// Test interval flag
	_, err := test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "-i", "500ms")
	s.NoError(err)

	// Test timeout flag
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "-t", "0")
	s.NoError(err)

	// Test quiet mode
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "-q")
	s.NoError(err)

	// Test no color flag
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "--no-color")
	s.NoError(err)

	// Test dash separator for command execution
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "--", "echo", "success")
	s.NoError(err)
}

// TestMysqlCommandInvertCheck tests invert check functionality
func (s *MySQLCommandSuite) TestMysqlCommandInvertCheck() {
	// With invert check, a successful connection should fail
	_, err := test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "-v", "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)

	// With invert check, a failed connection should succeed
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", "user:password@tcp(localhost:8080)/dbname", "-v", "-t", "2s")
	s.NoError(err)
}

// TestMysqlCommandBackoffPolicy tests backoff policy functionality
func (s *MySQLCommandSuite) TestMysqlCommandBackoffPolicy() {
	// Test exponential backoff
	_, err := test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "--backoff-policy", "exponential")
	s.NoError(err)

	// Test invalid backoff policy
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "--backoff-policy", "invalid")
	s.Error(err)
	s.Contains(err.Error(), "--backoff-policy must be one of")

	// Test backoff coefficient
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "--backoff-policy", "exponential", "--backoff-exponential-coefficient", "1.5")
	s.NoError(err)

	// Test backoff max interval
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "--backoff-policy", "exponential", "--backoff-exponential-max-interval", "3s")
	s.NoError(err)

	// Test invalid backoff max interval
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", s.connectionString, "--backoff-policy", "exponential", "--backoff-exponential-max-interval", "100ms", "-i", "200ms")
	s.Require().Error(err)
	s.Contains(err.Error(), "--backoff-exponential-max-interval must be greater than --interval")
}

// TestMysqlCommandDSNFormats tests various DSN formats
func (s *MySQLCommandSuite) TestMysqlCommandDSNFormats() {
	// Test invalid DSN format
	_, err := test.ExecuteCommand(s.rootCmd, "mysql", "invalid-dsn-format", "-t", "2s")
	s.Require().Error(err)
	s.Contains(err.Error(), "can't retrieve the checker identity")

	// Test Unix socket DSN (will timeout)
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", "user:password@unix(/tmp/mysql.sock)/dbname", "-t", "2s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)

	// Test TLS DSN (will timeout)
	_, err = test.ExecuteCommand(s.rootCmd, "mysql", "user:password@tcp(240.0.0.1:3306)/dbname?tls=skip-verify", "-t", "1s")
	s.Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestMysqlCommandTableDriven tests the MySQL command with various scenarios using table-driven tests
func (s *MySQLCommandSuite) TestMysqlCommandTableDriven() {
	tests := []struct {
		name         string
		args         []string
		shouldError  bool
		errorType    string
		errorMessage string
	}{
		{
			name:        "valid connection",
			args:        []string{"mysql", s.connectionString},
			shouldError: false,
		},
		{
			name:         "invalid DSN format",
			args:         []string{"mysql", "invalid-dsn"},
			shouldError:  true,
			errorType:    "validation",
			errorMessage: "can't retrieve the checker identity",
		},
		{
			name:         "connection timeout",
			args:         []string{"mysql", "user:password@tcp(240.0.0.1:3306)/dbname", "-t", "1s"},
			shouldError:  true,
			errorType:    "timeout",
			errorMessage: "",
		},
		{
			name:         "missing DSN argument",
			args:         []string{"mysql"},
			shouldError:  true,
			errorType:    "validation",
			errorMessage: "DSN is required argument for the mysql command",
		},
		{
			name:        "multiple valid DSNs",
			args:        []string{"mysql", s.connectionString, s.connectionString},
			shouldError: false,
		},
		{
			name:         "mixed valid and invalid DSNs",
			args:         []string{"mysql", s.connectionString, "user:password@tcp(localhost:8080)/dbname", "-t", "2s"},
			shouldError:  true,
			errorType:    "timeout",
			errorMessage: "",
		},
		{
			name:        "with custom interval",
			args:        []string{"mysql", s.connectionString, "-i", "500ms"},
			shouldError: false,
		},
		{
			name:        "with exponential backoff",
			args:        []string{"mysql", s.connectionString, "--backoff-policy", "exponential"},
			shouldError: false,
		},
		{
			name:         "with invalid backoff policy",
			args:         []string{"mysql", s.connectionString, "--backoff-policy", "invalid"},
			shouldError:  true,
			errorType:    "validation",
			errorMessage: "--backoff-policy must be one of",
		},
		{
			name:         "with invert check on valid connection",
			args:         []string{"mysql", s.connectionString, "-v", "-t", "2s"},
			shouldError:  true,
			errorType:    "timeout",
			errorMessage: "",
		},
		{
			name:        "with invert check on invalid connection",
			args:        []string{"mysql", "user:password@tcp(localhost:8080)/dbname", "-v", "-t", "2s"},
			shouldError: false,
		},
		{
			name:        "with quiet mode",
			args:        []string{"mysql", s.connectionString, "-q"},
			shouldError: false,
		},
		{
			name:        "with no color",
			args:        []string{"mysql", s.connectionString, "--no-color"},
			shouldError: false,
		},
		{
			name:        "with zero timeout",
			args:        []string{"mysql", s.connectionString, "-t", "0"},
			shouldError: false,
		},
		{
			name:        "with command execution",
			args:        []string{"mysql", s.connectionString, "--", "echo", "success"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create a fresh root command for each test to avoid flag pollution
			rootCmd := NewRootCommand()
			mysqlCmd := NewMysqlCommand()
			rootCmd.AddCommand(mysqlCmd)

			_, err := test.ExecuteCommand(rootCmd, tt.args...)

			if tt.shouldError {
				s.Require().Error(err)
				if tt.errorType == "timeout" {
					s.Require().ErrorIs(err, context.DeadlineExceeded)
				} else if tt.errorType == "validation" {
					s.Require().ErrorContains(err, tt.errorMessage)
				} else if tt.errorType == "connection" {
					s.Require().ErrorContains(err, tt.errorMessage)
				}
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestMysqlCommandHelp tests the MySQL command help
func (s *MySQLCommandSuite) TestMysqlCommandHelp() {
	output, err := test.ExecuteCommand(s.rootCmd, "mysql", "--help")
	s.Require().NoError(err)
	s.Contains(output, "Check MySQL connection")
	s.Contains(output, "DSN")
}

// TestMysqlCommandExample tests the MySQL command example
func (s *MySQLCommandSuite) TestMysqlCommandExample() {
	// The example should be present in the command
	s.Contains(s.mysqlCmd.Example, "wait4x mysql user:password@tcp(localhost:5555)/dbname?tls=skip-verify")
	s.Contains(s.mysqlCmd.Example, "wait4x mysql username:password@unix(/tmp/mysql.sock)/myDatabase")
}

// TestMySQLCommandSuite runs the MySQL command test suite
func TestMySQLCommandSuite(t *testing.T) {
	suite.Run(t, new(MySQLCommandSuite))
}
