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

// Package main is the main package for the Reverse Checking example.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"wait4x.dev/v3/checker/tcp"
	"wait4x.dev/v3/waiter"
)

// main is the main function for the Reverse Checking example
func main() {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Port to check
	port := "localhost:8080"

	// Create a TCP checker for the port
	tcpChecker := tcp.New(port, tcp.WithTimeout(2*time.Second))

	fmt.Printf("Waiting for port %s to become free...\n", port)

	// Use invert check to wait until the TCP connection fails (port is free)
	err := waiter.WaitContext(
		ctx,
		tcpChecker,
		waiter.WithTimeout(45*time.Second),
		waiter.WithInterval(3*time.Second),
		// The InvertCheck option is key here - it inverts the success condition
		// So we wait until the checker fails (port is closed)
		waiter.WithInvertCheck(true),
	)

	if err != nil {
		log.Fatalf("Failed waiting for port to become free: %v", err)
	}

	fmt.Printf("Port %s is now free!\n", port)

	// Example: Now that the port is free, start our own service on it
	startServiceOnPort(port)
}

// startServiceOnPort is a helper function to start a service on the given port
func startServiceOnPort(port string) {
	fmt.Printf("Starting new service on port %s...\n", port)
	// Your code to start a service on the now-free port
}
