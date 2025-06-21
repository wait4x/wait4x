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

// Package main is the main package for the Parallel Services example.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-logr/stdr"
	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/http"
	"wait4x.dev/v3/checker/postgresql"
	"wait4x.dev/v3/checker/redis"
	"wait4x.dev/v3/waiter"
)

// main is the main function for the Parallel Services example
func main() {
	// Set up a logger
	stdr.SetVerbosity(4) // Set log level
	logger := stdr.New(log.New(log.Writer(), "[Wait4X] ", log.LstdFlags|log.Lshortfile))

	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create checkers for different services
	checkers := []checker.Checker{
		// Redis checker
		redis.New(
			"redis://localhost:6379",
			redis.WithTimeout(5*time.Second),
			redis.WithExpectKey("app:status=ready"), // Check if key exists with specific value
		),

		// PostgreSQL checker
		postgresql.New(
			"postgres://postgres:password@localhost:5432/app_db?sslmode=disable",
		),

		// HTTP API checker
		http.New(
			"http://localhost:8080/health",
			http.WithTimeout(3*time.Second),
			http.WithExpectStatusCode(200),
			http.WithExpectBodyJSON("status.healthy"),
		),
	}

	// Set up common options for all waiters
	waitOptions := []waiter.Option{
		waiter.WithTimeout(time.Minute),
		waiter.WithInterval(2 * time.Second),
		waiter.WithBackoffPolicy(waiter.BackoffPolicyExponential),
		waiter.WithBackoffCoefficient(1.5),
		waiter.WithBackoffExponentialMaxInterval(15 * time.Second),
		waiter.WithLogger(logger),
	}

	// Wait for all services in parallel
	fmt.Println("Waiting for all required services to be available...")
	err := waiter.WaitParallelContext(ctx, checkers, waitOptions...)
	if err != nil {
		log.Fatalf("Failed waiting for services: %v", err)
	}

	fmt.Println("All services are ready!")

	// Continue with application startup
	startApplication()
}

// startApplication is a helper function to start the application
func startApplication() {
	fmt.Println("Starting application...")
	// Your application code here
}
