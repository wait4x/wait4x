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

// Package postgresql provides the PostgreSQL checker for the Wait4X application.
package postgresql

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"wait4x.dev/v4/checker"
)

// PostgreSQLSuite is a test suite for PostgreSQL checker
type PostgreSQLSuite struct {
	suite.Suite
	container *postgres.PostgresContainer
}

// SetupSuite starts a PostgreSQL container
func (s *PostgreSQLSuite) SetupSuite() {
	var err error
	s.container, err = postgres.Run(
		context.Background(),
		"postgres:16-alpine",
		testcontainers.WithLogger(log.TestLogger(s.T())),
		// Postgres restarts after the first "ready" log. Waiting only for the
		// port lets the first query hit a connection reset.
		postgres.BasicWaitStrategies(),
	)

	s.Require().NoError(err)
}

// TearDownSuite stops the PostgreSQL container
func (s *PostgreSQLSuite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoError(err)
}

// TestIdentity tests the identity of the PostgreSQL checker
func (s *PostgreSQLSuite) TestIdentity() {
	chk := New("postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full")
	identity, err := chk.Identity()

	s.Require().NoError(err)
	s.Assert().Equal("1.2.3.4:5432", identity)
}

// TestInvalidIdentity tests the invalid identity of the PostgreSQL checker
func (s *PostgreSQLSuite) TestInvalidIdentity() {
	chk := New("127.0.0.1:5432")
	_, err := chk.Identity()

	s.Assert().ErrorContains(err, "first path segment in URL cannot contain colon")
}

// TestValidConnection tests the valid connection of the PostgreSQL server
func (s *PostgreSQLSuite) TestInvalidConnection() {
	var expectedError *checker.ExpectedError
	// Use localhost:8080 to get immediate connection refused (matches MySQL test pattern)
	chk := New("postgres://bob:secret@localhost:8080/mydb?sslmode=disable")

	s.Assert().ErrorAs(chk.Check(context.Background()), &expectedError)
}

// TestValidAddress tests the valid address of the PostgreSQL server
func (s *PostgreSQLSuite) TestValidAddress() {
	ctx := context.Background()

	endpoint, err := s.container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	chk := New(endpoint)
	s.Assert().Nil(chk.Check(ctx))
}

func (s *PostgreSQLSuite) TestTableNotExists() {
	var expectedError *checker.ExpectedError

	ctx := context.Background()

	endpoint, err := s.container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	chk := New(endpoint, WithExpectTable("not_existing_table"))

	s.Assert().ErrorAs(chk.Check(ctx), &expectedError)
}

func (s *PostgreSQLSuite) TestExpectTable() {
	ctx := context.Background()
	endpoint, err := s.container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	// Create the table from the host. container.Exec's error is nil when psql
	// exits non-zero, so the previous in-container CREATE TABLE was a no-op.
	db, err := sql.Open("postgres", endpoint)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		if cerr := db.Close(); cerr != nil {
			s.T().Errorf("close postgres: %v", cerr)
		}
	})

	_, err = db.ExecContext(ctx, "CREATE TABLE my_table (id INT)")
	s.Require().NoError(err)

	chk := New(endpoint, WithExpectTable("my_table"))
	s.Assert().Nil(chk.Check(ctx))
}

// TestPostgreSQL runs the PostgreSQL test suite
func TestPostgreSQL(t *testing.T) {
	suite.Run(t, new(PostgreSQLSuite))
}
