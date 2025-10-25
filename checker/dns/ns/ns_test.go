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

// Package ns provides the NS checker for the Wait4X application.
package ns

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"wait4x.dev/v4/checker"
)

// server is the server to use for the tests
const server = "wait4x.dev"

// TestSuite is a test suite for the DNS nameserver checker
type TestSuite struct {
	suite.Suite
}

// TestCheckExistenceNS tests that the DNS nameserver checker correctly checks the existence of the nameservers for the given domain.
func (s *TestSuite) TestCheckExistenceNS() {
	d := New(server)
	s.Assert().Nil(d.Check(context.Background()))
}

// TestCorrectNS tests that the DNS nameserver checker correctly checks the existence of the expected nameservers for the given domain.
func (s *TestSuite) TestCorrectNS() {
	d := New(server, WithExpectedNameservers([]string{"gordon.ns.cloudflare.com.", "emma.ns.cloudflare.com"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestIncorrectNS tests that the DNS nameserver checker correctly identifies when the expected nameservers
// do not match the actual nameservers for the given domain.
func (s *TestSuite) TestIncorrectNS() {
	var expectedError *checker.ExpectedError
	d := New(server, WithExpectedNameservers([]string{"127.0.0.1"}))
	s.Assert().ErrorAs(d.Check(context.Background()), &expectedError)
}

// TestCustomNSCorrectNS tests that the DNS nameserver checker correctly checks the existence of the expected
// nameservers for the given domain using a custom nameserver.
func (s *TestSuite) TestCustomNSCorrectNS() {
	d := New(server, WithNameServer("8.8.8.8:53"), WithExpectedNameservers([]string{"gordon.ns.cloudflare.com."}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestRegexCorrectNS tests that the DNS nameserver checker correctly checks the existence of the expected
// nameservers for the given domain using a regular expression.
func (s *TestSuite) TestRegexCorrectNS() {
	d := New(server, WithExpectedNameservers([]string{".*.cloudflare.com"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestNS runs the test suite for the DNS nameserver checker.
func TestNS(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
