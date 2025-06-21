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

// TestNewACommand tests the NewACommand function
func TestNewACommand(t *testing.T) {
	cmd := NewACommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "A ADDRESS [-- command [args...]]", cmd.Use)
	assert.Equal(t, []string{"a"}, cmd.Aliases)
}

// TestACommand_NoArgs tests the ACommand with no arguments
func TestACommand_NoArgs(t *testing.T) {
	cmd := NewACommand()
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err)
	assert.Equal(t, "ADDRESS is required argument for the A command", err.Error())
}

// TestACommand_WithArgs tests the ACommand with arguments
func TestACommand_WithArgs(t *testing.T) {
	cmd := NewACommand()
	err := cmd.Args(cmd, []string{"example.com"})
	assert.NoError(t, err)
}

// TestRunA tests the ACommand with different flags and arguments
func TestRunA(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		flags map[string]string
	}{
		{
			name: "basic check",
			args: []string{"example.com"},
		},
		{
			name: "with expected IP",
			args: []string{"example.com"},
			flags: map[string]string{
				"expect-ip": "93.184.216.34",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewACommand()

			for flag, value := range tt.flags {
				err := cmd.Flags().Set(flag, value)
				assert.NoError(t, err)
			}

			err := cmd.Args(cmd, tt.args)
			assert.NoError(t, err)
		})
	}
}
