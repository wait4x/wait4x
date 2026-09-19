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

// Package exec provides the Exec checker for the Wait4X application.
package exec

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"wait4x.dev/v3/checker"
)

// Option configures an Exec checker
type Option func(e *Exec)

// Exec is a command execution checker
type Exec struct {
	command           string
	args              []string
	expectExitCode    int
	expectStdoutRegex string
	expectStderrRegex string
}

// New creates a new Exec checker
func New(command string, opts ...Option) checker.Checker {
	e := &Exec{
		command:        command,
		expectExitCode: 0, // Default to expecting exit code 0
	}

	// apply the list of options to Exec
	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithArgs configures command arguments
func WithArgs(args []string) Option {
	return func(e *Exec) {
		e.args = args
	}
}

// WithExpectExitCode configures expected exit code
func WithExpectExitCode(code int) Option {
	return func(e *Exec) {
		e.expectExitCode = code
	}
}

// WithExpectStdoutRegex configures the regular expression
// which is evaluated against STDOUT.
func WithExpectStdoutRegex(pattern string) Option {
	return func(e *Exec) {
		e.expectStdoutRegex = pattern
	}
}

// WithExpectStderrRegex configures the regular expression
// which is evaluated against STDERR.
func WithExpectStderrRegex(pattern string) Option {
	return func(e *Exec) {
		e.expectStderrRegex = pattern
	}
}

// Identity returns the identity of the checker
func (e *Exec) Identity() (string, error) {
	if len(e.args) > 0 {
		return fmt.Sprintf("%s %s", e.command, strings.Join(e.args, " ")), nil
	}
	return e.command, nil
}

// Check executes the command and checks if it returns the expected exit code
func (e *Exec) Check(ctx context.Context) error {
	var stdout, stderr *pipeProcessor
	var errs chan error
	var err error

	// Create command with context for cancellation support
	cmd := exec.CommandContext(ctx, e.command, e.args...)

	stdout, err = newPipeProcessor(e.expectStdoutRegex, cmd.StdoutPipe)
	if err != nil {
		return checker.NewExpectedError(
			"failed to prepare STDOUT processor", err,
			"command", e.command,
			"args", e.args,
		)
	}

	stderr, err = newPipeProcessor(e.expectStderrRegex, cmd.StderrPipe)
	if err != nil {
		return checker.NewExpectedError(
			"failed to prepare STDERR processor", err,
			"command", e.command,
			"args", e.args,
		)
	}

	if err := cmd.Start(); err != nil {
		return checker.NewExpectedError(
			"unable to start command", err,
			"command", e.command,
			"args", e.args,
		)
	}

	if stdout != nil || stderr != nil {
		var wg sync.WaitGroup
		errs = make(chan error)

		go func() {
			// wait for the workers to finish processing all of the data
			// before closing the channel, otherwise they will be blocked
			wg.Wait()
			close(errs)
		}()

		if stdout != nil {
			stdout.Process(&wg, errs)
		}

		if stderr != nil {
			stderr.Process(&wg, errs)
		}
	}

	err = cmd.Wait()

	// Check if context was canceled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue with exit code checking
	}

	// Handle the command execution result
	exitCode := 0
	if err != nil {
		// If there's an error, try to get the exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// This is not an exit error but some other error (like command not found)
			return checker.NewExpectedError(
				"failed to execute command", err,
				"command", e.command,
				"args", e.args,
			)
		}
	}

	// Check if the exit code matches the expected one
	if exitCode != e.expectExitCode {
		return checker.NewExpectedError(
			"command exited with unexpected code", nil,
			"command", e.command,
			"args", e.args,
			"actual", exitCode,
			"expect", e.expectExitCode,
		)
	}

	return drainErrors(errs)
}

// drainErrors reads all items from the given channel
// and returns the last non-nil value or nil if the
// channel is nil or does not contain any errors.
func drainErrors(errs chan error) (err error) {
	if errs == nil {
		return
	}

	for e := range errs {
		if e != nil {
			err = e
		}
	}

	return
}
