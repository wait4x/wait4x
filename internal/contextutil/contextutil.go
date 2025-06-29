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
	"time"
)

// These are context keys used to store and retrieve various values in the context.
type (
	timeoutCtxKey                       struct{}
	intervalCtxKey                      struct{}
	invertCheckCtxKey                   struct{}
	backoffPolicyCtxKey                 struct{}
	backoffCoefficientCtxKey            struct{}
	backoffExponentialMaxIntervalCtxKey struct{}
)

// WithTimeout returns a new context with the given timeout value.
func WithTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, timeoutCtxKey{}, timeout)
}

// GetTimeout retrieves timeout from context
func GetTimeout(ctx context.Context) time.Duration {
	if v := ctx.Value(timeoutCtxKey{}); v != nil {
		return v.(time.Duration)
	}
	return 0
}

// WithInterval returns a new context with the given interval value.
func WithInterval(ctx context.Context, interval time.Duration) context.Context {
	return context.WithValue(ctx, intervalCtxKey{}, interval)
}

// GetInterval retrieves interval from context
func GetInterval(ctx context.Context) time.Duration {
	if v := ctx.Value(intervalCtxKey{}); v != nil {
		return v.(time.Duration)
	}
	return 0
}

// WithInvertCheck returns a new context with the given invert-check value.
func WithInvertCheck(ctx context.Context, invertCheck bool) context.Context {
	return context.WithValue(ctx, invertCheckCtxKey{}, invertCheck)
}

// GetInvertCheck retrieves invert-check from context
func GetInvertCheck(ctx context.Context) bool {
	if v := ctx.Value(invertCheckCtxKey{}); v != nil {
		return v.(bool)
	}
	return false
}

// WithBackoffPolicy returns a new context with the given backoff policy value.
func WithBackoffPolicy(ctx context.Context, backoffPolicy string) context.Context {
	return context.WithValue(ctx, backoffPolicyCtxKey{}, backoffPolicy)
}

// GetBackoffPolicy retrieves the backoff policy from the given context.
func GetBackoffPolicy(ctx context.Context) string {
	if v := ctx.Value(backoffPolicyCtxKey{}); v != nil {
		return v.(string)
	}
	return ""
}

// WithBackoffCoefficient returns a new context with the given backoff coefficient value.
func WithBackoffCoefficient(ctx context.Context, backoffCoefficient float64) context.Context {
	return context.WithValue(ctx, backoffCoefficientCtxKey{}, backoffCoefficient)
}

// GetBackoffCoefficient retrieves the backoff coefficient from the given context.
func GetBackoffCoefficient(ctx context.Context) float64 {
	if v := ctx.Value(backoffCoefficientCtxKey{}); v != nil {
		return v.(float64)
	}
	return 0
}

// WithBackoffExponentialMaxInterval returns a new context with the given backoff exponential max interval value.
func WithBackoffExponentialMaxInterval(ctx context.Context, backoffExponentialMaxInterval time.Duration) context.Context {
	return context.WithValue(ctx, backoffExponentialMaxIntervalCtxKey{}, backoffExponentialMaxInterval)
}

// GetBackoffExponentialMaxInterval retrieves the backoff exponential max interval from the given context.
func GetBackoffExponentialMaxInterval(ctx context.Context) time.Duration {
	if v := ctx.Value(backoffExponentialMaxIntervalCtxKey{}); v != nil {
		return v.(time.Duration)
	}
	return 0
}
