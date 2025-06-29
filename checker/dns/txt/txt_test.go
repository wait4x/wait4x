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

// Package txt provides the TXT checker for the Wait4X application.
package txt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"wait4x.dev/v3/checker"
)

// server is the server to use for the tests
const server = "wait4x.dev"

// TestSuite is a test suite for the TXT record checker
type TestSuite struct {
	suite.Suite
}

// TestCheckExistenceTXT checks that the TXT record for the specified server exists
func (s *TestSuite) TestCheckExistenceTXT() {
	d := New(server)
	s.Assert().Nil(d.Check(context.Background()))
}

// TestCorrectTXT checks that the TXT record for the specified server has the expected value
func (s *TestSuite) TestCorrectTXT() {
	d := New(server, WithExpectedValues([]string{"v=spf1 include:_spf.mx.cloudflare.net ~all"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestIncorrectTXT checks that the TXT record for the specified server has an incorrect value, and that the expected error is returned.
func (s *TestSuite) TestIncorrectTXT() {
	var expectedError *checker.ExpectedError
	d := New(server, WithExpectedValues([]string{"127.0.0.1"}))
	s.Assert().ErrorAs(d.Check(context.Background()), &expectedError)
}

// TestCustomNSCorrectTXT checks that the TXT record for the specified server has the expected value, using a custom name server.
func (s *TestSuite) TestCustomNSCorrectTXT() {
	d := New(server, WithNameServer("8.8.8.8:53"), WithExpectedValues([]string{"v=spf1 include:_spf.mx.cloudflare.net ~all"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestRegexCorrectTXT checks that the TXT record for the specified server has a value that matches the expected regular expression.
func (s *TestSuite) TestRegexCorrectTXT() {
	d := New(server, WithExpectedValues([]string{".* include:.*"}))
	s.Assert().Nil(d.Check(context.Background()))
}

// TestTXT runs the test suite for the TXT record checker.
func TestTXT(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
