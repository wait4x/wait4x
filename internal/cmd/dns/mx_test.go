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

// Package dns provides the DNS command-line interface for the Wait4X application.
package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewMXCommand tests the NewMXCommand function
func TestNewMXCommand(t *testing.T) {
	cmd := NewMXCommand()

	assert.Equal(t, "MX ADDRESS [-- command [args...]]", cmd.Use)
	assert.Equal(t, []string{"mx"}, cmd.Aliases)
	assert.Equal(t, "Check DNS MX (mail exchanger) records for a given domain", cmd.Short)

	err := cmd.Args(cmd, []string{})
	assert.EqualError(t, err, "ADDRESS is required argument for the MX command")

	err = cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)

	flags := cmd.Flags()
	expectDomain, err := flags.GetStringArray("expect-domain")
	assert.NoError(t, err)
	assert.Empty(t, expectDomain)
}
