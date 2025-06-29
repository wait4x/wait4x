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

// Package dns provides the DNS checker for the Wait4X application.
package dns

import (
	"net"
	"os"
	"time"

	"github.com/miekg/dns"
)

// DefaultTimeout is the default timeout for DNS requests
var DefaultTimeout = 5 * time.Second

// defaultRR is the default DNS resolver address
var defaultRR = "1.1.1.1:53" // Cloudflare DNS resolver

// RR returns the default DNS resolver address, or the first resolver address
// from the system's resolv.conf file if it exists and is readable
func RR(nameserver string) string {
	if nameserver != "" {
		return nameserver
	}

	// Check if resolv.conf exists and is readable
	if _, err := os.Stat("/etc/resolv.conf"); err != nil {
		return defaultRR
	}

	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil || len(conf.Servers) == 0 {
		return defaultRR
	}

	return net.JoinHostPort(conf.Servers[0], conf.Port)
}
