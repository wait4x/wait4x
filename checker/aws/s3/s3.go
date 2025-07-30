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

// Package s3 provides the S3 bucket checker for the Wait4X application.
package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"wait4x.dev/v4/checker"
)

// Checker is a struct that implements the checker.Checker interface
type Checker struct {
	bucketName string
	region     string
	client     *s3.Client
}

// New creates a new S3 bucket checker.
func New(bucketPath string, opts ...Option) *Checker {
	bucketName := parseBucketName(bucketPath)

	checker := &Checker{
		bucketName: bucketName,
	}

	for _, opt := range opts {
		opt(checker)
	}

	return checker
}

// parseBucketName extracts the bucket name from a given S3 bucket path.
func parseBucketName(bucketPath string) string {
	if strings.HasPrefix(bucketPath, "s3://") {
		u, err := url.Parse(bucketPath)
		if err != nil {
			return bucketPath
		}
		return u.Host
	}

	if after, ok := strings.CutPrefix(bucketPath, "arn:aws:s3:::"); ok {
		return after
	}

	return bucketPath
}

// Identity returns the identity of the checker, which is the bucket name.
func (c *Checker) Identity() (string, error) {
	return c.bucketName, nil
}

// Check verifies if the specified S3 bucket exists and is accessible.
func (c *Checker) Check(ctx context.Context) error {
	if c.client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("unable to load SDK config: %w", err)
		}

		if c.region != "" {
			cfg.Region = c.region
		}

		c.client = s3.NewFromConfig(cfg)
	}

	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucketName),
	})
	if err != nil {
		return checker.NewExpectedError(fmt.Sprintf("bucket %s does not exist or is not accessible", c.bucketName), err)
	}

	return nil
}

// Option is a functional option type for configuring the S3 checker.
type Option func(*Checker)

// WithClient sets a custom S3 client for the checker.
func WithClient(client *s3.Client) Option {
	return func(c *Checker) {
		c.client = client
	}
}
