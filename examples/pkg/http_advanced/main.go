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

// Package main is the main package for the HTTP Advanced example.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	httpChecker "wait4x.dev/v4/checker/http"
	"wait4x.dev/v4/waiter"
)

// main is the main function for the HTTP Advanced example
func main() {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create custom HTTP headers
	headers := http.Header{}
	headers.Add("Authorization", "Bearer my-token")
	headers.Add("Content-Type", "application/json")

	// Prepare a request body
	requestBody := strings.NewReader(`{"query": "status"}`)

	// Create an HTTP checker with multiple validations
	checker := httpChecker.New(
		"https://api.example.com/health",
		httpChecker.WithTimeout(5*time.Second),
		httpChecker.WithExpectStatusCode(200),
		httpChecker.WithExpectBodyJSON("status"),             // Check that 'status' field exists in JSON
		httpChecker.WithExpectBodyRegex(`"healthy":\s*true`), // Regex to check response
		httpChecker.WithExpectHeader("Content-Type=application/json"),
		httpChecker.WithRequestHeaders(headers),
		httpChecker.WithRequestBody(requestBody),
		httpChecker.WithInsecureSkipTLSVerify(true), // Skip TLS verification
	)

	// Wait for the API to be available and responding correctly
	fmt.Println("Waiting for API health endpoint...")

	err := waiter.WaitContext(
		ctx,
		checker,
		waiter.WithTimeout(2*time.Minute),
		waiter.WithInterval(5*time.Second),
		waiter.WithBackoffPolicy(waiter.BackoffPolicyExponential),
	)

	if err != nil {
		log.Fatalf("API health check failed: %v", err)
	}

	fmt.Println("API is healthy and ready!")
}
