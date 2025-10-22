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
