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

package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewAAAACommand tests the NewAAAACommand function
func TestNewAAAACommand(t *testing.T) {
	cmd := NewAAAACommand()

	assert.Equal(t, "AAAA ADDRESS [-- command [args...]]", cmd.Use)
	assert.Equal(t, []string{"aaaa"}, cmd.Aliases)
	assert.Equal(t, "Check DNS AAAA (IPv6) records for a given domain", cmd.Short)

	err := cmd.Args(cmd, []string{})
	assert.Error(t, err)
	assert.Equal(t, "ADDRESS is required argument for the AAAA command", err.Error())

	err = cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)

	flags := cmd.Flags()
	expectIP, err := flags.GetStringArray("expect-ip")
	assert.NoError(t, err)
	assert.Empty(t, expectIP)
}

// TestRunAAAA tests the AAAACommand with different flags and arguments
func TestRunAAAA(t *testing.T) {
	cmd := NewAAAACommand()

	err := cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)

	cmd.Flags().Set("expect-ip", "2606:4700:3033::ac43:9ab4")
	err = cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)

	cmd.Flags().Set("nameserver", "8.8.8.8")
	err = cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)

	cmd.Flags().Set("interval", "1s")
	err = cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)

	cmd.Flags().Set("timeout", "5s")
	err = cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)

	cmd.Flags().Set("invert-check", "true")
	err = cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)
}
