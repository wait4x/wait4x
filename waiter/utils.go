// Copyright 2019-2025 The Wait4X Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package waiter provides the Waiter for the Wait4X application.
package waiter

import (
	"fmt"
	"math"
	"time"
)

// exponentialBackoff calculates the exponential backoff duration
func exponentialBackoff(retries int, backoffCoefficient float64, initialInterval, maxInterval time.Duration) (time.Duration, error) {
	multiplier := math.Pow(backoffCoefficient, float64(retries))

	// Handle overflow: if multiplier is infinity or too large, return maxInterval
	if math.IsInf(multiplier, 1) || multiplier > float64(maxInterval)/float64(initialInterval) {
		return maxInterval, nil
	}

	interval := initialInterval * time.Duration(multiplier)

	// This should never happen with validated inputs; if it does, it's a bug
	if interval <= 0 {
		return 0, fmt.Errorf("calculated interval is invalid (%v): initialInterval=%v, multiplier=%v, retries=%d",
			interval, initialInterval, multiplier, retries)
	}

	if interval > maxInterval {
		return maxInterval, nil
	}
	return interval, nil
}
