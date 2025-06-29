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

package checker

import (
	"context"
	"github.com/stretchr/testify/mock"
)

// MockChecker is the struct that mocks the Checker.
type MockChecker struct {
	mock.Mock
}

// Identity mocks the checker's identity
func (mc *MockChecker) Identity() (string, error) {
	args := mc.Called()

	return args.String(0), args.Error(1)
}

// Check mocks the checker's check
func (mc *MockChecker) Check(ctx context.Context) error {
	args := mc.Called(ctx)

	return args.Error(0)
}
