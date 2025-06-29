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

import "fmt"

// ExpectedError defines the expectation error
type ExpectedError struct {
	msg     string
	cause   error
	details []any
}

// NewExpectedError creates the ExpectedError
func NewExpectedError(msg string, cause error, details ...any) error {
	ee := &ExpectedError{
		msg:     msg,
		cause:   cause,
		details: details,
	}

	return ee
}

// Details returns the error details
func (ee *ExpectedError) Details() []any {
	return ee.details
}

func (ee *ExpectedError) Unwrap() error {
	return ee.cause
}

func (ee *ExpectedError) Error() string {
	if ee.cause != nil {
		return fmt.Sprintf("%s, caused by: %s", ee.msg, ee.cause.Error())
	}

	return ee.msg
}
