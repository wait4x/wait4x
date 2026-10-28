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
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/exec"
	"wait4x.dev/v3/checker/http"
	"wait4x.dev/v3/checker/tcp"
	"wait4x.dev/v3/internal/test"
	"wait4x.dev/v3/waiter"
)

// ConfigCommandSuite is a test suite for configuration file loading and validation
type ConfigCommandSuite struct {
	suite.Suite

	rootCmd    *cobra.Command
	configCmd  *cobra.Command
	listener   net.Listener
	port       int
	serverDone chan struct{}
	tempDir    string
}

func (s *ConfigCommandSuite) SetupSuite() {
	var err error
	s.listener, err = net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)

	_, portStr, err := net.SplitHostPort(s.listener.Addr().String())
	s.Require().NoError(err)
	s.port, err = strconv.Atoi(portStr)
	s.Require().NoError(err)

	s.serverDone = make(chan struct{})
	go func() {
		defer close(s.serverDone)
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				return
			}
			if conn != nil {
				_ = conn.Close()
			}
		}
	}()

	s.tempDir, err = os.MkdirTemp("", "wait4x-config-test-*")
	s.Require().NoError(err)
}

func (s *ConfigCommandSuite) TearDownSuite() {
	if s.listener != nil {
		_ = s.listener.Close()
		<-s.serverDone
	}
	if s.tempDir != "" {
		_ = os.RemoveAll(s.tempDir)
	}
}

func (s *ConfigCommandSuite) SetupTest() {
	s.rootCmd = NewRootCommand()
	s.configCmd = NewConfigCommand()
	s.rootCmd.AddCommand(s.configCmd)
	s.rootCmd.AddCommand(NewTCPCommand())
}

// TestParseValidYAMLConfig verifies decoding a complete YAML configuration.
func (s *ConfigCommandSuite) TestParseValidYAMLConfig() {
	yamlContent := `
timeout: 15s
interval: 2s
invert-check: false
backoff-policy: exponential
backoff-exponential-max-interval: 8s
backoff-exponential-coefficient: 1.8
services:
  - name: local-tcp
    type: tcp
    target: 127.0.0.1:9090
    connection-timeout: 4s
  - name: local-http
    type: http
    target: http://127.0.0.1:8080
    expect-status-code: 200
    connection-timeout: 3s
`
	cfg, err := ParseConfig([]byte(yamlContent))
	s.Require().NoError(err)
	s.Require().NotNil(cfg)

	s.Require().NotNil(cfg.Timeout)
	s.Equal(15*time.Second, *cfg.Timeout)
	s.Require().NotNil(cfg.Interval)
	s.Equal(2*time.Second, *cfg.Interval)
	s.Require().NotNil(cfg.InvertCheck)
	s.False(*cfg.InvertCheck)
	s.Require().NotNil(cfg.BackoffPolicy)
	s.Equal("exponential", *cfg.BackoffPolicy)
	s.Require().NotNil(cfg.BackoffExponentialMaxInterval)
	s.Equal(8*time.Second, *cfg.BackoffExponentialMaxInterval)
	s.Require().NotNil(cfg.BackoffCoefficient)
	s.Equal(1.8, *cfg.BackoffCoefficient)

	s.Require().Len(cfg.Services, 2)
	s.Equal("local-tcp", cfg.Services[0].Name)
	s.Equal("tcp", cfg.Services[0].Type)
	s.Equal("127.0.0.1:9090", cfg.Services[0].Target)
	s.Require().NotNil(cfg.Services[0].ConnectionTimeout)
	s.Equal(4*time.Second, *cfg.Services[0].ConnectionTimeout)

	s.Equal("local-http", cfg.Services[1].Name)
	s.Equal("http", cfg.Services[1].Type)
	s.Equal("http://127.0.0.1:8080", cfg.Services[1].Target)
	s.Require().NotNil(cfg.Services[1].ExpectStatusCode)
	s.Equal(200, *cfg.Services[1].ExpectStatusCode)
}

// TestParseValidJSONConfig verifies decoding JSON configurations.
func (s *ConfigCommandSuite) TestParseValidJSONConfig() {
	jsonContent := `{
  "timeout": "10s",
  "interval": "1s",
  "services": [
    {
      "name": "json-service",
      "type": "tcp",
      "target": "127.0.0.1:8080"
    }
  ]
}`
	cfg, err := ParseConfig([]byte(jsonContent))
	s.Require().NoError(err)
	s.Require().NotNil(cfg)
	s.Require().Len(cfg.Services, 1)
	s.Equal("json-service", cfg.Services[0].Name)
	s.Equal("tcp", cfg.Services[0].Type)
	s.Equal("127.0.0.1:8080", cfg.Services[0].Target)
}

// TestParseSnakeCaseKeys verifies that snake_case keys are normalized properly.
func (s *ConfigCommandSuite) TestParseSnakeCaseKeys() {
	yamlContent := `
timeout: 20s
invert_check: true
backoff_policy: exponential
backoff_exponential_coefficient: 2.5
backoff_exponential_max_interval: 12s
services:
  - name: test-snake
    type: http
    target: http://127.0.0.1:8080/test
    expect_status_code: 204
    connection_timeout: 6s
    insecure_skip_tls_verify: true
    no_redirect: true
`
	cfg, err := ParseConfig([]byte(yamlContent))
	s.Require().NoError(err)
	s.Require().NotNil(cfg)
	s.Require().NotNil(cfg.InvertCheck)
	s.True(*cfg.InvertCheck)
	s.Require().NotNil(cfg.BackoffPolicy)
	s.Equal("exponential", *cfg.BackoffPolicy)
	s.Require().NotNil(cfg.BackoffCoefficient)
	s.Equal(2.5, *cfg.BackoffCoefficient)
	s.Require().NotNil(cfg.BackoffExponentialMaxInterval)
	s.Equal(12*time.Second, *cfg.BackoffExponentialMaxInterval)

	s.Require().Len(cfg.Services, 1)
	svc := cfg.Services[0]
	s.Require().NotNil(svc.ExpectStatusCode)
	s.Equal(204, *svc.ExpectStatusCode)
	s.Require().NotNil(svc.ConnectionTimeout)
	s.Equal(6*time.Second, *svc.ConnectionTimeout)
	s.Require().NotNil(svc.InsecureSkipTLSVerify)
	s.True(*svc.InsecureSkipTLSVerify)
	s.Require().NotNil(svc.NoRedirect)
	s.True(*svc.NoRedirect)
}

// TestBuildCheckersTCP verifies TCP checker construction and options.
func (s *ConfigCommandSuite) TestBuildCheckersTCP() {
	timeout := 5 * time.Second
	sc := ServiceConfig{
		Type:              "tcp",
		Target:            "127.0.0.1:8080",
		ConnectionTimeout: &timeout,
	}

	checkers, err := BuildCheckers(sc)
	s.Require().NoError(err)
	s.Require().Len(checkers, 1)
	s.IsType(&tcp.TCP{}, checkers[0])

	id, err := checkers[0].Identity()
	s.Require().NoError(err)
	s.Equal("127.0.0.1:8080", id)
}

// TestBuildCheckersMultipleTargets verifies multiple targets generate multiple checker instances.
func (s *ConfigCommandSuite) TestBuildCheckersMultipleTargets() {
	sc := ServiceConfig{
		Type:    "tcp",
		Targets: []string{"1.1.1.1:53", "8.8.8.8:53"},
	}

	checkers, err := BuildCheckers(sc)
	s.Require().NoError(err)
	s.Require().Len(checkers, 2)
}

// TestBuildCheckersHTTP verifies HTTP checker construction with all flags.
func (s *ConfigCommandSuite) TestBuildCheckersHTTP() {
	code := 200
	skipTLS := true
	noRedirect := true
	h2c := false
	timeout := 4 * time.Second

	sc := ServiceConfig{
		Type:                  "http",
		Target:                "https://example.com/api",
		ExpectStatusCode:      &code,
		ExpectBodyRegex:       "test-regex",
		ExpectBodyJSON:        "status.code",
		ExpectBodyXPath:       "//div[@id='ok']",
		ExpectHeader:          "Content-Type=application/json",
		RequestHeaders:        map[string]string{"Authorization": "Bearer token123"},
		RequestBody:           `{"query":"ping"}`,
		ConnectionTimeout:     &timeout,
		InsecureSkipTLSVerify: &skipTLS,
		NoRedirect:            &noRedirect,
		H2C:                   &h2c,
	}

	checkers, err := BuildCheckers(sc)
	s.Require().NoError(err)
	s.Require().Len(checkers, 1)
	s.IsType(&http.HTTP{}, checkers[0])
}

// TestBuildCheckersHTTPCertMismatch verifies validation when only one of cert or key is specified.
func (s *ConfigCommandSuite) TestBuildCheckersHTTPCertMismatch() {
	sc := ServiceConfig{
		Type:     "http",
		Target:   "https://example.com",
		CertFile: "/path/to/cert",
	}

	_, err := BuildCheckers(sc)
	s.Require().Error(err)
	s.Contains(err.Error(), "both cert-file and key-file should be assigned values")
}

// TestBuildCheckersExec verifies Exec checker construction.
func (s *ConfigCommandSuite) TestBuildCheckersExec() {
	exitCode := 0
	sc := ServiceConfig{
		Type:           "exec",
		Command:        "echo hello world",
		ExpectExitCode: &exitCode,
	}

	checkers, err := BuildCheckers(sc)
	s.Require().NoError(err)
	s.Require().Len(checkers, 1)
	s.IsType(&exec.Exec{}, checkers[0])
}

// TestBuildCheckersDNS verifies DNS checker construction across record subtypes.
func (s *ConfigCommandSuite) TestBuildCheckersDNS() {
	subtypes := []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT"}
	for _, st := range subtypes {
		sc := ServiceConfig{
			Type:           "dns",
			Subtype:        st,
			Target:         "wait4x.dev",
			Nameserver:     "1.1.1.1:53",
			ExpectedValues: []string{"test-val"},
		}

		checkers, err := BuildCheckers(sc)
		s.Require().NoError(err)
		s.Require().Len(checkers, 1)
	}
}

// TestBuildCheckersValidation verifies errors on invalid configuration entries.
func (s *ConfigCommandSuite) TestBuildCheckersValidation() {
	// Missing type
	_, err := BuildCheckers(ServiceConfig{Target: "127.0.0.1:8080"})
	s.Require().Error(err)
	s.Contains(err.Error(), "service type, protocol, or checker is required")

	// Unsupported type
	_, err = BuildCheckers(ServiceConfig{Type: "unsupported_protocol", Target: "127.0.0.1:8080"})
	s.Require().Error(err)
	s.Contains(err.Error(), "unsupported service type")

	// Missing target for TCP
	_, err = BuildCheckers(ServiceConfig{Type: "tcp"})
	s.Require().Error(err)
	s.Contains(err.Error(), "target address is required for tcp service")

	// Missing target for HTTP
	_, err = BuildCheckers(ServiceConfig{Type: "http"})
	s.Require().Error(err)
	s.Contains(err.Error(), "target address is required for http service")

	// Missing command for Exec
	_, err = BuildCheckers(ServiceConfig{Type: "exec"})
	s.Require().Error(err)
	s.Contains(err.Error(), "command is required for exec service")
}

// TestConfigCommandMissingArgument verifies error when no config file is passed.
func (s *ConfigCommandSuite) TestConfigCommandMissingArgument() {
	_, err := test.ExecuteCommand(s.rootCmd, "config")
	s.Require().Error(err)
	s.Contains(err.Error(), "CONFIG_FILE argument or --config flag is required")
}

// TestConfigCommandNonexistentFile verifies error on missing file.
func (s *ConfigCommandSuite) TestConfigCommandNonexistentFile() {
	_, err := test.ExecuteCommand(s.rootCmd, "config", "/nonexistent/path/wait4x.yaml")
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to read configuration file")
}

// TestConfigCommandInvalidYAML verifies error on invalid syntax.
func (s *ConfigCommandSuite) TestConfigCommandInvalidYAML() {
	cfgFile := filepath.Join(s.tempDir, "invalid.yaml")
	err := os.WriteFile(cfgFile, []byte("invalid: yaml: content: ["), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "config", cfgFile)
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to parse configuration")
}

// TestConfigCommandEmptyServices verifies error when no services are listed.
func (s *ConfigCommandSuite) TestConfigCommandEmptyServices() {
	cfgFile := filepath.Join(s.tempDir, "empty.yaml")
	err := os.WriteFile(cfgFile, []byte("timeout: 10s\nservices: []\n"), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "config", cfgFile)
	s.Require().Error(err)
	s.Contains(err.Error(), "no services or checkers specified in configuration")
}

// TestConfigExecutionSuccess verifies end-to-end execution with active TCP server.
func (s *ConfigCommandSuite) TestConfigExecutionSuccess() {
	cfgContent := `
timeout: 5s
interval: 100ms
services:
  - name: test-tcp
    type: tcp
    target: ` + s.listener.Addr().String() + `
`
	cfgFile := filepath.Join(s.tempDir, "valid_tcp.yaml")
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0600)
	s.Require().NoError(err)

	output, err := test.ExecuteCommand(s.rootCmd, "config", cfgFile)
	s.Require().NoError(err)
	_ = output
}

// TestRootFlagConfigExecution verifies root command with --config flag.
func (s *ConfigCommandSuite) TestRootFlagConfigExecution() {
	cfgContent := `
timeout: 5s
interval: 100ms
services:
  - name: test-tcp-root
    type: tcp
    target: ` + s.listener.Addr().String() + `
`
	cfgFile := filepath.Join(s.tempDir, "root_flag_tcp.yaml")
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "--config", cfgFile)
	s.Require().NoError(err)
}

// TestRootShortFlagConfigExecution verifies root command with -c short flag.
func (s *ConfigCommandSuite) TestRootShortFlagConfigExecution() {
	cfgContent := `
timeout: 5s
interval: 100ms
services:
  - name: test-tcp-short
    type: tcp
    target: ` + s.listener.Addr().String() + `
`
	cfgFile := filepath.Join(s.tempDir, "root_short_flag.yaml")
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "-c", cfgFile)
	s.Require().NoError(err)
}

// TestConfigExecutionFailure verifies failure on unreachable address with timeout.
func (s *ConfigCommandSuite) TestConfigExecutionFailure() {
	cfgContent := `
timeout: 1s
interval: 50ms
services:
  - name: unreachable-tcp
    type: tcp
    target: 240.0.0.1:12345
`
	cfgFile := filepath.Join(s.tempDir, "unreachable.yaml")
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "config", cfgFile)
	s.Require().Error(err)
	s.Equal(context.DeadlineExceeded, err)
}

// TestConfigWithPostRunCommand verifies post-run command execution with --.
func (s *ConfigCommandSuite) TestConfigWithPostRunCommand() {
	cfgContent := `
timeout: 5s
interval: 100ms
services:
  - name: valid-tcp
    type: tcp
    target: ` + s.listener.Addr().String() + `
`
	cfgFile := filepath.Join(s.tempDir, "post_run.yaml")
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "config", cfgFile, "--", "echo", "post_run_success")
	s.Require().NoError(err)
}

// TestConfigInvalidBackoffPolicy verifies error on invalid backoff policy value.
func (s *ConfigCommandSuite) TestConfigInvalidBackoffPolicy() {
	cfgContent := `
timeout: 2s
backoff-policy: invalid_policy
services:
  - type: tcp
    target: ` + s.listener.Addr().String() + `
`
	cfgFile := filepath.Join(s.tempDir, "invalid_backoff.yaml")
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "config", cfgFile)
	s.Require().Error(err)
	s.Contains(err.Error(), "--backoff-policy must be one of")
}

// TestConfigInvalidBackoffExponentialInterval verifies interval validation.
func (s *ConfigCommandSuite) TestConfigInvalidBackoffExponentialInterval() {
	cfgContent := `
timeout: 5s
interval: 2s
backoff-policy: exponential
backoff-exponential-max-interval: 500ms
services:
  - type: tcp
    target: ` + s.listener.Addr().String() + `
`
	cfgFile := filepath.Join(s.tempDir, "invalid_interval.yaml")
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0600)
	s.Require().NoError(err)

	_, err = test.ExecuteCommand(s.rootCmd, "config", cfgFile)
	s.Require().Error(err)
	s.Contains(err.Error(), "--backoff-exponential-max-interval must be greater than --interval")
}

// TestNewConfigCommandProperties verifies subcommand metadata.
func (s *ConfigCommandSuite) TestNewConfigCommandProperties() {
	cmd := NewConfigCommand()
	s.Equal("config [CONFIG_FILE] [flags] [-- command [args...]]", cmd.Use)
	s.Equal("Load services and checks from a configuration file", cmd.Short)
	s.Contains(cmd.Long, "Load services and checks from a YAML or JSON configuration file")
	s.Contains(cmd.Example, "wait4x config wait4x.yaml")
}

// TestWaitParallelContextWithCheckers verifies waiter executes multiple checkers in parallel.
func (s *ConfigCommandSuite) TestWaitParallelContextWithCheckers() {
	chk1 := tcp.New(s.listener.Addr().String(), tcp.WithTimeout(2*time.Second))
	chk2 := exec.New("echo", exec.WithArgs([]string{"parallel_test"}))

	err := waiter.WaitParallelContext(
		context.Background(),
		[]checker.Checker{chk1, chk2},
		waiter.WithTimeout(5*time.Second),
		waiter.WithInterval(100*time.Millisecond),
	)
	s.Require().NoError(err)
}

func TestConfigCommandSuite(t *testing.T) {
	suite.Run(t, new(ConfigCommandSuite))
}
