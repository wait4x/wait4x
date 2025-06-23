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

// Package contextutil provides utilities for working with the Go context package.
package contextutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ContextUtilSuite is a test suite for contextutil package
type ContextUtilSuite struct {
	suite.Suite
}

// TestTimeoutFunctions tests both WithTimeout and GetTimeout functions
func (s *ContextUtilSuite) TestTimeoutFunctions() {
	ctx := context.Background()
	timeout := 5 * time.Second

	// Test setting and getting timeout
	ctxWithTimeout := WithTimeout(ctx, timeout)
	s.Equal(timeout, GetTimeout(ctxWithTimeout))

	// Test that original context is not modified
	s.Equal(time.Duration(0), GetTimeout(ctx))

	// Test nested contexts (overwrite)
	ctxWithTimeout2 := WithTimeout(ctxWithTimeout, 10*time.Second)
	s.Equal(10*time.Second, GetTimeout(ctxWithTimeout2))

	// Test zero timeout
	ctxWithZeroTimeout := WithTimeout(ctx, 0)
	s.Equal(time.Duration(0), GetTimeout(ctxWithZeroTimeout))
}

// TestIntervalFunctions tests both WithInterval and GetInterval functions
func (s *ContextUtilSuite) TestIntervalFunctions() {
	ctx := context.Background()
	interval := 2 * time.Second

	// Test setting and getting interval
	ctxWithInterval := WithInterval(ctx, interval)
	s.Equal(interval, GetInterval(ctxWithInterval))

	// Test that original context is not modified
	s.Equal(time.Duration(0), GetInterval(ctx))

	// Test nested contexts (overwrite)
	ctxWithInterval2 := WithInterval(ctxWithInterval, 7*time.Second)
	s.Equal(7*time.Second, GetInterval(ctxWithInterval2))

	// Test zero interval
	ctxWithZeroInterval := WithInterval(ctx, 0)
	s.Equal(time.Duration(0), GetInterval(ctxWithZeroInterval))
}

// TestInvertCheckFunctions tests both WithInvertCheck and GetInvertCheck functions
func (s *ContextUtilSuite) TestInvertCheckFunctions() {
	ctx := context.Background()

	// Test setting and getting invert check to true
	ctxWithInvertTrue := WithInvertCheck(ctx, true)
	s.True(GetInvertCheck(ctxWithInvertTrue))

	// Test setting and getting invert check to false
	ctxWithInvertFalse := WithInvertCheck(ctx, false)
	s.False(GetInvertCheck(ctxWithInvertFalse))

	// Test that original context is not modified
	s.False(GetInvertCheck(ctx))

	// Test nested contexts (overwrite)
	ctxWithInvertTrue2 := WithInvertCheck(ctxWithInvertTrue, false)
	s.False(GetInvertCheck(ctxWithInvertTrue2))
}

// TestBackoffPolicyFunctions tests both WithBackoffPolicy and GetBackoffPolicy functions
func (s *ContextUtilSuite) TestBackoffPolicyFunctions() {
	ctx := context.Background()
	policy := "exponential"

	// Test setting and getting backoff policy
	ctxWithPolicy := WithBackoffPolicy(ctx, policy)
	s.Equal(policy, GetBackoffPolicy(ctxWithPolicy))

	// Test that original context is not modified
	s.Equal("", GetBackoffPolicy(ctx))

	// Test nested contexts (overwrite)
	ctxWithPolicy2 := WithBackoffPolicy(ctxWithPolicy, "linear")
	s.Equal("linear", GetBackoffPolicy(ctxWithPolicy2))

	// Test empty policy
	ctxWithEmptyPolicy := WithBackoffPolicy(ctx, "")
	s.Equal("", GetBackoffPolicy(ctxWithEmptyPolicy))
}

// TestBackoffCoefficientFunctions tests both WithBackoffCoefficient and GetBackoffCoefficient functions
func (s *ContextUtilSuite) TestBackoffCoefficientFunctions() {
	ctx := context.Background()
	coefficient := 2.5

	// Test setting and getting backoff coefficient
	ctxWithCoefficient := WithBackoffCoefficient(ctx, coefficient)
	s.Equal(coefficient, GetBackoffCoefficient(ctxWithCoefficient))

	// Test that original context is not modified
	s.Equal(0.0, GetBackoffCoefficient(ctx))

	// Test nested contexts (overwrite)
	ctxWithCoefficient2 := WithBackoffCoefficient(ctxWithCoefficient, 1.8)
	s.Equal(1.8, GetBackoffCoefficient(ctxWithCoefficient2))

	// Test zero coefficient
	ctxWithZeroCoefficient := WithBackoffCoefficient(ctx, 0.0)
	s.Equal(0.0, GetBackoffCoefficient(ctxWithZeroCoefficient))

	// Test negative coefficient
	ctxWithNegativeCoefficient := WithBackoffCoefficient(ctx, -1.5)
	s.Equal(-1.5, GetBackoffCoefficient(ctxWithNegativeCoefficient))
}

// TestBackoffExponentialMaxIntervalFunctions tests both WithBackoffExponentialMaxInterval and GetBackoffExponentialMaxInterval functions
func (s *ContextUtilSuite) TestBackoffExponentialMaxIntervalFunctions() {
	ctx := context.Background()
	maxInterval := 30 * time.Second

	// Test setting and getting backoff exponential max interval
	ctxWithMaxInterval := WithBackoffExponentialMaxInterval(ctx, maxInterval)
	s.Equal(maxInterval, GetBackoffExponentialMaxInterval(ctxWithMaxInterval))

	// Test that original context is not modified
	s.Equal(time.Duration(0), GetBackoffExponentialMaxInterval(ctx))

	// Test nested contexts (overwrite)
	ctxWithMaxInterval2 := WithBackoffExponentialMaxInterval(ctxWithMaxInterval, 60*time.Second)
	s.Equal(60*time.Second, GetBackoffExponentialMaxInterval(ctxWithMaxInterval2))

	// Test zero interval
	ctxWithZeroInterval := WithBackoffExponentialMaxInterval(ctx, 0)
	s.Equal(time.Duration(0), GetBackoffExponentialMaxInterval(ctxWithZeroInterval))
}

// TestMultipleValues tests setting multiple values on the same context
func (s *ContextUtilSuite) TestMultipleValues() {
	ctx := context.Background()

	// Set multiple values
	ctx = WithTimeout(ctx, 10*time.Second)
	ctx = WithInterval(ctx, 2*time.Second)
	ctx = WithInvertCheck(ctx, true)
	ctx = WithBackoffPolicy(ctx, "exponential")
	ctx = WithBackoffCoefficient(ctx, 2.0)
	ctx = WithBackoffExponentialMaxInterval(ctx, 60*time.Second)

	// Verify all values are set correctly
	s.Equal(10*time.Second, GetTimeout(ctx))
	s.Equal(2*time.Second, GetInterval(ctx))
	s.True(GetInvertCheck(ctx))
	s.Equal("exponential", GetBackoffPolicy(ctx))
	s.Equal(2.0, GetBackoffCoefficient(ctx))
	s.Equal(60*time.Second, GetBackoffExponentialMaxInterval(ctx))
}

// TestContextCompatibility tests that our values work with standard context operations
func (s *ContextUtilSuite) TestContextCompatibility() {
	ctx := context.Background()

	// Set values
	ctx = WithTimeout(ctx, 5*time.Second)
	ctx = WithInterval(ctx, 1*time.Second)
	ctx = WithInvertCheck(ctx, true)

	// Test with cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	s.Equal(5*time.Second, GetTimeout(cancelCtx))
	s.Equal(1*time.Second, GetInterval(cancelCtx))
	s.True(GetInvertCheck(cancelCtx))

	// Test with deadline
	deadlineCtx, cancelDeadline := context.WithDeadline(ctx, time.Now().Add(1*time.Hour))
	defer cancelDeadline()
	s.Equal(5*time.Second, GetTimeout(deadlineCtx))
	s.Equal(1*time.Second, GetInterval(deadlineCtx))
	s.True(GetInvertCheck(deadlineCtx))

	// Test with timeout
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()
	s.Equal(5*time.Second, GetTimeout(timeoutCtx))
	s.Equal(1*time.Second, GetInterval(timeoutCtx))
	s.True(GetInvertCheck(timeoutCtx))
}

// TestDefaultValues tests default values for all getters
func (s *ContextUtilSuite) TestDefaultValues() {
	ctx := context.TODO()

	s.Equal(time.Duration(0), GetTimeout(ctx))
	s.Equal(time.Duration(0), GetInterval(ctx))
	s.False(GetInvertCheck(ctx))
	s.Equal("", GetBackoffPolicy(ctx))
	s.Equal(0.0, GetBackoffCoefficient(ctx))
	s.Equal(time.Duration(0), GetBackoffExponentialMaxInterval(ctx))
}

// TestTableDriven tests multiple scenarios in a table-driven approach
func (s *ContextUtilSuite) TestTableDriven() {
	tests := []struct {
		name     string
		setup    func(context.Context) context.Context
		expected map[string]interface{}
	}{
		{
			name: "all values set",
			setup: func(ctx context.Context) context.Context {
				ctx = WithTimeout(ctx, 15*time.Second)
				ctx = WithInterval(ctx, 3*time.Second)
				ctx = WithInvertCheck(ctx, true)
				ctx = WithBackoffPolicy(ctx, "constant")
				ctx = WithBackoffCoefficient(ctx, 1.5)
				ctx = WithBackoffExponentialMaxInterval(ctx, 90*time.Second)
				return ctx
			},
			expected: map[string]interface{}{
				"timeout":                       15 * time.Second,
				"interval":                      3 * time.Second,
				"invertCheck":                   true,
				"backoffPolicy":                 "constant",
				"backoffCoefficient":            1.5,
				"backoffExponentialMaxInterval": 90 * time.Second,
			},
		},
		{
			name: "zero values",
			setup: func(ctx context.Context) context.Context {
				ctx = WithTimeout(ctx, 0)
				ctx = WithInterval(ctx, 0)
				ctx = WithInvertCheck(ctx, false)
				ctx = WithBackoffPolicy(ctx, "")
				ctx = WithBackoffCoefficient(ctx, 0.0)
				ctx = WithBackoffExponentialMaxInterval(ctx, 0)
				return ctx
			},
			expected: map[string]interface{}{
				"timeout":                       time.Duration(0),
				"interval":                      time.Duration(0),
				"invertCheck":                   false,
				"backoffPolicy":                 "",
				"backoffCoefficient":            0.0,
				"backoffExponentialMaxInterval": time.Duration(0),
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()
			ctx = tt.setup(ctx)

			s.Equal(tt.expected["timeout"], GetTimeout(ctx))
			s.Equal(tt.expected["interval"], GetInterval(ctx))
			s.Equal(tt.expected["invertCheck"], GetInvertCheck(ctx))
			s.Equal(tt.expected["backoffPolicy"], GetBackoffPolicy(ctx))
			s.Equal(tt.expected["backoffCoefficient"], GetBackoffCoefficient(ctx))
			s.Equal(tt.expected["backoffExponentialMaxInterval"], GetBackoffExponentialMaxInterval(ctx))
		})
	}
}

// TestContextutilSuite runs the test suite
func TestContextutilSuite(t *testing.T) {
	suite.Run(t, new(ContextUtilSuite))
}
