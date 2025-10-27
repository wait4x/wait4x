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

//go:build !disable_aws_s3

// Package aws provides the command-line interface for the Wait4X application.
package aws

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v4/checker"
	"wait4x.dev/v4/checker/aws/s3"
	"wait4x.dev/v4/internal/contextutil"
	"wait4x.dev/v4/waiter"
)

// NewS3Command returns a new S3 subcommand for checking S3 bucket existence.
func NewS3Command() *cobra.Command {
	s3Cmd := &cobra.Command{
		Use:   "s3 BUCKET_PATH... [flags] [-- command [args...]]",
		Short: "Check S3 bucket existence",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("BUCKET_PATH is required argument for the s3 command")
			}
			return nil
		},
		Example: `
  # Check if S3 bucket exists using bucket name
  wait4x aws s3 my-bucket-name

  # Check if S3 bucket exists using s3:// URI
  wait4x aws s3 s3://my-bucket-name

  # Check if S3 bucket exists using ARN
  wait4x aws s3 arn:aws:s3:::my-bucket-name

  # Check multiple buckets
  wait4x aws s3 bucket1 bucket2 s3://bucket3
`,
		RunE: runS3,
	}

	return s3Cmd
}

// runS3 executes the S3 command.
func runS3(cmd *cobra.Command, args []string) error {
	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	checkers := make([]checker.Checker, len(args))
	for i, arg := range args {
		checkers[i] = s3.New(arg)
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
