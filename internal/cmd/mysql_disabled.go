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

//go:build disable_mysql

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// NewMySQLCommand creates a new mysql sub-command
func NewMysqlCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mysql",
		Short: "Check MySQL connection - this feature is disabled",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("MySQL feature disabled in this build.")
		},
	}
}
