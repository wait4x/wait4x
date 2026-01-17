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

// Package file provides the File checker for the Wait4X application.
package file

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.withmatt.com/size"
)

// TestCheck runs various test scenarios against
// the File checker implementation
func TestCheck(t *testing.T) {
	tests := map[string]struct {
		havePaths   []string
		haveOptions []Option

		wantErr string
	}{
		"no files exist": {
			havePaths: []string{"testdata/404.txt"},
			wantErr:   "no such file or directory",
		},
		"some files exist": {
			haveOptions: []Option{
				WithAllowAny(true),
			},
			havePaths: []string{"testdata/404.txt", "testdata/4kb.txt", "testdata/4lines.txt"},
		},
		"content match": {
			haveOptions: []Option{
				WithExpectContentRegex("two"),
			},
			havePaths: []string{"testdata/2lines.txt", "testdata/3lines.txt"},
		},
		"content mismatch": {
			haveOptions: []Option{
				WithExpectContentRegex("three"),
			},
			havePaths: []string{"testdata/1lines.txt", "testdata/2lines.txt"},
			wantErr:   "file did not pass validation",
		},
		"any content match": {
			haveOptions: []Option{
				WithAllowAny(true),
				WithExpectContentRegex("four"),
			},
			havePaths: []string{"testdata/1lines.txt", "testdata/2lines.txt", "testdata/3lines.txt", "testdata/4lines.txt"},
		},
		"any content mismatch": {
			haveOptions: []Option{
				WithAllowAny(true),
				WithExpectContentRegex("three"),
			},
			havePaths: []string{"testdata/1lines.txt", "testdata/2lines.txt"},
			wantErr:   "one or more files did not pass validation",
		},
		"small size": {
			haveOptions: []Option{
				WithExpectSize(2 * size.Kibibyte),
			},
			havePaths: []string{"testdata/1kb.txt"},
		},
		"equal size": {
			haveOptions: []Option{
				WithExpectSize(1 * size.Kibibyte),
			},
			havePaths: []string{"testdata/1kb.txt"},
		},
		"too big": {
			haveOptions: []Option{
				WithExpectSize(1 * size.Kibibyte),
			},
			havePaths: []string{"testdata/2kb.txt"},
			wantErr:   "is larger",
		},
		// this is a potentially flaky test as it assumes
		// the testdata is younger than 10 years
		"fresh file": {
			haveOptions: []Option{
				WithExpectAge(24 * 356 * 10 * time.Hour),
			},
			havePaths: []string{"testdata/2kb.txt"},
		},
		// this is a potentially flaky test as it assumes
		// the testdata is older than 1 second
		"too old": {
			haveOptions: []Option{
				WithExpectAge(1 * time.Second),
			},
			havePaths: []string{"testdata/2kb.txt"},
			wantErr:   "is older",
		},
	}

	for name, test := range tests {
		scenario := func(t *testing.T) {
			subject := New(test.havePaths, test.haveOptions...)
			gotErr := subject.Check(context.TODO())

			if test.wantErr == "" {
				assert.NoError(t, gotErr)
			} else {
				assert.ErrorContains(t, gotErr, test.wantErr)
			}
		}

		t.Run(name, scenario)
	}
}
