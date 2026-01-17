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

	"github.com/getsentry/size"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker/file"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewFileCommand creates a new file sub-command
func NewFileCommand() *cobra.Command {
	var defaultSize sizeValue
	fileCommand := &cobra.Command{
		Use:   "file FILE... [flags]",
		Short: "Check file states",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("File paths are required arguments for the file command")
			}
			return nil
		},
		Example: `
  # Wait for a file to appear
  wait4x file /run/mail/daemon.lock

  # Wait for a file to disappear
  wait4x file /var/spool/cron.lock --invert-check

  # Wait for any file to appear
  wait4x file /dev/vdb /dev/vdc /dev/vdd --allow-any

  # Wait for a file content to match a regular expression
  wait4x file /run/observer/health.json --expect-content-regex healthy

  # Wait for a file to grow above a certain size
  wait4x file /mnt/backups/daily.tgz --expect-size 3M --invert-check

  # Wait for a file to be rotated (i.e. modification time is below a certain threshold)
  wait4x file /var/log/server.log.1.gz --expect-age 5s`,
		RunE: runFile,
	}

	fileCommand.Flags().Bool("allow-any", false, "Only a single file must match the provided conditions")
	fileCommand.Flags().String("expect-content-regex", "", "Expected pattern to match against the file content")
	fileCommand.Flags().Var(&defaultSize, "expect-size", "Maximum size the file must no exceed")
	fileCommand.Flags().Duration("expect-age", 0, "Maximum age the file must not exceed")

	return fileCommand
}

// runFile runs the file command
func runFile(cmd *cobra.Command, args []string) error {
	allowAny, _ := cmd.Flags().GetBool("allow-any")
	expectRegex, _ := cmd.Flags().GetString("expect-content-regex")
	expectSize, _ := cmd.Flags().GetUint64("expect-size")
	expectAge, _ := cmd.Flags().GetDuration("expect-age")

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("no files specified")
	}

	checker := file.New(args,
		file.WithAllowAny(allowAny),
		file.WithExpectContentRegex(expectRegex),
		file.WithExpectSize(size.Capacity(expectSize)),
		file.WithExpectAge(expectAge),
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
