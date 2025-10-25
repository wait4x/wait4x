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

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-logr/stdr"
	"wait4x.dev/v4/checker/grpc"
	"wait4x.dev/v4/waiter"
)

// main demonstrates using the gRPC health checker
func main() {
	// Set up a logger
	stdr.SetVerbosity(4)
	logger := stdr.New(log.Default())

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Example 1: Check overall gRPC server health")
	fmt.Println("============================================")

	// Create gRPC health checker for overall server health
	serverChecker := grpc.New(
		"localhost:50051",
		grpc.WithTimeout(5*time.Second),
		grpc.WithInsecureTransport(true), // Use insecure for local development
	)

	// Wait for gRPC server to be healthy
	fmt.Println("Waiting for gRPC server to be healthy...")
	err := waiter.WaitContext(
		ctx,
		serverChecker,
		waiter.WithTimeout(30*time.Second),
		waiter.WithInterval(2*time.Second),
		waiter.WithBackoffPolicy(waiter.BackoffPolicyExponential),
		waiter.WithBackoffCoefficient(1.5),
		waiter.WithLogger(logger),
	)
	if err != nil {
		log.Fatalf("gRPC server not ready: %v", err)
	}

	fmt.Println("✅ gRPC server is healthy!")
	fmt.Println()

	// Example 2: Check specific service health
	fmt.Println("Example 2: Check specific service health")
	fmt.Println("=========================================")

	serviceChecker := grpc.New(
		"localhost:50051",
		grpc.WithTimeout(5*time.Second),
		grpc.WithInsecureTransport(true),
		grpc.WithServiceName("myapp.UserService"), // Check specific service
	)

	fmt.Println("Waiting for UserService to be healthy...")
	err = waiter.WaitContext(
		ctx,
		serviceChecker,
		waiter.WithTimeout(30*time.Second),
		waiter.WithInterval(2*time.Second),
		waiter.WithLogger(logger),
	)
	if err != nil {
		log.Fatalf("UserService not ready: %v", err)
	}

	fmt.Println("✅ UserService is healthy!")
	fmt.Println()

	// Example 3: Production setup with TLS
	fmt.Println("Example 3: Production gRPC with TLS")
	fmt.Println("====================================")

	prodChecker := grpc.New(
		"api.example.com:443",
		grpc.WithTimeout(10*time.Second),
		// TLS is enabled by default (insecureTransport=false)
		// grpc.WithInsecureSkipTLSVerify(true), // Uncomment to skip cert verification
	)

	fmt.Println("Checking production gRPC service...")
	err = waiter.WaitContext(
		ctx,
		prodChecker,
		waiter.WithTimeout(30*time.Second),
		waiter.WithInterval(5*time.Second),
		waiter.WithLogger(logger),
	)
	if err != nil {
		log.Printf("Production service check: %v", err)
		fmt.Println("⚠️  This is expected if api.example.com is not a real endpoint")
	} else {
		fmt.Println("✅ Production service is healthy!")
	}

	fmt.Println()
	fmt.Println("All examples completed!")
}
