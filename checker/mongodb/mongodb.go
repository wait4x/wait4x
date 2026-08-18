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
	"errors"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
	"wait4x.dev/v4/checker"
)

var hidePasswordRegexp = regexp.MustCompile(`^(mongodb://[^/:]+):[^:@]+@`)

// MongoDB is a MongoDB checker
type MongoDB struct {
	dsn string
}

// New creates a new MongoDB checker
func New(dsn string) checker.Checker {
	i := &MongoDB{
		dsn: dsn,
	}

	return i
}

// Identity returns the identity of the MongoDB checker
func (m *MongoDB) Identity() (string, error) {
	cops := options.Client().ApplyURI(m.dsn)
	if len(cops.Hosts) == 0 {
		return "", errors.New("can't retrieve the checker identity")
	}

	return strings.Join(cops.Hosts, ","), nil
}

// Check checks the MongoDB connection
func (m *MongoDB) Check(ctx context.Context) (err error) {
	// Creates a new Client and then initializes it using the Connect method.
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(m.dsn))
	if err != nil {
		return err
	}

	defer func(c *mongo.Client, ctx context.Context) {
		if merr := c.Disconnect(ctx); merr != nil {
			err = merr
		}
	}(c, ctx)

	// Ping the primary
	err = c.Ping(ctx, readpref.Primary())
	if err != nil {
		if checker.IsConnectionRefused(err) || errors.Is(err, topology.ErrServerSelectionTimeout) {
			return checker.NewExpectedError(
				"failed to establish a connection to the MongoDB server", err,
				"dsn", hidePasswordRegexp.ReplaceAllString(m.dsn, `$1:***@`),
			)
		}

		return err
	}

	return nil
}
