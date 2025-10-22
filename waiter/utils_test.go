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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestExponentialBackoff tests the exponential backoff calculation
func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name               string
		retries            int
		backoffCoefficient float64
		initialInterval    time.Duration
		maxInterval        time.Duration
		expectedDuration   time.Duration
	}{
		{
			name:               "first retry",
			retries:            0,
			backoffCoefficient: 2.0,
			initialInterval:    100 * time.Millisecond,
			maxInterval:        5 * time.Second,
			expectedDuration:   100 * time.Millisecond, // 100 * 2^0 = 100
		},
		{
			name:               "second retry",
			retries:            1,
			backoffCoefficient: 2.0,
			initialInterval:    100 * time.Millisecond,
			maxInterval:        5 * time.Second,
			expectedDuration:   200 * time.Millisecond, // 100 * 2^1 = 200
		},
		{
			name:               "third retry",
			retries:            2,
			backoffCoefficient: 2.0,
			initialInterval:    100 * time.Millisecond,
			maxInterval:        5 * time.Second,
			expectedDuration:   400 * time.Millisecond, // 100 * 2^2 = 400
		},
		{
			name:               "reaches max interval",
			retries:            10,
			backoffCoefficient: 2.0,
			initialInterval:    100 * time.Millisecond,
			maxInterval:        1 * time.Second,
			expectedDuration:   1 * time.Second, // capped at maxInterval
		},
		{
			name:               "with coefficient 1.5",
			retries:            2,
			backoffCoefficient: 1.5,
			initialInterval:    100 * time.Millisecond,
			maxInterval:        5 * time.Second,
			expectedDuration:   200 * time.Millisecond, // 100 * time.Duration(1.5^2) = 100 * 2 = 200 (truncated)
		},
		{
			name:               "zero retries",
			retries:            0,
			backoffCoefficient: 3.0,
			initialInterval:    50 * time.Millisecond,
			maxInterval:        10 * time.Second,
			expectedDuration:   50 * time.Millisecond, // 50 * 3^0 = 50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exponentialBackoff(tt.retries, tt.backoffCoefficient, tt.initialInterval, tt.maxInterval)
			assert.Equal(t, tt.expectedDuration, result)
		})
	}
}

// TestExponentialBackoffEdgeCases tests edge cases for exponential backoff
func TestExponentialBackoffEdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		retries            int
		backoffCoefficient float64
		initialInterval    time.Duration
		maxInterval        time.Duration
		description        string
	}{
		{
			name:               "very large retry count",
			retries:            100,
			backoffCoefficient: 2.0,
			initialInterval:    1 * time.Millisecond,
			maxInterval:        10 * time.Second,
			description:        "should cap at maxInterval and not overflow",
		},
		{
			name:               "very large coefficient",
			retries:            5,
			backoffCoefficient: 100.0,
			initialInterval:    1 * time.Millisecond,
			maxInterval:        5 * time.Second,
			description:        "should cap at maxInterval",
		},
		{
			name:               "very small initial interval",
			retries:            10,
			backoffCoefficient: 2.0,
			initialInterval:    1 * time.Nanosecond,
			maxInterval:        1 * time.Millisecond,
			description:        "should handle tiny durations",
		},
		{
			name:               "very large initial interval",
			retries:            0,
			backoffCoefficient: 2.0,
			initialInterval:    1 * time.Hour,
			maxInterval:        2 * time.Hour,
			description:        "should handle large durations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exponentialBackoff(tt.retries, tt.backoffCoefficient, tt.initialInterval, tt.maxInterval)

			// Verify result is within bounds
			assert.GreaterOrEqual(t, result, tt.initialInterval, "result should be >= initialInterval")
			assert.LessOrEqual(t, result, tt.maxInterval, "result should be <= maxInterval")

			// Verify result is positive
			assert.Greater(t, result, time.Duration(0), "result should be positive")
		})
	}
}

// TestExponentialBackoffConsistency tests that backoff is consistent and monotonic
func TestExponentialBackoffConsistency(t *testing.T) {
	initialInterval := 100 * time.Millisecond
	maxInterval := 10 * time.Second
	coefficient := 2.0

	var previousDuration time.Duration

	for retries := 0; retries < 20; retries++ {
		currentDuration := exponentialBackoff(retries, coefficient, initialInterval, maxInterval)

		// Verify duration is positive
		assert.Greater(t, currentDuration, time.Duration(0), "duration should be positive at retry %d", retries)

		// Verify duration is within bounds
		assert.GreaterOrEqual(t, currentDuration, initialInterval, "duration should be >= initialInterval at retry %d", retries)
		assert.LessOrEqual(t, currentDuration, maxInterval, "duration should be <= maxInterval at retry %d", retries)

		// Verify monotonic increase until max is reached
		if retries > 0 && previousDuration < maxInterval {
			assert.GreaterOrEqual(t, currentDuration, previousDuration,
				"duration should not decrease at retry %d (prev: %v, current: %v)",
				retries, previousDuration, currentDuration)
		}

		previousDuration = currentDuration
	}
}
