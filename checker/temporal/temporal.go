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

// Package temporal provides the Temporal checker for the Wait4X application.
package temporal

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"os"
	"regexp"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"wait4x.dev/v4/checker"
)

// Option configures a Temporal checker
type Option func(t *Temporal)

// CheckMode specifies the check mode
type CheckMode string

const (
	// DefaultConnectionTimeout is the default connection timeout duration
	DefaultConnectionTimeout = 3 * time.Second

	// DefaultInsecureTransport is the default insecure transport security
	DefaultInsecureTransport = false

	// DefaultInsecureSkipTLSVerify is the default insecure skip tls verify
	DefaultInsecureSkipTLSVerify = false

	// CheckModeServer is the "server" check mode
	CheckModeServer = "server"

	// CheckModeWorker is the "worker" check mode
	CheckModeWorker = "worker"
)

var (
	// ErrInvalidMode defines invalid mode error
	ErrInvalidMode = errors.New("invalid checkMode provided")
	// ErrNoNamespace defines no namespace error
	ErrNoNamespace = errors.New(`no namespace provided (use temporal.WithNamespace("__namespace__"))`)
	// ErrNoTaskQueue defines no task queue error
	ErrNoTaskQueue = errors.New(`no task queue provided (use temporal.WithTaskQueue("__task_queue__"))`)
)

// Temporal is a Temporal checker
type Temporal struct {
	checkMode                 CheckMode
	target                    string
	timeout                   time.Duration
	namespace                 string
	taskQueue                 string
	insecureTransport         bool
	insecureSkipTLSVerify     bool
	expectWorkerIdentityRegex string
}

// New creates a new Temporal checker
func New(checkMode CheckMode, target string, opts ...Option) checker.Checker {
	t := &Temporal{
		checkMode:             checkMode,
		target:                target,
		timeout:               DefaultConnectionTimeout,
		insecureTransport:     DefaultInsecureTransport,
		insecureSkipTLSVerify: DefaultInsecureSkipTLSVerify,
	}

	// apply the list of options to Temporal
	for _, opt := range opts {
		opt(t)
	}

	return t
}

// WithTimeout configures a timeout for maximum amount of time a dial will wait for a GRPC connection to complete
func WithTimeout(timeout time.Duration) Option {
	return func(t *Temporal) {
		t.timeout = timeout
	}
}

// WithNamespace configures the Temporal namespace that is mandatory for the CheckModeWorker
func WithNamespace(namespace string) Option {
	return func(t *Temporal) {
		t.namespace = namespace
	}
}

// WithTaskQueue configures the Temporal task queue that is mandatory for the CheckModeWorker
func WithTaskQueue(taskQueue string) Option {
	return func(t *Temporal) {
		t.taskQueue = taskQueue
	}
}

// WithInsecureTransport disables transport security
func WithInsecureTransport(insecureTransport bool) Option {
	return func(t *Temporal) {
		t.insecureTransport = insecureTransport
	}
}

// WithInsecureSkipTLSVerify configures insecure skip tls verify
func WithInsecureSkipTLSVerify(insecureSkipTLSVerify bool) Option {
	return func(t *Temporal) {
		t.insecureSkipTLSVerify = insecureSkipTLSVerify
	}
}

// WithExpectWorkerIdentityRegex configures worker (Poller) identity expectation that is mandatory for the CheckModeWorker
func WithExpectWorkerIdentityRegex(expectWorkerIdentityRegex string) Option {
	return func(t *Temporal) {
		t.expectWorkerIdentityRegex = expectWorkerIdentityRegex
	}
}

// Identity returns the identity of the Temporal checker
func (t *Temporal) Identity() (string, error) {
	return t.target, nil
}

// Check checks the Temporal connection
func (t *Temporal) Check(ctx context.Context) (err error) {
	conn, err := t.getGRPCConn()
	if err != nil {
		return err
	}
	defer func(conn *grpc.ClientConn) {
		if connErr := conn.Close(); connErr != nil {
			err = connErr
		}
	}(conn)
	switch t.checkMode {
	case CheckModeWorker:
		if t.namespace == "" {
			return ErrNoNamespace
		}

		if t.taskQueue == "" {
			return ErrNoTaskQueue
		}

		return t.checkWorker(ctx, conn)

	case CheckModeServer:
		return t.checkServer(ctx, conn)

	default:
		return ErrInvalidMode
	}
}

// getGRPCConn gets a GRPC connection
func (t *Temporal) getGRPCConn() (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: t.timeout}
			return d.DialContext(ctx, "tcp", addr)
		}),
	}

	if t.insecureTransport {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(
			opts,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: t.insecureSkipTLSVerify})),
		)
	}

	conn, err := grpc.NewClient(t.target, opts...)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, checker.NewExpectedError("timed out while making a grpc call", err)
		} else if checker.IsConnectionRefused(err) {
			return nil, checker.NewExpectedError("failed to establish a grpc connection", err)
		}

		return nil, err
	}

	return conn, nil
}

// checkServer checks the Temporal server
func (t *Temporal) checkServer(ctx context.Context, conn grpc.ClientConnInterface) error {
	healthClient := grpc_health_v1.NewHealthClient(conn)
	req := &grpc_health_v1.HealthCheckRequest{
		Service: "temporal.api.workflowservice.v1.WorkflowService",
	}

	resp, err := healthClient.Check(ctx, req)
	if err != nil {
		return checker.NewExpectedError("failed to health check", err)
	}

	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return checker.NewExpectedError(
			"health check returned unhealthy",
			nil,
			"status", resp.GetStatus(),
		)
	}

	return nil
}

// checkWorker checks the Temporal worker
func (t *Temporal) checkWorker(ctx context.Context, conn grpc.ClientConnInterface) error {
	client := workflowservice.NewWorkflowServiceClient(conn)
	req := &workflowservice.DescribeTaskQueueRequest{
		Namespace: t.namespace,
		TaskQueue: &taskqueue.TaskQueue{
			Name: t.taskQueue,
		},
		TaskQueueType: enums.TASK_QUEUE_TYPE_WORKFLOW,
	}

	resp, err := client.DescribeTaskQueue(ctx, req)
	if err != nil {
		return checker.NewExpectedError(
			"failed to describe the task queue",
			err,
		)
	}

	if len(resp.Pollers) == 0 {
		return checker.NewExpectedError("no worker (poller) registered", nil)
	}

	if t.expectWorkerIdentityRegex != "" {
		workerMatched := false
		for _, poller := range resp.Pollers {
			matched, err := regexp.MatchString(t.expectWorkerIdentityRegex, poller.Identity)
			if err != nil {
				return checker.NewExpectedError("failed to match string", err)
			}

			if matched {
				workerMatched = true
			}
		}

		if !workerMatched {
			return checker.NewExpectedError(
				"the worker (poller) hasn't registered yet",
				nil,
				"pattern", t.expectWorkerIdentityRegex,
			)
		}
	}

	return nil
}
