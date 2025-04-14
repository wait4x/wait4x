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

//go:build !disable_mongodb

package cmd

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/mongodb"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewMongoDBCommand creates the mongodb sub-command
func NewMongoDBCommand() *cobra.Command {
	mongodbCommand := &cobra.Command{
		Use:   "mongodb DSN... [flags] [-- command [args...]]",
		Short: "Check MongoDB connection",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("DSN is required argument for the mongodb command")
			}

			return nil
		},
		Example: `
  # Checking MongoDB connection
  wait4x mongodb 'mongodb://127.0.0.1:27017'

  # Checking MongoDB connection with credentials and options
  wait4x mongodb 'mongodb://user:pass@127.0.0.1:27017/?maxPoolSize=20&w=majority'
`,
		RunE: runMongoDB,
	}

	return mongodbCommand
}

func runMongoDB(cmd *cobra.Command, args []string) error {
	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("unable to get logger from context: %w", err)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	checkers := make([]checker.Checker, len(args))
	for i, arg := range args {
		checkers[i] = mongodb.New(arg)
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
