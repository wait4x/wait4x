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

// Package cname provides the CNAME checker for the Wait4X application.
package cname

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"wait4x.dev/v3/checker"
)

// server is the server to use for the tests
const server = "www.company.info"

// TestSuite is a test suite for CNAME DNS checks
type TestSuite struct {
	suite.Suite
}

// TestCheckExistenceCNAME tests that the CNAME DNS check passes when the expected CNAME domain is present.
func (s *TestSuite) TestCheckExistenceCNAME() {
	d := New(server)
	s.Assert().Nil(d.Check(context.Background()))
}

// TestCorrectCNAME tests that the CNAME DNS check passes when the expected CNAME domain is present.
func (s *TestSuite) TestCorrectCNAME() {
	d := New(server, WithExpectedDomains([]string{"company.info"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestIncorrectCNAME tests that the CNAME DNS check fails when the expected CNAME domain is not present.
func (s *TestSuite) TestIncorrectCNAME() {
	var expectedError *checker.ExpectedError
	d := New(server, WithExpectedDomains([]string{"something wrong"}))
	s.Assert().ErrorAs(d.Check(context.Background()), &expectedError)
}

// TestCustomNSCorrectCNAME tests that the CNAME DNS check passes when the expected CNAME domain is present
// and a custom name server is used.
func (s *TestSuite) TestCustomNSCorrectCNAME() {
	d := New(server, WithNameServer("8.8.8.8:53"), WithExpectedDomains([]string{"company.info"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestRegexCorrectCNAME tests that the CNAME DNS check passes when the expected CNAME domain matches a regular expression.
func (s *TestSuite) TestRegexCorrectCNAME() {
	d := New(server, WithExpectedDomains([]string{"company.*"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestCNAME runs the test suite for CNAME DNS checks.
func TestCNAME(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
