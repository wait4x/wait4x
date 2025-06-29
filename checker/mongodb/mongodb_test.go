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

// Package mongodb provides the MongoDB checker for the Wait4X application.
package mongodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"wait4x.dev/v3/checker"
)

// MongoDBSuite is a test suite for MongoDB checker
type MongoDBSuite struct {
	suite.Suite
	container *mongodb.MongoDBContainer
}

// SetupSuite starts a MongoDB container
func (s *MongoDBSuite) SetupSuite() {
	var err error
	s.container, err = mongodb.Run(
		context.Background(),
		"mongo:6",
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)

	s.Require().NoError(err)
}

// TearDownSuite stops the MongoDB container
func (s *MongoDBSuite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoError(err)
}

// TestIdentity tests the identity of the MongoDB checker
func (s *MongoDBSuite) TestIdentity() {
	chk := New("mongodb://127.0.0.1:27017")
	identity, err := chk.Identity()

	s.Require().NoError(err)
	s.Assert().Equal("127.0.0.1:27017", identity)
}

// TestInvalidIdentity tests the invalid identity of the MongoDB checker
func (s *MongoDBSuite) TestInvalidIdentity() {
	chk := New("xxx://127.0.0.1:3306")
	_, err := chk.Identity()

	s.Assert().ErrorContains(err, "can't retrieve the checker identity")
}

// TestValidConnection tests the invalid connection of the MongoDB server
func (s *MongoDBSuite) TestInvalidConnection() {
	var expectedError *checker.ExpectedError
	chk := New("mongodb://127.0.0.1:8080")

	s.Assert().ErrorAs(chk.Check(context.Background()), &expectedError)
}

// TestValidConnection tests the valid connection of the MongoDB server
func (s *MongoDBSuite) TestValidConnection() {
	ctx := context.Background()

	endpoint, err := s.container.ConnectionString(ctx)
	s.Require().NoError(err)

	chk := New(endpoint)
	s.Assert().Nil(chk.Check(ctx))
}

// TestMongoDB runs the MongoDB test suite
func TestMongoDB(t *testing.T) {
	suite.Run(t, new(MongoDBSuite))
}
