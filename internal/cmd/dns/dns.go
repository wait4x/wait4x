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
	"github.com/spf13/cobra"
)

// NewDNSCommand creates a new DNS command
func NewDNSCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "dns",
		Long:  "Check DNS records for various types like A, AAAA, CNAME, MX, TXT, and NS",
		Short: "Check DNS records",
	}

	command.PersistentFlags().StringP("nameserver", "n", "", "Nameserver to use for the DNS query (e.g. 8.8.8.8:53)")

	command.AddCommand(NewACommand())
	command.AddCommand(NewAAAACommand())
	command.AddCommand(NewCNAMECommand())
	command.AddCommand(NewMXCommand())
	command.AddCommand(NewTXTCommand())
	command.AddCommand(NewNSCommand())

	return command
}
