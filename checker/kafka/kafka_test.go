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

// Package kafka provides a Kafka checker.
package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"wait4x.dev/v4/checker"
)

// KafkaSuite is a test suite for Kafka checker
type KafkaSuite struct {
	suite.Suite
	container *kafka.KafkaContainer
}

// SetupSuite starts a Kafka container
func (s *KafkaSuite) SetupSuite() {
	var err error
	s.container, err = kafka.Run(
		context.Background(),
		"confluentinc/confluent-local:7.5.0",
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)

	s.Require().NoError(err)
}

// TearDownSuite stops the Kafka container
func (s *KafkaSuite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoError(err)
}

// TestIdentity tests the identity of the Kafka checker
func (s *KafkaSuite) TestIdentity() {
	chk := New("kafka://127.0.0.1:9093")
	identity, err := chk.Identity()

	s.Require().NoError(err)
	s.Assert().Equal("127.0.0.1:9093", identity)
}

// TestInvalidIdentity tests the invalid identity of the Kafka checker
func (s *KafkaSuite) TestInvalidIdentity() {
	chk := New("xxx://127.0.0.1:3306")
	_, err := chk.Identity()

	s.Assert().ErrorContains(err, "failed to parse DSN")
}

// TestValidConnection tests the invalid connection of the Kafka server
func (s *KafkaSuite) TestInvalidConnection() {
	var expectedError *checker.ExpectedError
	chk := New("kafka://127.0.0.1:8075")

	s.Assert().ErrorAs(chk.Check(context.Background()), &expectedError)
}

// TestValidConnection tests the valid connection of the Kafka server
func (s *KafkaSuite) TestValidConnection() {
	ctx := context.Background()

	bs, err := s.container.Brokers(ctx)

	s.T().Log("Kafka brokers:", bs, err)

	chk := New("kafka://" + bs[0])
	s.Assert().NoError(chk.Check(ctx))
}

// TestKafka runs the Kafka test suite
func TestKafka(t *testing.T) {
	suite.Run(t, new(KafkaSuite))
}
