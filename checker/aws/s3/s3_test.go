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

package s3

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/localstack"

	"wait4x.dev/v4/checker"
)

// S3Suite is a test suite for S3 checker
type S3Suite struct {
	suite.Suite
	container     *localstack.LocalStackContainer
	s3Client      *s3.Client
	testBucket    string
	missingBucket string
}

// SetupSuite starts a LocalStack container and creates a test bucket
func (s *S3Suite) SetupSuite() {
	ctx := context.Background()

	var err error
	s.container, err = localstack.Run(
		ctx,
		"localstack/localstack:3.8.1",
		testcontainers.WithLogger(log.TestLogger(s.T())),
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "s3",
		}),
	)
	s.Require().NoError(err)

	endpoint, err := s.container.PortEndpoint(ctx, "4566/tcp", "http")
	s.Require().NoError(err)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               endpoint,
					HostnameImmutable: true,
				}, nil
			})),
	)
	s.Require().NoError(err)

	s.s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	s.testBucket = "test-bucket"
	s.missingBucket = "missing-bucket"

	_, err = s.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.testBucket),
	})
	s.Require().NoError(err)
}

// TearDownSuite stops the LocalStack container
func (s *S3Suite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoError(err)
}

// TestParseBucketName tests the bucket name parsing logic
func (s *S3Suite) TestParseBucketName() {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-bucket", "my-bucket"},
		{"s3://my-bucket", "my-bucket"},
		{"s3://my-bucket/path/to/object", "my-bucket"},
		{"arn:aws:s3:::my-bucket", "my-bucket"},
		{"arn:aws:s3:::my-bucket/path", "my-bucket"},
	}

	for _, tt := range tests {
		s.T().Run(tt.input, func(t *testing.T) {
			result := parseBucketName(tt.input)
			if result != tt.expected {
				t.Errorf("parseBucketName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNew tests the creation of a new S3 checker
func (s *S3Suite) TestNew() {
	checker := New("my-bucket")

	s.Assert().Equal("my-bucket", checker.bucketName)
}

// TestIdentity tests the identity of the S3 checker
func (s *S3Suite) TestIdentity() {
	checker := New("test-bucket")

	identity, err := checker.Identity()
	s.Require().NoError(err)
	s.Assert().Equal("test-bucket", identity)
}

// TestWithClient tests the WithClient option
func (s *S3Suite) TestWithClient() {
	checker := New("my-bucket", WithClient(s.s3Client))

	s.Assert().Equal(s.s3Client, checker.client)
}

// TestValidBucket tests checking an existing bucket
func (s *S3Suite) TestValidBucket() {
	ctx := context.Background()

	checker := New(s.testBucket, WithClient(s.s3Client))
	err := checker.Check(ctx)

	s.Assert().NoError(err)
}

// TestMissingBucket tests checking a non-existent bucket
func (s *S3Suite) TestMissingBucket() {
	ctx := context.Background()

	var expectedError *checker.ExpectedError
	checker := New(s.missingBucket, WithClient(s.s3Client))
	err := checker.Check(ctx)

	s.Assert().ErrorAs(err, &expectedError)
	s.Assert().Contains(err.Error(), "does not exist or is not accessible")
}

// TestBucketPathFormats tests different bucket path formats
func (s *S3Suite) TestBucketPathFormats() {
	ctx := context.Background()

	testCases := []string{
		s.testBucket,
		"s3://" + s.testBucket,
		"arn:aws:s3:::" + s.testBucket,
	}

	for _, bucketPath := range testCases {
		s.T().Run(bucketPath, func(t *testing.T) {
			checker := New(bucketPath, WithClient(s.s3Client))
			err := checker.Check(ctx)
			s.Assert().NoError(err)
		})
	}
}

// TestS3 runs the S3 test suite
func TestS3(t *testing.T) {
	suite.Run(t, new(S3Suite))
}
