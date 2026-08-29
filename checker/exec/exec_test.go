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

// Package exec provides the Exec checker for the Wait4X application.
package exec

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheck tests the Exec checker against various test scenarios.
func TestCheck(t *testing.T) {
	const runner = "testdata/exec/runner.sh"

	tests := map[string]struct {
		haveOptions []Option
		wantErr     string
	}{
		"minimal": {},
		"unexpected exit code": {
			haveOptions: []Option{
				WithArgs([]string{"124"}), // instruct the runner to exit with code 124
			},
			wantErr: "command exited with unexpected code",
		},
		"expected exit code": {
			haveOptions: []Option{
				WithArgs([]string{"1"}),
				WithExpectExitCode(1),
			},
		},
		"stdout match": {
			haveOptions: []Option{
				WithExpectStdoutRegex("second line of stdout"),
			},
		},
		"stdout mismatch": {
			haveOptions: []Option{
				WithExpectStdoutRegex("stderr"),
			},
			wantErr: "no matching line found",
		},
		"stderr match": {
			haveOptions: []Option{
				WithExpectStderrRegex("second line of stderr"),
			},
		},
		"stderr mismatch": {
			haveOptions: []Option{
				WithExpectStderrRegex("stdout"),
			},
			wantErr: "no matching line found",
		},
		"both match": {
			haveOptions: []Option{
				WithExpectStdoutRegex("second line of stdout"),
				WithExpectStderrRegex("second line of stderr"),
			},
		},
		"both mismatch": {
			haveOptions: []Option{
				WithExpectStdoutRegex("stderr"),
				WithExpectStderrRegex("stdout"),
			},
			wantErr: "no matching line found",
		},
	}

	for name, test := range tests {
		scenario := func(t *testing.T) {
			subject := New(runner, test.haveOptions...)
			gotErr := subject.Check(context.TODO())

			if test.wantErr == "" {
				assert.Nil(t, gotErr)
			} else {
				assert.ErrorContains(t, gotErr, test.wantErr)
			}
		}

		t.Run(name, scenario)
	}
}
