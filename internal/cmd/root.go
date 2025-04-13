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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/internal/cmd/dns"
	"wait4x.dev/v3/internal/cmd/temporal"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

const (
	// ExitError is the exit code used when the command encounters an error.
	ExitError = 1

	// ExitTimedOut is the exit code used when the command times out.
	ExitTimedOut = 124
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "wait4x",
		Short: "Wait4X allows waiting for a port or a service to enter into specify state",
		Long:  `Wait4X allows waiting for a port to enter into specify state or waiting for a service e.g. redis, mysql, postgres, ... to enter inter ready state`,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			quiet, err := cmd.Flags().GetBool("quiet")
			if err != nil {
				return fmt.Errorf("unable to parse --quiet flag: %w", err)
			}

			noColor, err := cmd.Flags().GetBool("no-color")
			if err != nil {
				return fmt.Errorf("unable to parse --no-color flag: %w", err)
			}

			timeout, err := cmd.Flags().GetDuration("timeout")
			if err != nil {
				return fmt.Errorf("unable to parse --timeout flag: %w", err)
			}

			interval, err := cmd.Flags().GetDuration("interval")
			if err != nil {
				return fmt.Errorf("unable to parse --interval flag: %w", err)
			}

			invertCheck, err := cmd.Flags().GetBool("invert-check")
			if err != nil {
				return fmt.Errorf("unable to parse --invert-check flag: %w", err)
			}

			backoffPolicy, err := cmd.Flags().GetString("backoff-policy")
			if err != nil {
				return fmt.Errorf("unable to parse --backoff-policy flag: %w", err)
			}

			backoffCoefficient, err := cmd.Flags().GetFloat64("backoff-exponential-coefficient")
			if err != nil {
				return fmt.Errorf("unable to parse --backoff-exponential-coefficient flag: %w", err)
			}

			backoffExpMaxInterval, err := cmd.Flags().GetDuration("backoff-exponential-max-interval")
			if err != nil {
				return fmt.Errorf("unable to parse --backoff-exponential-max-interval flag: %w", err)
			}

			cmd.SetContext(contextutil.WithTimeout(cmd.Context(), timeout))
			cmd.SetContext(contextutil.WithInterval(cmd.Context(), interval))
			cmd.SetContext(contextutil.WithInvertCheck(cmd.Context(), invertCheck))
			cmd.SetContext(contextutil.WithBackoffPolicy(cmd.Context(), backoffPolicy))
			cmd.SetContext(contextutil.WithBackoffCoefficient(cmd.Context(), backoffCoefficient))
			cmd.SetContext(contextutil.WithBackoffExponentialMaxInterval(cmd.Context(), backoffExpMaxInterval))

			// Validate backoff policy value
			backoffPolicyValues := []string{waiter.BackoffPolicyExponential, waiter.BackoffPolicyLinear}
			if !contains(backoffPolicyValues, backoffPolicy) {
				return fmt.Errorf("--backoff-policy must be one of %v", backoffPolicyValues)
			}

			if backoffPolicy == waiter.BackoffPolicyExponential && backoffExpMaxInterval < interval {
				return fmt.Errorf("--backoff-exponential-max-interval must be greater than --interval")
			}

			// Prevent showing error when the quiet mode enabled.
			cmd.SilenceErrors = quiet

			lvl := zerolog.InfoLevel
			if quiet {
				lvl = zerolog.Disabled
			}

			// Prevent showing usage when subcommand return error.
			cmd.SilenceUsage = true

			zl := zerolog.New(
				zerolog.ConsoleWriter{
					Out:        os.Stderr,
					NoColor:    color.NoColor || noColor,
					TimeFormat: time.RFC3339,
				},
			).Level(lvl).
				With().
				Timestamp().
				Logger()
			logger := zerologr.New(&zl)
			// VerbosityFieldName (v) is not emitted.
			zerologr.VerbosityFieldName = ""
			cmd.SetContext(logr.NewContext(cmd.Context(), logger))

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.ArgsLenAtDash() != -1 && (len(args)-cmd.ArgsLenAtDash()) > 0 {
				command := args[cmd.ArgsLenAtDash():][0]
				arguments := args[cmd.ArgsLenAtDash():][1:]
				for i, arg := range arguments {
					arguments[i] = os.ExpandEnv(arg)
				}

				c := exec.CommandContext(cmd.Context(), command, arguments...)
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr

				return c.Run()
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().DurationP("interval", "i", 1*time.Second, "Interval time between each loop.")
	rootCmd.PersistentFlags().String("backoff-policy", "linear", `Select the backoff policy ("`+waiter.BackoffPolicyLinear+`"|"`+waiter.BackoffPolicyExponential+`".`)
	rootCmd.PersistentFlags().Duration("backoff-exponential-max-interval", 5*time.Second, "Maximum interval time between each loop when backoff-policy is exponential.")
	rootCmd.PersistentFlags().Float64("backoff-exponential-coefficient", 2.0, "Coefficient used to calculate the exponential backoff when backoff-policy is exponential.")
	rootCmd.PersistentFlags().DurationP("timeout", "t", 10*time.Second, "Timeout is the maximum amount of time that Wait4X will wait for a checking operation, 0 is unlimited.")
	rootCmd.PersistentFlags().BoolP("invert-check", "v", false, "Invert the sense of checking.")
	rootCmd.PersistentFlags().StringP("log-level", "l", zerolog.InfoLevel.String(), "Set the logging level (\"trace\"|\"debug\"|\"info\")")
	rootCmd.PersistentFlags().MarkDeprecated("log-level", "You don't need to the flag anymore. By default, Wait4X returns error logs. This flag will be removed in v4.0.0")
	rootCmd.PersistentFlags().Bool("no-color", false, "If specified, output won't contain any color.")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet or silent mode. Do not show logs or error messages.")

	return rootCmd
}

// Execute run Wait4X application
func Execute() {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewTCPCommand())
	rootCmd.AddCommand(dns.NewDNSCommand())
	rootCmd.AddCommand(NewHTTPCommand())
	rootCmd.AddCommand(NewPostgresqlCommand())
	rootCmd.AddCommand(NewMysqlCommand())
	rootCmd.AddCommand(NewRedisCommand())
	rootCmd.AddCommand(NewInfluxDBCommand())
	rootCmd.AddCommand(NewMongoDBCommand())
	rootCmd.AddCommand(NewRabbitMQCommand())
	rootCmd.AddCommand(temporal.NewTemporalCommand())
	rootCmd.AddCommand(NewVersionCommand())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			os.Exit(ExitTimedOut)
		}

		os.Exit(ExitError)
	}
}
