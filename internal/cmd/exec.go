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
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker/exec"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewExecCommand creates a new exec sub-command
func NewExecCommand() *cobra.Command {
	execCommand := &cobra.Command{
		Use:   "exec COMMAND [ARGS...] [flags]",
		Short: "Check command execution",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("COMMAND is required argument for the exec command")
			}
			return nil
		},
		Example: `
  # Wait for a command to exit with code 0
  wait4x exec "ls /tmp"

  # Wait for a command to exit with a specific code
  wait4x exec "ls /nonexistent" --exit-code 2

  # Enable exponential backoff retry
  wait4x exec "bash ./some-script.sh" --exit-code 0 --backoff-policy exponential --backoff-exponential-max-interval 120s --timeout 120s`,
		RunE: runExec,
	}

	execCommand.Flags().Int("exit-code", 0, "Expected exit code from the command")

	return execCommand
}

// runExec runs the exec command
func runExec(cmd *cobra.Command, args []string) error {
	exitCode, _ := cmd.Flags().GetInt("exit-code")

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	commandParts := strings.Fields(args[0])
	command := commandParts[0]
	var commandArgs []string

	if len(commandParts) > 1 {
		commandArgs = append(commandArgs, commandParts[1:]...)
	}

	if len(args) > 1 {
		commandArgs = append(commandArgs, args[1:]...)
	}

	checker := exec.New(command,
		exec.WithArgs(commandArgs),
		exec.WithExpectExitCode(exitCode),
	)

	return waiter.WaitContext(
		cmd.Context(),
		checker,
		waiter.WithTimeout(contextutil.GetTimeout(cmd.Context())),
		waiter.WithInterval(contextutil.GetInterval(cmd.Context())),
		waiter.WithInvertCheck(contextutil.GetInvertCheck(cmd.Context())),
		waiter.WithBackoffPolicy(contextutil.GetBackoffPolicy(cmd.Context())),
		waiter.WithBackoffCoefficient(contextutil.GetBackoffCoefficient(cmd.Context())),
		waiter.WithBackoffExponentialMaxInterval(
			contextutil.GetBackoffExponentialMaxInterval(cmd.Context()),
		),
		waiter.WithLogger(logger),
	)
}
