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

// Package main is the main package for the TCP Basic example.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"wait4x.dev/v4/checker/tcp"
	"wait4x.dev/v4/waiter"
)

// main is the main function for the TCP Basic example
func main() {
	// Create a context with a 30-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a TCP checker for localhost:6379 with a 5-second connection timeout
	tcpChecker := tcp.New("localhost:6379", tcp.WithTimeout(5*time.Second))

	// Specify waiter options
	options := []waiter.Option{
		waiter.WithTimeout(time.Minute),                            // Total wait timeout
		waiter.WithInterval(2 * time.Second),                       // Time between retry attempts
		waiter.WithBackoffPolicy("exponential"),                    // Use exponential backoff
		waiter.WithBackoffCoefficient(2.0),                         // Double the wait time each retry
		waiter.WithBackoffExponentialMaxInterval(10 * time.Second), // Max 10s between retries
	}

	// Wait for the TCP port to be available
	fmt.Println("Waiting for Redis to be available on port 6379...")
	err := waiter.WaitContext(ctx, tcpChecker, options...)
	if err != nil {
		log.Fatalf("Failed waiting for Redis: %v", err)
	}

	fmt.Println("Redis is available!")
}
