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

package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/tidwall/gjson"
	"wait4x.dev/v3/checker"
)

// Option configures an HTTP.
type Option func(h *HTTP)

const (
	// DefaultConnectionTimeout is the default connection timeout duration
	DefaultConnectionTimeout = 3 * time.Second
	// DefaultInsecureSkipTLSVerify is the default insecure skip tls verify
	DefaultInsecureSkipTLSVerify = false
	// DefaultNoRedirect is the default auto redirect
	DefaultNoRedirect = false
)

// HTTP represents HTTP checker
type HTTP struct {
	address               string
	timeout               time.Duration
	expectBodyRegex       string
	expectBodyJSON        string
	expectBodyXPath       string
	expectHeader          string
	requestHeaders        http.Header
	requestBody           io.Reader
	expectStatusCode      int
	insecureSkipTLSVerify bool
	noRedirect            bool
	caFile                string
	certFile              string
	keyFile               string
}

// New creates the HTTP checker
func New(address string, opts ...Option) checker.Checker {
	h := &HTTP{
		address:               address,
		timeout:               DefaultConnectionTimeout,
		insecureSkipTLSVerify: DefaultInsecureSkipTLSVerify,
		noRedirect:            DefaultNoRedirect,
	}

	// apply the list of options to HTTP
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// WithTimeout configures a time limit for requests made by the HTTP client
func WithTimeout(timeout time.Duration) Option {
	return func(h *HTTP) {
		h.timeout = timeout
	}
}

// WithExpectBodyRegex configures response body expectation
func WithExpectBodyRegex(regex string) Option {
	return func(h *HTTP) {
		h.expectBodyRegex = regex
	}
}

// WithExpectBodyJSON configures response json expectation
func WithExpectBodyJSON(json string) Option {
	return func(h *HTTP) {
		h.expectBodyJSON = json
	}
}

// WithExpectBodyXPath configures response xpath expectation
func WithExpectBodyXPath(xpath string) Option {
	return func(h *HTTP) {
		h.expectBodyXPath = xpath
	}
}

// WithExpectHeader configures response header expectation
func WithExpectHeader(header string) Option {
	return func(h *HTTP) {
		h.expectHeader = header
	}
}

// WithRequestHeaders configures request header
func WithRequestHeaders(headers http.Header) Option {
	return func(h *HTTP) {
		h.requestHeaders = headers
	}
}

// WithRequestBody configures request body
func WithRequestBody(body io.Reader) Option {
	return func(h *HTTP) {
		h.requestBody = body
	}
}

// WithRequestHeader configures request header
func WithRequestHeader(key string, value []string) Option {
	return func(h *HTTP) {
		if h.requestHeaders == nil {
			h.requestHeaders = http.Header{}
		}
		h.requestHeaders[key] = value
	}
}

// WithExpectStatusCode configures response status code expectation
func WithExpectStatusCode(code int) Option {
	return func(h *HTTP) {
		h.expectStatusCode = code
	}
}

// WithInsecureSkipTLSVerify configures insecure skip tls verify
func WithInsecureSkipTLSVerify(insecureSkipTLSVerify bool) Option {
	return func(h *HTTP) {
		h.insecureSkipTLSVerify = insecureSkipTLSVerify
	}
}

// WithNoRedirect configures auto redirect
func WithNoRedirect(noRedirect bool) Option {
	return func(h *HTTP) {
		h.noRedirect = noRedirect
	}
}

// WithCAFile configures CA file
func WithCAFile(path string) Option {
	return func(h *HTTP) {
		h.caFile = path
	}
}

// WithCertFile configures Cert file
func WithCertFile(path string) Option {
	return func(h *HTTP) {
		h.certFile = path
	}
}

// WithKeyFile configures key file
func WithKeyFile(path string) Option {
	return func(h *HTTP) {
		h.keyFile = path
	}
}

// Identity returns the identity of the checker
func (h *HTTP) Identity() (string, error) {
	return h.address, nil
}

// Check checks HTTP connection
func (h *HTTP) Check(ctx context.Context) (err error) {
	tlsConfig, err := h.getTLSConfig()
	if err != nil {
		return
	}
	httpClient := &http.Client{
		Timeout: h.timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		},
	}

	if h.noRedirect {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	method := http.MethodGet
	if h.requestBody != nil {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, method, h.address, h.requestBody)
	if err != nil {
		return err
	}

	req.Header = h.requestHeaders

	resp, err := httpClient.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			return checker.NewExpectedError(
				"timed out while making an http call", err,
				"timeout", h.timeout,
			)
		} else if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish an http connection", err,
				"address", h.address,
			)
		}

		return err
	}

	defer func(body io.ReadCloser) {
		if cerr := body.Close(); cerr != nil {
			err = cerr
		}
	}(resp.Body)

	if h.expectStatusCode != 0 {
		err := h.checkingStatusCodeExpectation(resp)
		if err != nil {
			return err
		}
	}

	if h.expectBodyRegex != "" {
		err := h.checkingBodyExpectation(resp)
		if err != nil {
			return err
		}
	}

	if h.expectBodyJSON != "" {
		err := h.checkingJSONExpectation(resp)
		if err != nil {
			return err
		}
	}

	if h.expectBodyXPath != "" {
		err := h.checkingXPathExpectation(resp)
		if err != nil {
			return err
		}
	}

	if h.expectHeader != "" {
		return h.checkingHeaderExpectation(resp)
	}

	return nil
}

// getTLSConfig prepares TLS config
func (h *HTTP) getTLSConfig() (*tls.Config, error) {
	cfg := tls.Config{
		InsecureSkipVerify: h.insecureSkipTLSVerify,
	}

	if h.insecureSkipTLSVerify {
		return &cfg, nil
	}

	// Cert and key files.
	if h.certFile != "" || h.keyFile != "" {
		cert, err := tls.LoadX509KeyPair(h.certFile, h.keyFile)
		if err != nil {
			return nil, err
		}
		cfg.Certificates = []tls.Certificate{cert}
	}

	// CA file.
	if h.caFile != "" {
		ca, err := os.ReadFile(h.caFile)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(ca) {
			return nil, errors.New("can't append the CA file")
		}
		cfg.RootCAs = certPool
	}

	return &cfg, nil
}

func (h *HTTP) checkingStatusCodeExpectation(resp *http.Response) error {
	if h.expectStatusCode != resp.StatusCode {
		return checker.NewExpectedError(
			"the status code doesn't expect", nil,
			"actual", resp.StatusCode, "expect", h.expectStatusCode,
		)
	}

	return nil
}

func (h *HTTP) checkingBodyExpectation(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodyString := string(bodyBytes)
	matched, _ := regexp.MatchString(h.expectBodyRegex, bodyString)

	if !matched {
		return checker.NewExpectedError(
			"the body doesn't expect", nil,
			"actual", h.truncateString(bodyString, 50), "expect", h.expectBodyRegex,
		)
	}

	return nil
}

func (h *HTTP) checkingJSONExpectation(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodyString := string(bodyBytes)
	value := gjson.Get(bodyString, h.expectBodyJSON)

	if !value.Exists() {
		return checker.NewExpectedError(
			"the JSON doesn't match", nil,
			"actual", h.truncateString(bodyString, 50), "expect", h.expectBodyJSON,
		)
	}

	return nil
}

func (h *HTTP) checkingXPathExpectation(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodyString := string(bodyBytes)
	doc, err := htmlquery.Parse(strings.NewReader(bodyString))
	if err != nil {
		return err
	}

	node, err := htmlquery.Query(doc, h.expectBodyXPath)
	if err != nil {
		return err
	}
	if node == nil {
		return checker.NewExpectedError(
			"the XPath doesn't match", nil,
			"actual", h.truncateString(bodyString, 50), "expect", h.expectBodyXPath,
		)
	}

	return nil
}

func (h *HTTP) checkingHeaderExpectation(resp *http.Response) error {
	// Key value. e.g. Content-Type=application/json
	expectedHeaderParsed := strings.SplitN(h.expectHeader, "=", 2)
	if len(expectedHeaderParsed) == 2 {
		headerValue := resp.Header.Get(expectedHeaderParsed[0])
		matched, _ := regexp.MatchString(expectedHeaderParsed[1], headerValue)
		if !matched {
			return checker.NewExpectedError(
				"the http header key and value doesn't expect", nil,
				"actual", headerValue, "expect", h.expectHeader,
			)
		}
	}

	// Only key.
	if _, ok := resp.Header[expectedHeaderParsed[0]]; !ok {
		return checker.NewExpectedError(
			"the http header key doesn't expect", nil,
			"actual", resp.Header, "expect", h.expectHeader,
		)
	}

	return nil
}

func (h *HTTP) truncateString(str string, num int) string {
	truncatedStr := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		truncatedStr = str[0:num] + "..."
	}

	return truncatedStr
}
