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
	"bufio"
	"io"
	"regexp"
	"sync"

	"wait4x.dev/v3/checker"
)

type pipeProcessor struct {
	reader  io.Reader
	pattern *regexp.Regexp
}

func newPipeProcessor(pattern string, provider func() (io.ReadCloser, error)) (*pipeProcessor, error) {
	if pattern == "" {
		return nil, nil
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	rdr, err := provider()
	if err != nil {
		return nil, err
	}

	result := &pipeProcessor{
		reader:  rdr,
		pattern: reg,
	}

	return result, nil
}

func (p *pipeProcessor) Process(wg *sync.WaitGroup, result chan<- error) {
	process := bufio.NewScanner(p.reader)
	worker := func(scanner *bufio.Scanner, pattern *regexp.Regexp) {
		defer wg.Done()

		match := false
		for scanner.Scan() {
			if !match && pattern.Match(scanner.Bytes()) {
				match = true
			}
		}

		if err := scanner.Err(); err != nil {
			result <- checker.NewExpectedError(
				"output scanner failed to process data", err,
			)
		}

		if !match {
			result <- checker.NewExpectedError(
				"no matching line found", nil,
				"regexp", pattern.String(),
			)
		}
	}

	wg.Add(1)
	go worker(process, p.pattern)
}
