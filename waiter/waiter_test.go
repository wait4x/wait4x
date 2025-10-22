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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tonglil/buflogr"
	"wait4x.dev/v3/checker"
)

// TestMain is the main function for the Waiter.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestWaitSuccessful tests the Waiter with a successful check.
func TestWaitSuccessful(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Check", mock.Anything).Return(nil).
		On("Identity").Return("ID", nil)

	err := Wait(mockChecker, WithInterval(time.Second))

	assert.Nil(t, err)
	mockChecker.AssertExpectations(t)
}

// TestWaitTimedOut tests the Waiter with a timed out check.
func TestWaitTimedOut(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).
		On("Identity").Return("ID", nil)

	err := Wait(mockChecker, WithTimeout(time.Second))

	assert.Equal(t, context.DeadlineExceeded, err)
	mockChecker.AssertExpectations(t)
}

// TestWaitInvalidIdentity tests the Waiter with an invalid identity.
func TestWaitInvalidIdentity(t *testing.T) {
	invalidIdentityError := errors.New("invalid identity")

	mockChecker := new(checker.MockChecker)
	mockChecker.On("Identity").Return(mock.Anything, invalidIdentityError)

	err := Wait(mockChecker)

	assert.Equal(t, invalidIdentityError, err)
	mockChecker.AssertExpectations(t)
}

// TestWaitLogger tests the Waiter with a logger.
func TestWaitLogger(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Check", mock.Anything).
		Return(fmt.Errorf("error message")).
		On("Identity").Return("ID", nil)

	var buf bytes.Buffer
	var log = buflogr.NewWithBuffer(&buf)
	err := WaitContext(context.TODO(), mockChecker, WithLogger(log), WithTimeout(time.Second))

	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Contains(t, buf.String(), "INFO [MockChecker] Checking ID ...")
	assert.Contains(t, buf.String(), "error message")
	mockChecker.AssertExpectations(t)
}

// TestWaitInvertCheck tests the Waiter with an inverted check.
func TestWaitInvertCheck(t *testing.T) {
	alwaysTrue := new(checker.MockChecker)
	alwaysTrue.On("Check", mock.Anything).Return(nil).
		On("Identity").Return("ID", nil)

	err := Wait(alwaysTrue, WithTimeout(time.Second*3), WithInvertCheck(true))
	assert.Equal(t, context.DeadlineExceeded, err)
	alwaysTrue.AssertExpectations(t)

	alwaysFalse := new(checker.MockChecker)
	alwaysFalse.On("Check", mock.Anything).Return(fmt.Errorf("error")).
		On("Identity").Return("ID", nil)

	err = Wait(alwaysFalse, WithTimeout(time.Second), WithInvertCheck(true))
	assert.Nil(t, err)
	alwaysFalse.AssertExpectations(t)
}

// TestWaitParallelSuccessful tests the Waiter with a parallel successful check.
func TestWaitParallelSuccessful(t *testing.T) {
	alwaysTrueFirst := new(checker.MockChecker)
	alwaysTrueFirst.On("Check", mock.Anything).Return(nil).
		On("Identity").Return("ID", nil)

	alwaysTrueSecond := new(checker.MockChecker)
	alwaysTrueSecond.On("Check", mock.Anything).Return(nil).
		On("Identity").Return("ID", nil)

	err := WaitParallel([]checker.Checker{alwaysTrueFirst, alwaysTrueSecond}, WithTimeout(time.Second*3))
	assert.Nil(t, err)
	alwaysTrueFirst.AssertExpectations(t)
	alwaysTrueSecond.AssertExpectations(t)
}

// TestWaitParallelFail tests the Waiter with a parallel failed check.
func TestWaitParallelFail(t *testing.T) {
	alwaysTrueFirst := new(checker.MockChecker)
	alwaysTrueFirst.On("Check", mock.Anything).Return(nil).
		On("Identity").Return("ID", nil)

	alwaysTrueSecond := new(checker.MockChecker)
	alwaysTrueSecond.On("Check", mock.Anything).Return(nil).
		On("Identity").Return("ID", nil)

	alwaysError := new(checker.MockChecker)
	alwaysError.On("Check", mock.Anything).Return(fmt.Errorf("error")).
		On("Identity").Return("ID", nil)

	err := WaitParallel([]checker.Checker{alwaysTrueFirst, alwaysTrueSecond, alwaysError}, WithTimeout(time.Second*3))
	assert.Equal(t, context.DeadlineExceeded, err)

	alwaysTrueFirst.AssertExpectations(t)
	alwaysTrueSecond.AssertExpectations(t)
	alwaysError.AssertExpectations(t)
}

// TestWaitInvalidBackoffPolicy tests the Waiter with an invalid backoff policy.
func TestWaitInvalidBackoffPolicy(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	// Note: Identity() is not called because validation happens first

	err := Wait(mockChecker, WithBackoffPolicy("invalid-policy"))

	assert.EqualError(t, err, "invalid backoff policy: invalid-policy")
}

// TestWaitExponentialBackoff tests the Waiter with exponential backoff policy.
func TestWaitExponentialBackoff(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Identity").Return("ID", nil)
	mockChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).Times(2)
	mockChecker.On("Check", mock.Anything).Return(nil).Once()

	start := time.Now()
	err := Wait(
		mockChecker,
		WithBackoffPolicy(BackoffPolicyExponential),
		WithInterval(100*time.Millisecond),
		WithBackoffCoefficient(2.0),
		WithBackoffExponentialMaxInterval(500*time.Millisecond),
		WithTimeout(5*time.Second),
	)
	elapsed := time.Since(start)

	assert.Nil(t, err)
	// First check: immediate (error)
	// Second check: after 100ms (2^0 * 100ms) (error)
	// Third check: after 200ms (2^1 * 100ms) (success)
	// Total should be around 300ms, allowing overhead
	assert.Greater(t, elapsed, 250*time.Millisecond)
	assert.Less(t, elapsed, 800*time.Millisecond)
	mockChecker.AssertExpectations(t)
}

// TestWaitLinearBackoff tests the Waiter with linear backoff policy.
func TestWaitLinearBackoff(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Identity").Return("ID", nil)
	mockChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).Times(2)
	mockChecker.On("Check", mock.Anything).Return(nil).Once()

	start := time.Now()
	err := Wait(
		mockChecker,
		WithBackoffPolicy(BackoffPolicyLinear),
		WithInterval(100*time.Millisecond),
		WithTimeout(5*time.Second),
	)
	elapsed := time.Since(start)

	assert.Nil(t, err)
	// First check: immediate (error)
	// Second check: after 100ms (error)
	// Third check: after 100ms (success)
	// Total should be around 200ms
	assert.Greater(t, elapsed, 180*time.Millisecond)
	assert.Less(t, elapsed, 300*time.Millisecond)
	mockChecker.AssertExpectations(t)
}

// TestWaitInvalidBackoffCoefficient tests the Waiter with an invalid backoff coefficient.
func TestWaitInvalidBackoffCoefficient(t *testing.T) {
	tests := []struct {
		name        string
		coefficient float64
		wantError   string
	}{
		{
			name:        "coefficient less than 1",
			coefficient: 0.5,
			wantError:   "backoff coefficient must be greater than 1.0, got: 0.500000",
		},
		{
			name:        "coefficient equal to 1",
			coefficient: 1.0,
			wantError:   "backoff coefficient must be greater than 1.0, got: 1.000000",
		},
		{
			name:        "coefficient zero",
			coefficient: 0.0,
			wantError:   "backoff coefficient must be greater than 1.0, got: 0.000000",
		},
		{
			name:        "negative coefficient",
			coefficient: -2.0,
			wantError:   "backoff coefficient must be greater than 1.0, got: -2.000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChecker := new(checker.MockChecker)

			err := Wait(
				mockChecker,
				WithBackoffPolicy(BackoffPolicyExponential),
				WithBackoffCoefficient(tt.coefficient),
			)

			assert.EqualError(t, err, tt.wantError)
		})
	}
}

// TestWaitInvalidMaxInterval tests the Waiter with an invalid max interval.
func TestWaitInvalidMaxInterval(t *testing.T) {
	mockChecker := new(checker.MockChecker)

	err := Wait(
		mockChecker,
		WithBackoffPolicy(BackoffPolicyExponential),
		WithInterval(time.Second),
		WithBackoffExponentialMaxInterval(500*time.Millisecond),
	)

	assert.EqualError(t, err, "backoff exponential max interval (500ms) must be greater than or equal to interval (1s)")
}

// TestWaitInvalidInterval tests the Waiter with an invalid interval.
func TestWaitInvalidInterval(t *testing.T) {
	tests := []struct {
		name      string
		interval  time.Duration
		wantError string
	}{
		{
			name:      "zero interval",
			interval:  0,
			wantError: "interval must be positive, got: 0s",
		},
		{
			name:      "negative interval",
			interval:  -1 * time.Second,
			wantError: "interval must be positive, got: -1s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChecker := new(checker.MockChecker)

			err := Wait(mockChecker, WithInterval(tt.interval))

			assert.EqualError(t, err, tt.wantError)
		})
	}
}

// TestWaitExponentialBackoffMaxInterval tests that exponential backoff respects max interval.
func TestWaitExponentialBackoffMaxInterval(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Identity").Return("ID", nil)
	mockChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).Times(5)
	mockChecker.On("Check", mock.Anything).Return(nil).Once()

	start := time.Now()
	err := Wait(
		mockChecker,
		WithBackoffPolicy(BackoffPolicyExponential),
		WithInterval(50*time.Millisecond),
		WithBackoffCoefficient(2.0),
		WithBackoffExponentialMaxInterval(150*time.Millisecond),
		WithTimeout(10*time.Second),
	)
	elapsed := time.Since(start)

	assert.Nil(t, err)
	// First check: immediate (error)
	// Second check: after 50ms (2^0 * 50ms = 50ms) (error)
	// Third check: after 100ms (2^1 * 50ms = 100ms) (error)
	// Fourth check: after 150ms (2^2 * 50ms = 200ms, capped at 150ms) (error)
	// Fifth check: after 150ms (capped) (error)
	// Sixth check: after 150ms (capped) (success)
	// Total should be around 600ms
	assert.Greater(t, elapsed, 550*time.Millisecond)
	assert.Less(t, elapsed, 850*time.Millisecond)
	mockChecker.AssertExpectations(t)
}

// TestWaitWithUnlimitedTimeout tests the Waiter with unlimited timeout (0 duration).
func TestWaitWithUnlimitedTimeout(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Identity").Return("ID", nil)
	mockChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).Times(2)
	mockChecker.On("Check", mock.Anything).Return(nil).Once()

	err := Wait(
		mockChecker,
		WithTimeout(0), // Unlimited timeout
		WithInterval(50*time.Millisecond),
	)

	assert.Nil(t, err)
	mockChecker.AssertExpectations(t)
}

// TestWaitParallelWithDifferentOptions tests parallel waiter with different configurations.
func TestWaitParallelWithDifferentOptions(t *testing.T) {
	// Test that options are applied correctly to each checker independently
	fastChecker := new(checker.MockChecker)
	fastChecker.On("Check", mock.Anything).Return(nil).
		On("Identity").Return("FastChecker", nil)

	slowChecker := new(checker.MockChecker)
	slowChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).Times(2)
	slowChecker.On("Check", mock.Anything).Return(nil).Once()
	slowChecker.On("Identity").Return("SlowChecker", nil)

	err := WaitParallel(
		[]checker.Checker{fastChecker, slowChecker},
		WithTimeout(5*time.Second),
		WithInterval(100*time.Millisecond),
	)

	assert.Nil(t, err)
	fastChecker.AssertExpectations(t)
	slowChecker.AssertExpectations(t)
}

// TestWaitContextCancellation tests the Waiter with context cancellation.
func TestWaitContextCancellation(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).
		On("Identity").Return("ID", nil)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := WaitContext(ctx, mockChecker, WithInterval(50*time.Millisecond))

	assert.Equal(t, context.Canceled, err)
	mockChecker.AssertExpectations(t)
}

// TestWaitExponentialBackoffWithLargeCoefficient tests exponential backoff with large coefficient.
func TestWaitExponentialBackoffWithLargeCoefficient(t *testing.T) {
	mockChecker := new(checker.MockChecker)
	mockChecker.On("Identity").Return("ID", nil)
	mockChecker.On("Check", mock.Anything).Return(fmt.Errorf("error")).Times(2)
	mockChecker.On("Check", mock.Anything).Return(nil).Once()

	start := time.Now()
	err := Wait(
		mockChecker,
		WithBackoffPolicy(BackoffPolicyExponential),
		WithInterval(10*time.Millisecond),
		WithBackoffCoefficient(10.0),
		WithBackoffExponentialMaxInterval(time.Second),
		WithTimeout(5*time.Second),
	)
	elapsed := time.Since(start)

	assert.Nil(t, err)
	// First check: immediate (error)
	// Second check: after 10ms (10^0 * 10ms = 10ms) (error)
	// Third check: after 100ms (10^1 * 10ms = 100ms) (success)
	// Total should be around 110ms, but allow for scheduling overhead
	assert.Greater(t, elapsed, 100*time.Millisecond)
	assert.Less(t, elapsed, 2*time.Second)
	mockChecker.AssertExpectations(t)
}
