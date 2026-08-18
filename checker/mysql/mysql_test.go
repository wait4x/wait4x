//go:build integration
// +build integration

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

// Package mysql provides the MySQL checker for the Wait4X application.
package mysql

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/mysql"

	"wait4x.dev/v4/checker"
)

// MySQLSuite is a test suite for MySQL checker
type MySQLSuite struct {
	suite.Suite
	container *mysql.MySQLContainer
}

// SetupSuite starts a MySQL container
func (s *MySQLSuite) SetupSuite() {
	var err error
	// Use the module default wait (log "port: 3306  MySQL Community Server").
	// Waiting only for 33060 (X Protocol) lets the first handshake hit EOF.
	s.container, err = mysql.Run(
		context.Background(),
		"mysql:8.0.36",
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)

	s.Require().NoError(err)
}

// TearDownSuite stops the MySQL container
func (s *MySQLSuite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoError(err)
}

// TestIdentity tests the identity of the MySQL checker
func (s *MySQLSuite) TestIdentity() {
	chk := New("user:password@tcp(localhost:3306)/dbname?tls=skip-verify")
	identity, err := chk.Identity()

	s.Require().NoError(err)
	s.Assert().Equal("localhost:3306", identity)
}

// TestInvalidIdentity tests the invalid identity of the MySQL checker
func (s *MySQLSuite) TestInvalidIdentity() {
	chk := New("xxx://127.0.0.1:3306")
	_, err := chk.Identity()

	s.Assert().ErrorContains(err, "default addr for network 'xxx:/' unknown")
}

// TestValidConnection tests the valid connection of the MySQL server
func (s *MySQLSuite) TestInvalidConnection() {
	var expectedError *checker.ExpectedError
	chk := New("user:password@tcp(localhost:8080)/dbname?tls=skip-verify")

	s.Assert().ErrorAs(chk.Check(context.Background()), &expectedError)
}

// TestValidAddress tests the valid address of the MySQL server
func (s *MySQLSuite) TestValidAddress() {
	ctx := context.Background()

	endpoint, err := s.container.ConnectionString(ctx)
	s.Require().NoError(err)

	chk := New(endpoint)
	s.Assert().Nil(chk.Check(ctx))
}

func (s *MySQLSuite) TestTableNotExists() {
	var expectedError *checker.ExpectedError

	ctx := context.Background()

	endpoint, err := s.container.ConnectionString(ctx)
	s.Require().NoError(err)

	chk := New(endpoint, WithExpectTable("not_existing_table"))
	s.Assert().ErrorAs(chk.Check(ctx), &expectedError)
}

func (s *MySQLSuite) TestExpectTable() {
	ctx := context.Background()
	endpoint, err := s.container.ConnectionString(ctx)
	s.Require().NoError(err)

	// Create the table from the host. container.Exec's error is nil when mysql
	// exits non-zero, so an in-container CREATE TABLE can silently fail.
	db, err := sql.Open("mysql", endpoint)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		if cerr := db.Close(); cerr != nil {
			s.T().Errorf("close mysql: %v", cerr)
		}
	})

	_, err = db.ExecContext(ctx, "CREATE TABLE my_table (id INT)")
	s.Require().NoError(err)

	chk := New(endpoint, WithExpectTable("my_table"))
	s.Assert().Nil(chk.Check(ctx))
}

// TestMySQL runs the MySQL test suite
func TestMySQL(t *testing.T) {
	suite.Run(t, new(MySQLSuite))
}
