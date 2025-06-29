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

// Package main is the main package for the Custom Checker example.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/waiter"
)

// FileChecker checks if a file exists and meets criteria
type FileChecker struct {
	filePath    string
	minSize     int64
	permissions os.FileMode
}

// NewFileChecker creates a new file checker
func NewFileChecker(filePath string, minSize int64, permissions os.FileMode) *FileChecker {
	return &FileChecker{
		filePath:    filePath,
		minSize:     minSize,
		permissions: permissions,
	}
}

// Identity returns the identity of the checker
func (f *FileChecker) Identity() (string, error) {
	return fmt.Sprintf("file(%s)", f.filePath), nil
}

// Check verifies the file exists and meets the criteria
func (f *FileChecker) Check(ctx context.Context) error {
	// Check if context is done
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue checking
	}

	// Check if file exists
	fileInfo, err := os.Stat(f.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return checker.NewExpectedError("file does not exist", err, "path", f.filePath)
		}
		return err
	}

	// Check file size if minimum size specified
	if f.minSize > 0 && fileInfo.Size() < f.minSize {
		return checker.NewExpectedError(
			"file is smaller than expected",
			nil,
			"path", f.filePath,
			"actual_size", fileInfo.Size(),
			"expected_min_size", f.minSize,
		)
	}

	// Check permissions if specified
	if f.permissions != 0 {
		actualPerms := fileInfo.Mode().Perm()
		if actualPerms&f.permissions != f.permissions {
			return checker.NewExpectedError(
				"file has incorrect permissions",
				nil,
				"path", f.filePath,
				"actual_permissions", actualPerms,
				"expected_permissions", f.permissions,
			)
		}
	}

	return nil
}

// main is the main function for the Custom Checker example
func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create our custom file checker
	// This will check that a file:
	// 1. Exists
	// 2. Is at least 1024 bytes (1KB) in size
	// 3. Has read permission for everyone
	fileChecker := NewFileChecker(
		"/tmp/application.log",
		1024,              // 1KB minimum size
		os.FileMode(0444), // Read permission for all
	)

	// Wait for the file to be ready
	fmt.Println("Waiting for log file to be created with correct size and permissions...")
	err := waiter.WaitContext(
		ctx,
		fileChecker,
		waiter.WithTimeout(time.Minute),
		waiter.WithInterval(2*time.Second),
	)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("Timed out waiting for file: %v", err)
		}
		log.Fatalf("Error waiting for file: %v", err)
	}

	fmt.Println("File is ready!")

	// Now we could proceed with reading or processing the file
	processLogFile(fileChecker.filePath)
}

func processLogFile(path string) {
	fmt.Printf("Processing log file at %s...\n", path)
	// Your file processing code here
}
