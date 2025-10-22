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

// Package waiter provides the Waiter for the Wait4X application.
package waiter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"wait4x.dev/v3/checker"
)

// Constants representing the available backoff policies for retry mechanisms
const (
	// BackoffPolicyLinear indicates a linear backoff policy,
	BackoffPolicyLinear = "linear"
	// BackoffPolicyExponential indicates an exponential backoff policy.
	BackoffPolicyExponential = "exponential"
)

// Check represents the checker's check method
type Check func(ctx context.Context) error

// Option configures an options for the Waiter
type Option func(s *options)

// options represents the Waiter options
type options struct {
	timeout                       time.Duration
	interval                      time.Duration
	invertCheck                   bool
	logger                        logr.Logger
	backoffPolicy                 string
	backoffExponentialMaxInterval time.Duration
	backoffCoefficient            float64
}

// WithTimeout configures a time limit for whole of checking
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// WithInterval configures time duration for each of checking interval
func WithInterval(interval time.Duration) Option {
	return func(o *options) {
		o.interval = interval
	}
}

// WithInvertCheck configures invert checking
func WithInvertCheck(invertCheck bool) Option {
	return func(o *options) {
		o.invertCheck = invertCheck
	}
}

// WithLogger configures waiter logger
func WithLogger(logger logr.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithBackoffPolicy returns an Option that sets the backoff policy for retries
func WithBackoffPolicy(backoffPolicy string) Option {
	return func(o *options) {
		o.backoffPolicy = backoffPolicy
	}
}

// WithBackoffExponentialMaxInterval is a function that returns an Option which sets the
// maximum interval time duration of the exponential backoff algorithm.
func WithBackoffExponentialMaxInterval(backoffExponentialMaxInterval time.Duration) Option {
	return func(o *options) {
		o.backoffExponentialMaxInterval = backoffExponentialMaxInterval
	}
}

// WithBackoffCoefficient sets the backoffCoefficient for use in retry backoff calculations.
func WithBackoffCoefficient(backoffCoefficient float64) Option {
	return func(o *options) {
		o.backoffCoefficient = backoffCoefficient
	}
}

// WaitParallel waits for end up all of checks execution.
func WaitParallel(checkers []checker.Checker, opts ...Option) error {
	return WaitParallelContext(context.Background(), checkers, opts...)
}

// WaitParallelContext waits for end up all of checks execution.
func WaitParallelContext(ctx context.Context, checkers []checker.Checker, opts ...Option) error {
	// Make channels to pass wgErrors in WaitGroup
	// Use buffered channel to prevent blocking when error occurs
	wgErrors := make(chan error, len(checkers))
	wgDone := make(chan bool)

	var wg sync.WaitGroup

	for _, chr := range checkers {
		wg.Add(1)

		go func(chr checker.Checker) {
			defer wg.Done()

			err := WaitContext(ctx, chr, opts...)
			if err != nil {
				// Non-blocking send to prevent goroutine leak
				select {
				case wgErrors <- err:
				default:
					// Another error was already received, ignore this one
				}
			}
		}(chr)
	}

	// Important final goroutine to wait until WaitGroup is done
	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received through the channel
	select {
	case <-wgDone:
		return nil
	case err := <-wgErrors:
		// Don't close the channel here to avoid race condition
		// It will be garbage collected when the function returns
		return err
	}
}

// Wait waits for end up of check execution.
func Wait(checker checker.Checker, opts ...Option) error {
	return WaitContext(context.Background(), checker, opts...)
}

// WaitContext waits for end up of check execution.
func WaitContext(ctx context.Context, chk checker.Checker, opts ...Option) error {
	options := &options{
		timeout:                       10 * time.Second,
		interval:                      time.Second,
		invertCheck:                   false,
		logger:                        logr.Discard(),
		backoffPolicy:                 BackoffPolicyLinear,
		backoffExponentialMaxInterval: 5 * time.Second,
		backoffCoefficient:            2.0,
	}

	// apply the list of options to waiter
	for _, opt := range opts {
		opt(options)
	}

	// Validate backoff policy once outside the loop
	if options.backoffPolicy != BackoffPolicyLinear && options.backoffPolicy != BackoffPolicyExponential {
		return fmt.Errorf("invalid backoff policy: %s", options.backoffPolicy)
	}

	// Validate exponential backoff parameters
	if options.backoffPolicy == BackoffPolicyExponential {
		if options.backoffCoefficient <= 1.0 {
			return fmt.Errorf("backoff coefficient must be greater than 1.0, got: %f", options.backoffCoefficient)
		}
		if options.backoffExponentialMaxInterval < options.interval {
			return fmt.Errorf("backoff exponential max interval (%v) must be greater than or equal to interval (%v)", options.backoffExponentialMaxInterval, options.interval)
		}
	}

	// Validate interval is positive
	if options.interval <= 0 {
		return fmt.Errorf("interval must be positive, got: %v", options.interval)
	}

	// Ignore timeout context when the timeout is unlimited
	if options.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, options.timeout)
		defer cancel()
	}

	var chkName string
	if t := reflect.TypeOf(chk); t.Kind() == reflect.Ptr {
		chkName = t.Elem().Name()
	} else {
		chkName = t.Name()
	}

	chkID, err := chk.Identity()
	if err != nil {
		return err
	}

	// This is a counter for exponential backoff
	// Maximum value to prevent overflow in exponential calculations
	const maxRetries = 1000000
	retries := 0

	for {
		options.logger.Info(fmt.Sprintf("[%s] Checking %s ...", chkName, chkID))

		err := chk.Check(ctx)
		if err != nil {
			var expectedError *checker.ExpectedError
			if errors.As(err, &expectedError) {
				options.logger.Error(expectedError, "Expectation failed", expectedError.Details()...)
			} else {
				if !errors.Is(err, context.DeadlineExceeded) {
					options.logger.Error(err, "Error occurred")
				}
			}
		}

		// Check if we should stop based on the check result
		// For normal checks: stop when err is nil (success)
		// For inverted checks: stop when err is not nil (failure is success)
		shouldStop := (err == nil && !options.invertCheck) || (err != nil && options.invertCheck)
		if shouldStop {
			break
		}

		// Increment retry counter with bounds checking
		if retries < maxRetries {
			retries++
		}

		// Calculate wait duration based on backoff policy
		var waitDuration time.Duration
		if options.backoffPolicy == BackoffPolicyExponential {
			waitDuration = exponentialBackoff(retries, options.backoffCoefficient, options.interval, options.backoffExponentialMaxInterval)
		} else {
			waitDuration = options.interval
		}

		timer := time.NewTimer(waitDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return nil
}
