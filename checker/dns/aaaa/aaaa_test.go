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

// Package aaaa provides functionality for checking the AAAA records of a domain.
package aaaa

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"wait4x.dev/v3/checker"
)

const server = "wait4x.dev"

// TestSuite is a test suite for the AAAA DNS checker.
type TestSuite struct {
	suite.Suite
}

// TestCheckExistenceAAAA tests that the AAAA DNS checker correctly checks for the
// existence of the expected AAAA record for the given server.
func (s *TestSuite) TestCheckExistenceAAAA() {
	d := New(server)
	s.Assert().Nil(d.Check(context.Background()))
}

// TestCorrectAAAA tests that the AAAA DNS checker correctly checks for the
// existence of the expected AAAA record for the given server.
func (s *TestSuite) TestCorrectAAAA() {
	d := New(server, WithExpectedIPV6s([]string{"2606:4700:3034::6815:591"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestIncorrectAAAA tests that the AAAA DNS checker correctly handles the case where
// the expected AAAA record does not match the actual AAAA record for the given server.
func (s *TestSuite) TestIncorrectAAAA() {
	var expectedError *checker.ExpectedError
	d := New(server, WithExpectedIPV6s([]string{"127.0.0.1"}))
	s.Assert().ErrorAs(d.Check(context.Background()), &expectedError)
}

// TestCustomNSCorrectAAAA tests that the AAAA DNS checker correctly checks for the
// existence of the expected AAAA record for the given server using a custom name server.
func (s *TestSuite) TestCustomNSCorrectAAAA() {
	d := New(server, WithNameServer("8.8.8.8:53"), WithExpectedIPV6s([]string{"2606:4700:3034::6815:591"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestAAAA runs the test suite for the AAAA DNS checker.
func TestAAAA(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
