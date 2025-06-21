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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"wait4x.dev/v3/internal/test"
)

// TestMain is the main function for the TCP command tests
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestTcpCommandInvalidArgument tests the TCP command with invalid arguments
func TestTcpCommandInvalidArgument(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewTCPCommand())

	_, err := test.ExecuteCommand(rootCmd, "tcp")

	assert.Equal(t, "ADDRESS is required argument for the tcp command", err.Error())
}

// TestTcpConnectionSuccess tests the TCP connection success
func TestTcpConnectionSuccess(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewTCPCommand())

	_, err := test.ExecuteCommand(rootCmd, "tcp", "1.1.1.1:53")

	assert.Nil(t, err)
}

// TestTcpConnectionFail tests the TCP connection failure
func TestTcpConnectionFail(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewTCPCommand())

	_, err := test.ExecuteCommand(rootCmd, "tcp", "127.0.0.1:8080", "-t", "2s")

	assert.Equal(t, context.DeadlineExceeded, err)
}
