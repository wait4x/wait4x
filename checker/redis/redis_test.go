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

package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	redismodule "github.com/testcontainers/testcontainers-go/modules/redis"
	"testing"
	"time"
	"wait4x.dev/v3/checker"
)

// RedisSuite is a test suite for Redis checker
type RedisSuite struct {
	suite.Suite
	container *redismodule.RedisContainer
}

// SetupSuite starts a Redis container
func (s *RedisSuite) SetupSuite() {
	var err error
	s.container, err = redismodule.Run(
		context.Background(),
		"redis:7",
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)

	s.Require().NoError(err)
}

// TearDownSuite stops the Redis container
func (s *RedisSuite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoError(err)
}

// TestIdentity tests the identity of the Redis checker
func (s *RedisSuite) TestIdentity() {
	chk := New("redis://127.0.0.1:8787")
	identity, err := chk.Identity()

	s.Require().NoError(err)
	s.Assert().Equal("127.0.0.1:8787", identity)
}

// TestInvalidIdentity tests the invalid identity of the Redis checker
func (s *RedisSuite) TestInvalidIdentity() {
	chk := New("xxx://127.0.0.1:8787")
	_, err := chk.Identity()

	s.Assert().ErrorContains(err, "invalid URL scheme: xxx")
}

// TestValidConnection tests the valid connection of the Redis server
func (s *RedisSuite) TestInvalidConnection() {
	var expectedError *checker.ExpectedError
	chk := New("redis://127.0.0.1:8787", WithTimeout(5*time.Second))

	s.Assert().ErrorAs(chk.Check(context.Background()), &expectedError)
}

// TestValidAddress tests the valid address of the Redis server
func (s *RedisSuite) TestValidAddress() {
	ctx := context.Background()

	endpoint, err := s.container.ConnectionString(ctx)
	s.Require().NoError(err)

	chk := New(endpoint)
	s.Assert().Nil(chk.Check(ctx))
}

// TestKeyExistence tests the key existence of the Redis server
func (s *RedisSuite) TestKeyExistence() {
	ctx := context.Background()

	endpoint, err := s.container.ConnectionString(ctx)
	s.Require().NoError(err)

	opts, err := redis.ParseURL(endpoint)
	s.Require().NoError(err)

	redisClient := redis.NewClient(opts)
	redisClient.Set(ctx, "Foo", "Bar", time.Hour)

	chk := New(endpoint, WithExpectKey("Foo"))
	s.Assert().Nil(chk.Check(ctx))

	chk = New(endpoint, WithExpectKey("Foo=^B.*$"))
	s.Assert().Nil(chk.Check(ctx))

	var expectedError *checker.ExpectedError
	chk = New(endpoint, WithExpectKey("Foo=^b[A-Z]$"))
	s.Assert().ErrorAs(chk.Check(ctx), &expectedError)

	chk = New(endpoint, WithExpectKey("Bob"))
	s.Assert().ErrorAs(chk.Check(ctx), &expectedError)
}

// TestRedis runs the Redis test suite
func TestRedis(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}
