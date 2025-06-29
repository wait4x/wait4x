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

// Package a provides the A checker for the Wait4X application.
package a

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"wait4x.dev/v3/checker"
)

// server is the server to use for the tests
const server = "wait4x.dev"

// TestSuite is a test suite for the A checker
type TestSuite struct {
	suite.Suite
}

// TestCheckExistenceA tests that the A checker correctly checks for the existence of an A record
func (s *TestSuite) TestCheckExistenceA() {
	d := New(server)
	s.Assert().Nil(d.Check(context.Background()))
}

// TestCorrectA tests that the A checker correctly checks for the existence of an A record with the expected IP addresses.
func (s *TestSuite) TestCorrectA() {
	d := New(server, WithExpectedIPV4s([]string{"172.67.154.180", "127.0.0.1"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestIncorrectA tests that the A checker correctly checks for the existence of an A record with an unexpected IP address.
func (s *TestSuite) TestIncorrectA() {
	var expectedError *checker.ExpectedError
	d := New(server, WithExpectedIPV4s([]string{"127.0.0.1"}))
	s.Assert().ErrorAs(d.Check(context.Background()), &expectedError)
}

// TestCustomNSCorrectA tests that the A checker correctly checks for the existence of an A record
// with the expected IP addresses using a custom name server.
func (s *TestSuite) TestCustomNSCorrectA() {
	d := New(server, WithNameServer("8.8.8.8:53"), WithExpectedIPV4s([]string{"172.67.154.180"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestA is a test function that runs the TestSuite for the A checker.
func TestA(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
