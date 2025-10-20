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

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"wait4x.dev/v3/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestHTTPCommandInvalidArgument(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(rootCmd, "http")

	assert.Equal(t, "ADDRESS is required argument for the http command", err.Error())
}

func TestHTTPCommandInvalidAddress(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(rootCmd, "http", "http://local host")

	assert.Contains(t, err.Error(), "invalid character \" \" in host name")
}

func TestHTTPConnectionSuccess(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(rootCmd, "http", "https://wait4x.dev")

	assert.Nil(t, err)
}

func TestHTTPConnectionSuccessThenExecuteCommand(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(rootCmd, "http", "https://wait4x.dev", "--", "date")

	assert.Nil(t, err)
}

func TestHTTPConnectionFail(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(rootCmd, "http", "http://not-exists-doomain.tld", "-t", "2s")

	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestHTTPRequestHeaderSuccess(t *testing.T) {
	hts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := new(bytes.Buffer)
		for key, value := range r.Header {
			_, err := fmt.Fprintf(resp, "%s=%s,", key, value)
			assert.Nil(t, err)
		}

		_, err := w.Write(resp.Bytes())
		assert.Nil(t, err)
	}))
	defer hts.Close()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(
		rootCmd,
		"http",
		hts.URL,
		"--request-header",
		"X-Foo: value1",
		"--request-header",
		"X-Foo: value2",
		"--request-header",
		"X-Bar: long \n value",
		"--expect-body-regex",
		"(.*X-Foo=\\[value1 value2\\].*X-Bar=\\[long value\\].*)|(.*X-Bar=\\[long value\\].*X-Foo=\\[value1 value2\\].*)",
	)

	assert.Nil(t, err)
}

func TestHTTPRequestHeaderFail(t *testing.T) {
	hts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer hts.Close()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(
		rootCmd,
		"http",
		hts.URL,
		"--request-header",
		"X-Bar: long value\n\r",
	)

	assert.Contains(t, err.Error(), "can't parse the request header")
}

func TestHTTPH2CFlagSuccess(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("NO_PROXY", "*")

	h2cOnly := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			http.Error(w, "h2c required", http.StatusHTTPVersionNotSupported)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewUnstartedServer(h2c.NewHandler(h2cOnly, &http2.Server{}))
	srv.Start()
	defer srv.Close()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewHTTPCommand())

	_, err := test.ExecuteCommand(
		rootCmd,
		"http",
		srv.URL,
		"--h2c",
		"--no-redirect",
		"--expect-status-code", "200",
		"-t", "2s",
	)
	assert.Nil(t, err)
}
