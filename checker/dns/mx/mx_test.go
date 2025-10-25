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

// Package mx provides the MX checker for the Wait4X application.
package mx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"wait4x.dev/v4/checker"
)

// server is the server to use for the tests
const server = "wait4x.dev"

// TestSuite is a test suite for the MX checker
type TestSuite struct {
	suite.Suite
}

// TestCheckExistenceMX tests that the MX checker correctly checks the existence of MX records for the given server.
func (s *TestSuite) TestCheckExistenceMX() {
	d := New(server)
	s.Assert().Nil(d.Check(context.Background()))
}

// TestCorrectMX tests that the MX checker correctly checks the existence of MX records for the given server with the expected domains.
func (s *TestSuite) TestCorrectMX() {
	d := New(server, WithExpectedDomains([]string{"route1.mx.cloudflare.net", "route2.mx.cloudflare.net"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestIncorrectMX tests that the MX checker correctly identifies when the expected MX records do not exist for the given server.
func (s *TestSuite) TestIncorrectMX() {
	var expectedError *checker.ExpectedError
	d := New(server, WithExpectedDomains([]string{"127.0.0.1"}))
	s.Assert().ErrorAs(d.Check(context.Background()), &expectedError)
}

// TestCustomNSCorrectA tests that the MX checker correctly checks the existence of MX records for the given server
// using a custom name server.
func (s *TestSuite) TestCustomNSCorrectA() {
	d := New(server, WithNameServer("8.8.8.8:53"), WithExpectedDomains([]string{"route1.mx.cloudflare.net"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestRegexCorrectA tests that the MX checker correctly checks the existence of MX records for the given server
// using a regular expression to match the expected domains.
func (s *TestSuite) TestRegexCorrectA() {
	d := New(server, WithExpectedDomains([]string{".*.mx.cloudflare.net"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestMX runs the test suite for the MX checker.
func TestMX(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
