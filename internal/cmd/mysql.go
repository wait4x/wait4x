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

//go:build !disable_mysql

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v4/checker"
	"wait4x.dev/v4/checker/mysql"
	"wait4x.dev/v4/internal/contextutil"
	"wait4x.dev/v4/waiter"
)

// NewMysqlCommand creates a new mysql sub-command
func NewMysqlCommand() *cobra.Command {
	mysqlCommand := &cobra.Command{
		Use:   "mysql DSN... [flags] [-- command [args...]]",
		Short: "Check MySQL connection",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("DSN is required argument for the mysql command")
			}

			return nil
		},
		Example: `
  # Checking MySQL TCP connection
  wait4x mysql user:password@tcp(localhost:5555)/dbname?tls=skip-verify

  # Checking MySQL UNIX Socket existence
  wait4x mysql username:password@unix(/tmp/mysql.sock)/myDatabase
`,
		RunE: runMysql,
	}

	mysqlCommand.Flags().String("expect-table", "", "Expect a table to exist in the database")

	return mysqlCommand
}

func runMysql(cmd *cobra.Command, args []string) error {
	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("unable to get logger from context: %w", err)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	expectTable, err := cmd.Flags().GetString("expect-table")
	if err != nil {
		return fmt.Errorf("failed to parse --expect-table flag: %w", err)
	}

	checkers := make([]checker.Checker, len(args))
	for i, arg := range args {
		checkers[i] = mysql.New(arg, mysql.WithExpectTable(expectTable))
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
