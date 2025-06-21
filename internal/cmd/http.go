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
	"bufio"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/http"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewHTTPCommand creates a new http sub-command
func NewHTTPCommand() *cobra.Command {
	httpCommand := &cobra.Command{
		Use:   "http ADDRESS... [flags] [-- command [args...]]",
		Short: "Check HTTP connection",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the http command")
			}

			_, err := url.Parse(args[0])
			if err != nil {
				return err
			}

			return nil
		},
		Example: `
  # If you want checking just http connection
  wait4x http https://ifconfig.co

  # If you want checking http connection and expect specify http status code
  wait4x http https://ifconfig.co --expect-status-code 200

  # If you want to check a http response header
  # NOTE: the value in the expected header is regex.
  # Sample response header: Authorization Token 1234ABCD
  # You can match it by these ways:

  # Full key value:
  wait4x http https://ifconfig.co --expect-header "Authorization=Token 1234ABCD"

  # Value starts with:
  wait4x http https://ifconfig.co --expect-header "Authorization=Token"

  # Regex value:
  wait4x http https://ifconfig.co --expect-header "Authorization=Token\s.+"

  # Body JSON:
  wait4x http https://ifconfig.co/json --expect-body-json "user_agent.product"
  To know more about JSON syntax https://github.com/tidwall/gjson/blob/master/SYNTAX.md

  # Body XPath:
  wait4x http https://www.kernel.org/ --expect-body-xpath "//*[@id="tux-gear"]"

  # Request headers:
  wait4x http https://ifconfig.co --request-header "Content-Type: application/json" --request-header "Authorization: Token 123"

  # Post request (application/x-www-form-urlencoded):
  wait4x http https://httpbin.org/post --request-header "Content-Type: application/x-www-form-urlencoded" --request-body 'key=value&name=test'

  # Post request (application/json):
  wait4x http https://httpbin.org/post --request-header "Content-Type: application/json" --request-body '{"key": "value", "name": "test"}'

  # Disable auto redirect
  wait4x http https://www.wait4x.dev --expect-status-code 301 --no-redirect

  # Enable exponential backoff retry
  wait4x http https://ifconfig.co --expect-status-code 200 --backoff-policy exponential --backoff-exponential-max-interval 120s --timeout 120s

  # Self-signed certificates
  wait4x http https://www.wait4x.dev --cert-file /path/to/certfile --key-file /path/to/keyfile

  # CA file
  wait4x http https://www.wait4x.dev --ca-file /path/to/cafile`,
		RunE: runHTTP,
	}

	httpCommand.Flags().Int("expect-status-code", 0, "Expect response code e.g. 200, 204, ... .")
	httpCommand.Flags().String("expect-body-regex", "", "Expect response body pattern.")
	httpCommand.Flags().String("expect-body-json", "", "Expect response body JSON pattern.")
	httpCommand.Flags().String("expect-body-xpath", "", "Expect response body XPath pattern.")
	httpCommand.Flags().String("expect-header", "", "Expect response header pattern.")
	httpCommand.Flags().StringArray("request-header", nil, "User request headers.")
	httpCommand.Flags().String("request-body", "", "User request body.")
	httpCommand.Flags().
		Duration("connection-timeout", http.DefaultConnectionTimeout, "Http connection timeout, The timeout includes connection time, any redirects, and reading the response body.")
	httpCommand.Flags().
		Bool("insecure-skip-tls-verify", http.DefaultInsecureSkipTLSVerify, "Skips tls certificate checks for the HTTPS request.")
	httpCommand.Flags().
		Bool("no-redirect", http.DefaultNoRedirect, "Do not follow HTTP 3xx redirects.")
	httpCommand.Flags().
		String("ca-file", "", "Use this CA bundle to authenticate certificates of servers with HTTPS enabled.")
	httpCommand.Flags().
		String("cert-file", "", "Utilize this SSL certificate file to identify the HTTPS client.")
	httpCommand.Flags().
		String("key-file", "", "Utilize this SSL key file to identify the HTTPS client.")

	return httpCommand
}

func runHTTP(cmd *cobra.Command, args []string) error {
	expectStatusCode, _ := cmd.Flags().GetInt("expect-status-code")
	expectBodyRegex, _ := cmd.Flags().GetString("expect-body-regex")
	expectBodyJSON, _ := cmd.Flags().GetString("expect-body-json")
	expectBodyXPath, _ := cmd.Flags().GetString("expect-body-xpath")
	expectHeader, _ := cmd.Flags().GetString("expect-header")
	requestRawHeaders, _ := cmd.Flags().GetStringArray("request-header")
	requestBody, _ := cmd.Flags().GetString("request-body")
	connectionTimeout, _ := cmd.Flags().GetDuration("connection-timeout")
	insecureSkipTLSVerify, _ := cmd.Flags().GetBool("insecure-skip-tls-verify")
	noRedirect, _ := cmd.Flags().GetBool("no-redirect")
	caFile, _ := cmd.Flags().GetString("ca-file")
	certFile, _ := cmd.Flags().GetString("cert-file")
	keyFile, _ := cmd.Flags().GetString("key-file")

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return err
	}

	// Validate cert and key files.
	if (certFile != "" && keyFile == "") || (keyFile != "" && certFile == "") {
		return fmt.Errorf(
			"both certFile and keyFile should be assigned values, not just one of them",
		)
	}

	// Convert raw headers (e.g. 'a: b') into a http Header.
	var requestHeaders nethttp.Header
	if len(requestRawHeaders) > 0 {
		rawHTTPHeaders := strings.Join(requestRawHeaders, "\r\n")
		tpReader := textproto.NewReader(
			bufio.NewReader(strings.NewReader(rawHTTPHeaders + "\r\n\n")),
		)
		MIMEHeaders, err := tpReader.ReadMIMEHeader()
		if err != nil {
			return fmt.Errorf("can't parse the request header: %w", err)
		}
		requestHeaders = nethttp.Header(MIMEHeaders)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	// Request body.
	var requestBodyReader io.Reader
	if len(requestBody) != 0 {
		requestBodyReader = strings.NewReader(requestBody)
	}

	checkers := make([]checker.Checker, len(args))
	for i, arg := range args {
		checkers[i] = http.New(arg,
			http.WithExpectStatusCode(expectStatusCode),
			http.WithExpectBodyRegex(expectBodyRegex),
			http.WithExpectBodyJSON(expectBodyJSON),
			http.WithExpectBodyXPath(expectBodyXPath),
			http.WithExpectHeader(expectHeader),
			http.WithRequestHeaders(requestHeaders),
			http.WithRequestBody(requestBodyReader),
			http.WithTimeout(connectionTimeout),
			http.WithInsecureSkipTLSVerify(insecureSkipTLSVerify),
			http.WithNoRedirect(noRedirect),
			http.WithCAFile(caFile),
			http.WithCertFile(certFile),
			http.WithKeyFile(keyFile),
		)
	}

	return waiter.WaitParallelContext(
		cmd.Context(),
		checkers,
		waiter.WithTimeout(contextutil.GetTimeout(cmd.Context())),
		waiter.WithInterval(contextutil.GetInterval(cmd.Context())),
		waiter.WithInvertCheck(contextutil.GetInvertCheck(cmd.Context())),
		waiter.WithBackoffPolicy(contextutil.GetBackoffPolicy(cmd.Context())),
		waiter.WithBackoffCoefficient(contextutil.GetBackoffCoefficient(cmd.Context())),
		waiter.WithBackoffExponentialMaxInterval(
			contextutil.GetBackoffExponentialMaxInterval(cmd.Context()),
		),
		waiter.WithLogger(logger),
	)
}
