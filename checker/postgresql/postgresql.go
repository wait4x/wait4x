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
	"fmt"
	"net/url"
	"regexp"

	"wait4x.dev/v3/checker"

	// Needed for the PostgreSQL driver
	_ "github.com/lib/pq"
)

var hidePasswordRegexp = regexp.MustCompile(`^(postgres://[^/:]+):[^:@]+@`)

const (
	tableExistsQuery = "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s')"
)

// Option is a function that configures the PostgreSQL checker
type Option func(p *PostgreSQL)

// PostgreSQL is a PostgreSQL checker
type PostgreSQL struct {
	dsn         string
	tableExists string
}

// New creates a new PostgreSQL checker
func New(dsn string, opts ...Option) checker.Checker {
	p := &PostgreSQL{
		dsn: dsn,
	}

	// apply the list of options to PostgreSQL
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithTableExists configures the table existence check
func WithTableExists(table string) Option {
	return func(p *PostgreSQL) {
		p.tableExists = table
	}
}

// Identity returns the identity of the PostgreSQL checker
func (p *PostgreSQL) Identity() (string, error) {
	u, err := url.Parse(p.dsn)
	if err != nil {
		return "", fmt.Errorf("can't retrieve the checker identity: %w", err)
	}

	return u.Host, nil
}

// Check checks the PostgreSQL connection
func (p *PostgreSQL) Check(ctx context.Context) (err error) {
	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return err
	}

	defer func(db *sql.DB) {
		if dberr := db.Close(); dberr != nil {
			err = dberr
		}
	}(db)

	err = db.PingContext(ctx)
	if err != nil {
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the postgresql server", err,
				"dsn", hidePasswordRegexp.ReplaceAllString(p.dsn, `$1:***@`),
			)
		}

		return err
	}

	// check if the table exists if option has been set
	if p.tableExists != "" {
		query := fmt.Sprintf(tableExistsQuery, p.tableExists)
		var exists bool
		err = db.QueryRowContext(ctx, query).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return checker.NewExpectedError(
				"table does not exist", nil,
				"table", p.tableExists,
			)
		}
	}

	return nil
}
