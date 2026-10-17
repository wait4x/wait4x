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
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
	"mvdan.cc/sh/v3/shell"

	"wait4x.dev/v3/checker"
	dnsA "wait4x.dev/v3/checker/dns/a"
	dnsAAAA "wait4x.dev/v3/checker/dns/aaaa"
	dnsCNAME "wait4x.dev/v3/checker/dns/cname"
	dnsMX "wait4x.dev/v3/checker/dns/mx"
	dnsNS "wait4x.dev/v3/checker/dns/ns"
	dnsTXT "wait4x.dev/v3/checker/dns/txt"
	"wait4x.dev/v3/checker/exec"
	"wait4x.dev/v3/checker/http"
	"wait4x.dev/v3/checker/tcp"
	"wait4x.dev/v3/internal/cmd/temporal"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// Config represents the Wait4X configuration file structure.
type Config struct {
	Timeout                       *time.Duration  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Interval                      *time.Duration  `yaml:"interval,omitempty" json:"interval,omitempty"`
	InvertCheck                   *bool           `yaml:"invert-check,omitempty" json:"invert-check,omitempty"`
	BackoffPolicy                 *string         `yaml:"backoff-policy,omitempty" json:"backoff-policy,omitempty"`
	BackoffCoefficient            *float64        `yaml:"backoff-exponential-coefficient,omitempty" json:"backoff-exponential-coefficient,omitempty"`
	BackoffExponentialMaxInterval *time.Duration  `yaml:"backoff-exponential-max-interval,omitempty" json:"backoff-exponential-max-interval,omitempty"`
	Services                      []ServiceConfig `yaml:"services,omitempty" json:"services,omitempty"`
	Checkers                      []ServiceConfig `yaml:"checkers,omitempty" json:"checkers,omitempty"`
}

// ServiceConfig represents the configuration for a service checker.
type ServiceConfig struct {
	Name     string `yaml:"name,omitempty" json:"name,omitempty"`
	Type     string `yaml:"type,omitempty" json:"type,omitempty"`
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	Checker  string `yaml:"checker,omitempty" json:"checker,omitempty"`

	Target    string   `yaml:"target,omitempty" json:"target,omitempty"`
	Address   string   `yaml:"address,omitempty" json:"address,omitempty"`
	Targets   []string `yaml:"targets,omitempty" json:"targets,omitempty"`
	Addresses []string `yaml:"addresses,omitempty" json:"addresses,omitempty"`

	ConnectionTimeout *time.Duration `yaml:"connection-timeout,omitempty" json:"connection-timeout,omitempty"`

	// TLS options
	InsecureSkipTLSVerify *bool  `yaml:"insecure-skip-tls-verify,omitempty" json:"insecure-skip-tls-verify,omitempty"`
	InsecureTransport     *bool  `yaml:"insecure-transport,omitempty" json:"insecure-transport,omitempty"`
	CAFile                string `yaml:"ca-file,omitempty" json:"ca-file,omitempty"`
	CertFile              string `yaml:"cert-file,omitempty" json:"cert-file,omitempty"`
	KeyFile               string `yaml:"key-file,omitempty" json:"key-file,omitempty"`

	// HTTP options
	ExpectStatusCode *int              `yaml:"expect-status-code,omitempty" json:"expect-status-code,omitempty"`
	ExpectBodyRegex  string            `yaml:"expect-body-regex,omitempty" json:"expect-body-regex,omitempty"`
	ExpectBodyJSON   string            `yaml:"expect-body-json,omitempty" json:"expect-body-json,omitempty"`
	ExpectBodyXPath  string            `yaml:"expect-body-xpath,omitempty" json:"expect-body-xpath,omitempty"`
	ExpectHeader     string            `yaml:"expect-header,omitempty" json:"expect-header,omitempty"`
	RequestHeaders   map[string]string `yaml:"request-headers,omitempty" json:"request-headers,omitempty"`
	RequestHeader    []string          `yaml:"request-header,omitempty" json:"request-header,omitempty"`
	RequestBody      string            `yaml:"request-body,omitempty" json:"request-body,omitempty"`
	NoRedirect       *bool             `yaml:"no-redirect,omitempty" json:"no-redirect,omitempty"`
	H2C              *bool             `yaml:"h2c,omitempty" json:"h2c,omitempty"`

	// Redis options
	ExpectKey string `yaml:"expect-key,omitempty" json:"expect-key,omitempty"`

	// MySQL & PostgreSQL options
	ExpectTable string `yaml:"expect-table,omitempty" json:"expect-table,omitempty"`

	// Temporal options
	Mode                      string `yaml:"mode,omitempty" json:"mode,omitempty"`
	Namespace                 string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	TaskQueue                 string `yaml:"task-queue,omitempty" json:"task-queue,omitempty"`
	ExpectWorkerIdentityRegex string `yaml:"expect-worker-identity-regex,omitempty" json:"expect-worker-identity-regex,omitempty"`

	// Exec options
	Command        string   `yaml:"command,omitempty" json:"command,omitempty"`
	Args           []string `yaml:"args,omitempty" json:"args,omitempty"`
	ExpectExitCode *int     `yaml:"expect-exit-code,omitempty" json:"expect-exit-code,omitempty"`

	// DNS options
	Subtype        string   `yaml:"subtype,omitempty" json:"subtype,omitempty"`
	Nameserver     string   `yaml:"nameserver,omitempty" json:"nameserver,omitempty"`
	ExpectedValues []string `yaml:"expected-values,omitempty" json:"expected-values,omitempty"`
	ExpectedIPs    []string `yaml:"expected-ips,omitempty" json:"expected-ips,omitempty"`
}

// normalizeMapKeys recursively normalizes map keys by converting underscores to hyphens and lowercasing.
func normalizeMapKeys(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(val))
		for k, item := range val {
			normKey := strings.ToLower(strings.ReplaceAll(k, "_", "-"))
			normalized[normKey] = normalizeMapKeys(item)
		}
		return normalized
	case map[interface{}]interface{}:
		normalized := make(map[string]interface{}, len(val))
		for k, item := range val {
			normKey := strings.ToLower(strings.ReplaceAll(fmt.Sprintf("%v", k), "_", "-"))
			normalized[normKey] = normalizeMapKeys(item)
		}
		return normalized
	case []interface{}:
		for i, item := range val {
			val[i] = normalizeMapKeys(item)
		}
		return val
	default:
		return v
	}
}

// ParseConfig decodes raw YAML or JSON bytes into a Config struct.
func ParseConfig(data []byte) (*Config, error) {
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	normalized := normalizeMapKeys(raw)
	normalizedBytes, err := yaml.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("failed to process configuration keys: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(normalizedBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &cfg, nil
}

// LoadConfigFile reads and parses a configuration file from path or stdin.
func LoadConfigFile(configPath string) (*Config, error) {
	var data []byte
	var err error

	if configPath == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read configuration from standard input: %w", err)
		}
	} else {
		data, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read configuration file: %w", err)
		}
	}

	return ParseConfig(data)
}

// BuildCheckers constructs checker instances from a ServiceConfig.
func BuildCheckers(sc ServiceConfig) ([]checker.Checker, error) {
	serviceType := strings.ToLower(sc.Type)
	if serviceType == "" {
		serviceType = strings.ToLower(sc.Protocol)
	}
	if serviceType == "" {
		serviceType = strings.ToLower(sc.Checker)
	}

	if serviceType == "" {
		return nil, errors.New("service type, protocol, or checker is required")
	}

	// Resolve targets
	targets := sc.Targets
	if len(targets) == 0 {
		targets = sc.Addresses
	}
	if len(targets) == 0 && sc.Target != "" {
		targets = []string{sc.Target}
	}
	if len(targets) == 0 && sc.Address != "" {
		targets = []string{sc.Address}
	}

	switch serviceType {
	case "tcp":
		if len(targets) == 0 {
			return nil, errors.New("target address is required for tcp service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			var opts []tcp.Option
			if sc.ConnectionTimeout != nil {
				opts = append(opts, tcp.WithTimeout(*sc.ConnectionTimeout))
			}
			checkers = append(checkers, tcp.New(target, opts...))
		}
		return checkers, nil

	case "http", "https":
		if len(targets) == 0 {
			return nil, errors.New("target address is required for http service")
		}
		if (sc.CertFile != "" && sc.KeyFile == "") || (sc.KeyFile != "" && sc.CertFile == "") {
			return nil, errors.New("both cert-file and key-file should be assigned values, not just one of them")
		}

		var requestHeaders nethttp.Header
		if len(sc.RequestHeaders) > 0 {
			requestHeaders = make(nethttp.Header, len(sc.RequestHeaders))
			for k, v := range sc.RequestHeaders {
				requestHeaders.Set(k, v)
			}
		} else if len(sc.RequestHeader) > 0 {
			rawHTTPHeaders := strings.Join(sc.RequestHeader, "\r\n")
			tpReader := textproto.NewReader(bufio.NewReader(strings.NewReader(rawHTTPHeaders + "\r\n\n")))
			mimeHeaders, err := tpReader.ReadMIMEHeader()
			if err != nil {
				return nil, fmt.Errorf("failed to parse request header: %w", err)
			}
			requestHeaders = nethttp.Header(mimeHeaders)
		}

		var requestBodyReader io.Reader
		if sc.RequestBody != "" {
			requestBodyReader = strings.NewReader(sc.RequestBody)
		}

		var checkers []checker.Checker
		for _, target := range targets {
			var opts []http.Option
			if sc.ExpectStatusCode != nil {
				opts = append(opts, http.WithExpectStatusCode(*sc.ExpectStatusCode))
			}
			if sc.ExpectBodyRegex != "" {
				opts = append(opts, http.WithExpectBodyRegex(sc.ExpectBodyRegex))
			}
			if sc.ExpectBodyJSON != "" {
				opts = append(opts, http.WithExpectBodyJSON(sc.ExpectBodyJSON))
			}
			if sc.ExpectBodyXPath != "" {
				opts = append(opts, http.WithExpectBodyXPath(sc.ExpectBodyXPath))
			}
			if sc.ExpectHeader != "" {
				opts = append(opts, http.WithExpectHeader(sc.ExpectHeader))
			}
			if requestHeaders != nil {
				opts = append(opts, http.WithRequestHeaders(requestHeaders))
			}
			if requestBodyReader != nil {
				opts = append(opts, http.WithRequestBody(requestBodyReader))
			}
			if sc.ConnectionTimeout != nil {
				opts = append(opts, http.WithTimeout(*sc.ConnectionTimeout))
			}
			if sc.InsecureSkipTLSVerify != nil {
				opts = append(opts, http.WithInsecureSkipTLSVerify(*sc.InsecureSkipTLSVerify))
			}
			if sc.NoRedirect != nil {
				opts = append(opts, http.WithNoRedirect(*sc.NoRedirect))
			}
			if sc.CAFile != "" {
				opts = append(opts, http.WithCAFile(sc.CAFile))
			}
			if sc.CertFile != "" {
				opts = append(opts, http.WithCertFile(sc.CertFile))
			}
			if sc.KeyFile != "" {
				opts = append(opts, http.WithKeyFile(sc.KeyFile))
			}
			if sc.H2C != nil {
				opts = append(opts, http.WithH2C(*sc.H2C))
			}

			checkers = append(checkers, http.New(target, opts...))
		}
		return checkers, nil

	case "redis":
		if len(targets) == 0 {
			return nil, errors.New("target address is required for redis service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			chk, err := newRedisChecker(target, sc)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "mysql":
		if len(targets) == 0 {
			return nil, errors.New("target DSN is required for mysql service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			chk, err := newMySQLChecker(target, sc)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "postgresql", "postgres":
		if len(targets) == 0 {
			return nil, errors.New("target DSN is required for postgresql service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			chk, err := newPostgreSQLChecker(target, sc)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "mongodb", "mongo":
		if len(targets) == 0 {
			return nil, errors.New("target DSN is required for mongodb service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			chk, err := newMongoDBChecker(target, sc)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "influxdb":
		if len(targets) == 0 {
			return nil, errors.New("target URL is required for influxdb service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			chk, err := newInfluxDBChecker(target, sc)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "kafka":
		if len(targets) == 0 {
			return nil, errors.New("target address is required for kafka service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			chk, err := newKafkaChecker(target, sc)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "rabbitmq", "amqp":
		if len(targets) == 0 {
			return nil, errors.New("target DSN is required for rabbitmq service")
		}
		var checkers []checker.Checker
		for _, target := range targets {
			chk, err := newRabbitMQChecker(target, sc)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "temporal":
		if len(targets) == 0 {
			return nil, errors.New("target address is required for temporal service")
		}
		var checkers []checker.Checker
		temporalCfg := temporal.TemporalServiceConfig{
			Mode:                      sc.Mode,
			Namespace:                 sc.Namespace,
			TaskQueue:                 sc.TaskQueue,
			ExpectWorkerIdentityRegex: sc.ExpectWorkerIdentityRegex,
			ConnectionTimeout:         sc.ConnectionTimeout,
			InsecureTransport:         sc.InsecureTransport,
			InsecureSkipTLSVerify:     sc.InsecureSkipTLSVerify,
		}
		for _, target := range targets {
			chk, err := temporal.CreateTemporalChecker(target, temporalCfg)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, chk)
		}
		return checkers, nil

	case "exec":
		commandStr := sc.Command
		if commandStr == "" && len(targets) > 0 {
			commandStr = targets[0]
		}

		var command string
		var commandArgs []string

		if commandStr != "" {
			parts, err := shell.Fields(commandStr, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to parse command: %w", err)
			}
			if len(parts) == 0 {
				return nil, errors.New("command is empty")
			}
			command = parts[0]
			if len(parts) > 1 {
				commandArgs = append(commandArgs, parts[1:]...)
			}
		} else if len(sc.Args) > 0 {
			command = sc.Args[0]
			commandArgs = sc.Args[1:]
		} else {
			return nil, errors.New("command is required for exec service")
		}

		if len(sc.Args) > 0 && commandStr != "" {
			commandArgs = append(commandArgs, sc.Args...)
		}

		var opts []exec.Option
		if len(commandArgs) > 0 {
			opts = append(opts, exec.WithArgs(commandArgs))
		}
		if sc.ExpectExitCode != nil {
			opts = append(opts, exec.WithExpectExitCode(*sc.ExpectExitCode))
		}

		return []checker.Checker{exec.New(command, opts...)}, nil

	case "dns", "dns-a", "dns-aaaa", "dns-cname", "dns-mx", "dns-ns", "dns-txt":
		if len(targets) == 0 {
			return nil, errors.New("target domain is required for dns service")
		}

		subtype := strings.ToUpper(sc.Subtype)
		if subtype == "" && strings.HasPrefix(serviceType, "dns-") {
			subtype = strings.ToUpper(strings.TrimPrefix(serviceType, "dns-"))
		}
		if subtype == "" {
			subtype = "A"
		}

		expectedValues := sc.ExpectedValues
		if len(expectedValues) == 0 {
			expectedValues = sc.ExpectedIPs
		}

		var checkers []checker.Checker
		for _, target := range targets {
			switch subtype {
			case "A":
				checkers = append(checkers, dnsA.New(
					target,
					dnsA.WithExpectedIPV4s(expectedValues),
					dnsA.WithNameServer(sc.Nameserver),
				))
			case "AAAA":
				checkers = append(checkers, dnsAAAA.New(
					target,
					dnsAAAA.WithExpectedIPV6s(expectedValues),
					dnsAAAA.WithNameServer(sc.Nameserver),
				))
			case "CNAME":
				var cnameOpts []dnsCNAME.Option
				if len(expectedValues) > 0 {
					cnameOpts = append(cnameOpts, dnsCNAME.WithExpectedDomains(expectedValues))
				}
				if sc.Nameserver != "" {
					cnameOpts = append(cnameOpts, dnsCNAME.WithNameServer(sc.Nameserver))
				}
				checkers = append(checkers, dnsCNAME.New(target, cnameOpts...))
			case "MX":
				var mxOpts []dnsMX.Option
				if len(expectedValues) > 0 {
					mxOpts = append(mxOpts, dnsMX.WithExpectedDomains(expectedValues))
				}
				if sc.Nameserver != "" {
					mxOpts = append(mxOpts, dnsMX.WithNameServer(sc.Nameserver))
				}
				checkers = append(checkers, dnsMX.New(target, mxOpts...))
			case "NS":
				var nsOpts []dnsNS.Option
				if len(expectedValues) > 0 {
					nsOpts = append(nsOpts, dnsNS.WithExpectedNameservers(expectedValues))
				}
				if sc.Nameserver != "" {
					nsOpts = append(nsOpts, dnsNS.WithNameServer(sc.Nameserver))
				}
				checkers = append(checkers, dnsNS.New(target, nsOpts...))
			case "TXT":
				var txtOpts []dnsTXT.Option
				if len(expectedValues) > 0 {
					txtOpts = append(txtOpts, dnsTXT.WithExpectedValues(expectedValues))
				}
				if sc.Nameserver != "" {
					txtOpts = append(txtOpts, dnsTXT.WithNameServer(sc.Nameserver))
				}
				checkers = append(checkers, dnsTXT.New(target, txtOpts...))
			default:
				return nil, fmt.Errorf("unsupported dns subtype: %s", subtype)
			}
		}
		return checkers, nil

	default:
		return nil, fmt.Errorf("unsupported service type: %s", serviceType)
	}
}

// runConfig loads the configuration file and executes parallel checks.
func runConfig(cmd *cobra.Command, configPath string) error {
	cfg, err := LoadConfigFile(configPath)
	if err != nil {
		return err
	}

	services := cfg.Services
	if len(services) == 0 {
		services = cfg.Checkers
	}
	if len(services) == 0 {
		return errors.New("no services or checkers specified in configuration")
	}

	var allCheckers []checker.Checker
	for _, sc := range services {
		chkList, err := BuildCheckers(sc)
		if err != nil {
			return err
		}
		allCheckers = append(allCheckers, chkList...)
	}

	if len(allCheckers) == 0 {
		return errors.New("no checkers created from configuration")
	}

	timeout := contextutil.GetTimeout(cmd.Context())
	if !cmd.Flags().Changed("timeout") && cfg.Timeout != nil {
		timeout = *cfg.Timeout
	}

	interval := contextutil.GetInterval(cmd.Context())
	if !cmd.Flags().Changed("interval") && cfg.Interval != nil {
		interval = *cfg.Interval
	}

	invertCheck := contextutil.GetInvertCheck(cmd.Context())
	if !cmd.Flags().Changed("invert-check") && cfg.InvertCheck != nil {
		invertCheck = *cfg.InvertCheck
	}

	backoffPolicy := contextutil.GetBackoffPolicy(cmd.Context())
	if !cmd.Flags().Changed("backoff-policy") && cfg.BackoffPolicy != nil {
		backoffPolicy = *cfg.BackoffPolicy
	}

	backoffCoefficient := contextutil.GetBackoffCoefficient(cmd.Context())
	if !cmd.Flags().Changed("backoff-exponential-coefficient") && cfg.BackoffCoefficient != nil {
		backoffCoefficient = *cfg.BackoffCoefficient
	}

	backoffExpMaxInterval := contextutil.GetBackoffExponentialMaxInterval(cmd.Context())
	if !cmd.Flags().Changed("backoff-exponential-max-interval") && cfg.BackoffExponentialMaxInterval != nil {
		backoffExpMaxInterval = *cfg.BackoffExponentialMaxInterval
	}

	backoffPolicyValues := []string{waiter.BackoffPolicyExponential, waiter.BackoffPolicyLinear}
	if !contains(backoffPolicyValues, backoffPolicy) {
		return fmt.Errorf("--backoff-policy must be one of %v", backoffPolicyValues)
	}

	if backoffPolicy == waiter.BackoffPolicyExponential && backoffExpMaxInterval < interval {
		return fmt.Errorf("--backoff-exponential-max-interval must be greater than --interval")
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	return waiter.WaitParallelContext(
		cmd.Context(),
		allCheckers,
		waiter.WithTimeout(timeout),
		waiter.WithInterval(interval),
		waiter.WithInvertCheck(invertCheck),
		waiter.WithBackoffPolicy(backoffPolicy),
		waiter.WithBackoffCoefficient(backoffCoefficient),
		waiter.WithBackoffExponentialMaxInterval(backoffExpMaxInterval),
		waiter.WithLogger(logger),
	)
}

// NewConfigCommand creates a new config subcommand.
func NewConfigCommand() *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "config [CONFIG_FILE] [flags] [-- command [args...]]",
		Short: "Load services and checks from a configuration file",
		Long:  "Load services and checks from a YAML or JSON configuration file and wait for all of them to be ready in parallel.",
		Args: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")
			if configFile == "" && len(args) < 1 {
				return errors.New("CONFIG_FILE argument or --config flag is required")
			}
			return nil
		},
		Example: `
  # Load configuration from a YAML file
  wait4x config wait4x.yaml

  # Load configuration using the flag
  wait4x config --config wait4x.yaml

  # Load configuration from stdin
  cat wait4x.yaml | wait4x config -

  # Execute a command after all services are ready
  wait4x config wait4x.yaml -- echo "All services are ready"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")
			if configFile == "" && len(args) > 0 {
				configFile = args[0]
			}
			return runConfig(cmd, configFile)
		},
	}

	return configCommand
}
