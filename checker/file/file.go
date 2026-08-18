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
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"go.withmatt.com/size"

	"wait4x.dev/v3/checker"
)

// Option configures an File checker
type Option func(e *File)

// File is a command execution checker
type File struct {
	paths              []string
	expectAge          time.Duration
	expectSize         size.Capacity
	expectContentRegex string
	allowAny           bool
}

// New creates a new File checker
func New(paths []string, opts ...Option) checker.Checker {
	e := &File{
		paths: paths,
	}

	// apply the list of options to File
	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithAllowAny changes the validation behaviour to either
// require all paths to match or just one.
func WithAllowAny(a bool) Option {
	return func(f *File) {
		f.allowAny = a
	}
}

// WithExpectContentRegex configures the file content validation rule.
func WithExpectContentRegex(pattern string) Option {
	return func(f *File) {
		f.expectContentRegex = pattern
	}
}

// WithExpectSize configures the max size the files must not exceed.
func WithExpectSize(s size.Capacity) Option {
	return func(f *File) {
		f.expectSize = s
	}
}

// WithExpectAge configures the max modification time the files must not exceed.
func WithExpectAge(a time.Duration) Option {
	return func(f *File) {
		f.expectAge = a
	}
}

// Identity returns the identity of the checker
func (f *File) Identity() (string, error) {
	return strings.Join(f.paths, " "), nil
}

// Check validates the configured paths against the internal validation rules.
func (f *File) Check(ctx context.Context) error {
	var pattern *regexp.Regexp
	var err error

	if f.expectContentRegex != "" {
		pattern, err = regexp.Compile(f.expectContentRegex)
		if err != nil {
			return checker.NewExpectedError(
				"invalid file content validation pattern", err,
				"pattern", f.expectContentRegex,
			)
		}
	}

	errs := make([]error, 0, len(f.paths))
	for _, p := range f.paths {
		err = f.checkFile(p, pattern)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			errs = append(errs, err)

			if !f.allowAny {
				return checker.NewExpectedError(
					"file did not pass validation", err,
					"file", p,
				)
			}
		}
	}

	// we never end up here with errors and allowAny == false
	// allowAny == true and therefor only one file needs to pass
	if len(errs) == len(f.paths) {
		return checker.NewExpectedError(
			"one or more files did not pass validation", nil,
			"errs", len(errs),
		)
	}

	return nil
}

func (f *File) checkFile(p string, pattern *regexp.Regexp) error {
	r, err := os.Open(p)
	if err != nil {
		return err
	}

	defer r.Close()

	s, err := r.Stat()
	if err != nil {
		return err
	}

	if f.expectAge > 0 {
		if err := checkFileAge(s, f.expectAge); err != nil {
			return err
		}
	}

	if f.expectSize > 0 {
		if err := checkFileSize(s, f.expectSize); err != nil {
			return err
		}
	}

	if pattern != nil {
		if err := checkFileContent(r, pattern); err != nil {
			return err
		}
	}

	return nil
}

func checkFileContent(r *os.File, pattern *regexp.Regexp) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if pattern.Match(scanner.Bytes()) {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return fmt.Errorf("File %q content does not match pattern %q", r.Name(), pattern.String())
}

func checkFileSize(i os.FileInfo, l size.Capacity) error {
	size := size.Capacity(i.Size())

	if size <= l {
		return nil
	}

	return fmt.Errorf("File %q is larger (%s) than %s", i.Name(), size, l)
}

func checkFileAge(i os.FileInfo, l time.Duration) error {
	age := time.Since(i.ModTime())

	if age <= l {
		return nil
	}

	return fmt.Errorf("File %q is older (%s) than %s", i.Name(), age, l)
}
