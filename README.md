<div align="center">
  <img src="logo.png" alt="Wait4X Logo" width="150">
  <h1>Wait4X</h1>
  <p>A lightweight tool to wait for services to be ready</p>

  [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/wait4x/wait4x/ci.yaml?branch=main&style=flat-square)](https://github.com/wait4x/wait4x/actions/workflows/ci.yaml)
  [![Coverage Status](https://img.shields.io/coverallsCoverage/github/wait4x/wait4x?branch=main&style=flat-square)](https://coveralls.io/github/wait4x/wait4x?branch=main)
  [![Go Report Card](https://goreportcard.com/badge/wait4x.dev/v3?style=flat-square)](https://goreportcard.com/report/wait4x.dev/v3)
  [![Docker Pulls](https://img.shields.io/docker/pulls/wait4x/wait4x?logo=docker&style=flat-square)](https://hub.docker.com/r/wait4x/wait4x)
  [![GitHub Downloads](https://img.shields.io/github/downloads/wait4x/wait4x/total?logo=github&style=flat-square)](https://github.com/wait4x/wait4x/releases)
  [![Packaging status](https://img.shields.io/repology/repositories/wait4x?style=flat-square)](https://repology.org/project/wait4x/versions)
  [![Go Reference](https://img.shields.io/badge/reference-007D9C.svg?style=flat-square&logo=go&logoColor=white&labelColor=5C5C5C)](https://pkg.go.dev/wait4x.dev/v3)

</div>

---

## 📑 Table of Contents

- [📋 Overview](#-overview)
- [✨ Features](#-features)
- [📥 Installation](#-installation)
- [🚀 Quick Start](#-quick-start)
- [📖 Detailed Usage](#-detailed-usage)
- [⚙️ Advanced Features](#️-advanced-features)
- [📦 Go Package Usage](#-go-package-usage)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

## 📋 Overview

**Wait4X** is a powerful, zero-dependency tool that waits for services to be ready before continuing. It supports multiple protocols and services, making it an essential component for:

- **CI/CD pipelines** - Ensure dependencies are available before tests run
- **Container orchestration** - Health checking services before application startup
- **Deployment processes** - Verify system readiness before deploying
- **Application initialization** - Validate external service availability
- **Local development** - Simplify localhost service readiness checks

## ✨ Features

| Feature | Description |
|---------|-------------|
| **Multi-Protocol Support** | TCP, HTTP, DNS |
| **Service Integrations** | Redis, MySQL, PostgreSQL, MongoDB, RabbitMQ, InfluxDB, Temporal |
| **Reverse Checking** | Invert checks to find free ports or non-ready services |
| **Parallel Checking** | Check multiple services simultaneously |
| **Exponential Backoff** | Retry with increasing delays to improve reliability |
| **CI/CD Integration** | Designed for automation workflows |
| **Cross-Platform** | Single binary for Linux, macOS, and Windows |
| **Go Package** | Import into your Go applications |
| **Command Execution** | Run commands after successful checks |

## 📥 Installation

<details>
<summary><b>🐳 With Docker</b></summary>

Wait4X provides automatically updated Docker images within Docker Hub:

```bash
# Pull the image
docker pull wait4x/wait4x:latest

# Run the container
docker run --rm wait4x/wait4x:latest --help
```
</details>

<details>
<summary><b>📦 From Binary</b></summary>

Download the appropriate version for your platform from the [releases page](https://github.com/wait4x/wait4x/releases):

**Linux:**
```bash
curl -LO https://github.com/wait4x/wait4x/releases/latest/download/wait4x-linux-amd64.tar.gz
tar -xf wait4x-linux-amd64.tar.gz -C /tmp
sudo mv /tmp/wait4x-linux-amd64/wait4x /usr/local/bin/
```

**macOS:**
```bash
curl -LO https://github.com/wait4x/wait4x/releases/latest/download/wait4x-darwin-amd64.tar.gz
tar -xf wait4x-darwin-amd64.tar.gz -C /tmp
sudo mv /tmp/wait4x-darwin-amd64/wait4x /usr/local/bin/
```

**Windows:**
```bash
curl -LO https://github.com/wait4x/wait4x/releases/latest/download/wait4x-windows-amd64.tar.gz
tar -xf wait4x-windows-amd64.tar.gz
# Move to a directory in your PATH
```

**Verify checksums:**
```bash
curl -LO https://github.com/wait4x/wait4x/releases/latest/download/wait4x-linux-amd64.tar.gz.sha256sum
sha256sum --check wait4x-linux-amd64.tar.gz.sha256sum
```
</details>

<details>
<summary><b>📦 From Package Managers</b></summary>

**Alpine Linux:**
```bash
apk add wait4x
```

**Arch Linux (AUR):**
```bash
yay -S wait4x-bin
```

**NixOS:**
```bash
nix-env -iA nixpkgs.wait4x
```

**Windows (Scoop):**
```bash
scoop install wait4x
```

[![Packaging status](https://repology.org/badge/vertical-allrepos/wait4x.svg?exclude_unsupported=1)](https://repology.org/project/wait4x/versions)
</details>

## 🚀 Quick Start

### Basic TCP Check

Wait for a port to become available:

```bash
wait4x tcp localhost:3306
```

### HTTP Health Check

Wait for a web server with specific response:

```bash
wait4x http https://example.com/health --expect-status-code 200 --expect-body-regex '"status":"UP"'
```

### Multi-Service Check (Parallel)

Wait for multiple services simultaneously:

```bash
wait4x tcp 127.0.0.1:5432 127.0.0.1:6379 127.0.0.1:27017
```

### Database Readiness

Wait for PostgreSQL to be ready:

```bash
wait4x postgresql 'postgres://user:pass@localhost:5432/mydb?sslmode=disable'
```

### Execute After Success

Run a command after services are ready:

```bash
wait4x tcp localhost:8080 -- echo "Service is ready!" && ./start-app.sh
```

### Self-signed certificates
```bash
wait4x http https://www.wait4x.dev --cert-file /path/to/certfile --key-file /path/to/keyfile
```

### CA file
```bash
wait4x http https://www.wait4x.dev --ca-file /path/to/cafile`,
```

## 📖 Detailed Usage

<details>
<summary><b>🌐 HTTP Checking</b></summary>

### Checking with Status Code

Wait for an HTTP endpoint to return a specific status code:

```bash
wait4x http https://api.example.com/health --expect-status-code 200
```

### Checking Response Body with Regex

Wait for an HTTP endpoint to return a response that matches a regex pattern:

```bash
wait4x http https://api.example.com/status --expect-body-regex '"status":\s*"healthy"'
```

### Checking Response Body with JSON Path

Wait for a specific JSON field to exist or have a specific value:

```bash
wait4x http https://api.example.com/status --expect-body-json "services.database.status"
```

This uses [GJSON Path Syntax](https://github.com/tidwall/gjson#path-syntax) for powerful JSON querying.

### Checking Response Body with XPath

Wait for an HTML/XML response to match an XPath query:

```bash
wait4x http https://example.com --expect-body-xpath "//div[@id='status']"
```

### Custom Request Headers

Send specific headers with your HTTP request:

```bash
wait4x http https://api.example.com \
  --request-header "Authorization: Bearer token123" \
  --request-header "Content-Type: application/json"
```

### Checking Response Headers

Wait for a response header to match a pattern:

```bash
wait4x http https://api.example.com --expect-header "Content-Type=application/json"
```
</details>

<details>
<summary><b>🔍 DNS Checking</b></summary>

### Check A Records

```bash
# Basic existence check
wait4x dns A example.com

# With expected IP
wait4x dns A example.com --expected-ip 93.184.216.34

# Using specific nameserver
wait4x dns A example.com --expected-ip 93.184.216.34 -n 8.8.8.8
```

### Check AAAA Records (IPv6)

```bash
wait4x dns AAAA example.com --expected-ip "2606:2800:220:1:248:1893:25c8:1946"
```

### Check CNAME Records

```bash
wait4x dns CNAME www.example.com --expected-domain example.com
```

### Check MX Records

```bash
wait4x dns MX example.com --expected-domain "mail.example.com"
```

### Check NS Records

```bash
wait4x dns NS example.com --expected-nameserver "ns1.example.com"
```

### Check TXT Records

```bash
wait4x dns TXT example.com --expected-value "v=spf1 include:_spf.example.com ~all"
```
</details>

<details>
<summary><b>💾 Database Checking</b></summary>

### MySQL

```bash
# TCP connection
wait4x mysql 'user:password@tcp(localhost:3306)/mydb'

# Unix socket
wait4x mysql 'user:password@unix(/var/run/mysqld/mysqld.sock)/mydb'
```

### PostgreSQL

```bash
# TCP connection
wait4x postgresql 'postgres://user:password@localhost:5432/mydb?sslmode=disable'

# Unix socket
wait4x postgresql 'postgres://user:password@/mydb?host=/var/run/postgresql'
```

### MongoDB

```bash
wait4x mongodb 'mongodb://user:password@localhost:27017/mydb?maxPoolSize=20'
```

### Redis

```bash
# Basic connection
wait4x redis redis://localhost:6379

# With authentication and database selection
wait4x redis redis://user:password@localhost:6379/0

# Check for key existence
wait4x redis redis://localhost:6379 --expect-key "session:active"

# Check for key with specific value (regex)
wait4x redis redis://localhost:6379 --expect-key "status=^ready$"
```

### InfluxDB

```bash
wait4x influxdb http://localhost:8086
```
</details>

<details>
<summary><b>🚌 Message Queue Checking</b></summary>

### RabbitMQ

```bash
wait4x rabbitmq 'amqp://guest:guest@localhost:5672/myvhost'
```

### Temporal

```bash
# Server check
wait4x temporal server localhost:7233

# Worker check (with namespace and task queue)
wait4x temporal worker localhost:7233 \
  --namespace my-namespace \
  --task-queue my-queue

# Check for specific worker identity
wait4x temporal worker localhost:7233 \
  --namespace my-namespace \
  --task-queue my-queue \
  --expect-worker-identity-regex "worker-.*"
```
</details>

<details>
<summary><b>📟 Shell Command</b></summary>

```bash
# Check connection.
wait4x exec 'ping wait4x.dev -c 2'

# Check file existence.
wait4x exec 'ls target/debug/main' --exit-code 2
```

</details>

## ⚙️ Advanced Features

<details>
<summary><b>⏱️ Timeout & Retry Control</b></summary>

### Setting Timeout

Limit the total time Wait4X will wait:

```bash
wait4x tcp localhost:8080 --timeout 30s
```

### Setting Interval

Control how frequently Wait4X retries:

```bash
wait4x tcp localhost:8080 --interval 2s
```

### Exponential Backoff

Use exponential backoff for more efficient retries:

```bash
wait4x http https://api.example.com \
  --backoff-policy exponential \
  --backoff-exponential-coefficient 2.0 \
  --backoff-exponential-max-interval 30s
```
</details>

<details>
<summary><b>↔️ Reverse Checking</b></summary>

Wait for a port to become free:

```bash
wait4x tcp localhost:8080 --invert-check
```

Wait for a service to stop:

```bash
wait4x http https://service.local/health --expect-status-code 200 --invert-check
```
</details>

<details>
<summary><b>⚡ Command Execution</b></summary>

Execute commands after successful wait:

```bash
wait4x tcp localhost:3306 -- ./deploy.sh
```

Chain multiple commands:

```bash
wait4x redis redis://localhost:6379 -- echo "Redis is ready" && ./init-redis.sh
```
</details>

<details>
<summary><b>🔄 Parallel Checking</b></summary>

Wait for multiple services simultaneously:

```bash
wait4x tcp localhost:3306 localhost:6379 localhost:27017
```

Note that this waits for ALL specified services to be ready.
</details>

## 📦 Go Package Usage

<details>
<summary><b>🔌 Installing as a Go Package</b></summary>

Add Wait4X to your Go project:

```bash
go get wait4x.dev/v3
```

Import the packages you need:

```go
import (
    "context"
    "time"

    "wait4x.dev/v3/checker/tcp"      // TCP checker
    "wait4x.dev/v3/checker/http"     // HTTP checker
    "wait4x.dev/v3/checker/redis"    // Redis checker
    "wait4x.dev/v3/waiter"           // Waiter functionality
)
```
</details>

<details>
<summary><b>🌟 Example: TCP Checking</b></summary>

```go
// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Create a TCP checker
tcpChecker := tcp.New("localhost:6379", tcp.WithTimeout(5*time.Second))

// Wait for the TCP port to be available
err := waiter.WaitContext(
    ctx,
    tcpChecker,
    waiter.WithTimeout(time.Minute),
    waiter.WithInterval(2*time.Second),
    waiter.WithBackoffPolicy("exponential"),
)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}

fmt.Println("Service is ready!")
```
</details>

<details>
<summary><b>🌟 Example: HTTP with Advanced Options</b></summary>

```go
// Create HTTP headers
headers := http.Header{}
headers.Add("Authorization", "Bearer token123")
headers.Add("Content-Type", "application/json")

// Create an HTTP checker with validation
checker := http.New(
    "https://api.example.com/health",
    http.WithTimeout(5*time.Second),
    http.WithExpectStatusCode(200),
    http.WithExpectBodyJSON("status"),
    http.WithExpectBodyRegex(`"healthy":\s*true`),
    http.WithExpectHeader("Content-Type=application/json"),
    http.WithRequestHeaders(headers),
)

// Wait for the API to be ready
err := waiter.WaitContext(ctx, checker, options...)
```
</details>

<details>
<summary><b>🌟 Example: Parallel Service Checking</b></summary>

```go
// Create checkers for multiple services
checkers := []checker.Checker{
    redis.New("redis://localhost:6379"),
    postgresql.New("postgres://user:pass@localhost:5432/db"),
    http.New("http://localhost:8080/health"),
}

// Wait for all services in parallel
err := waiter.WaitParallelContext(
    ctx,
    checkers,
    waiter.WithTimeout(time.Minute),
    waiter.WithBackoffPolicy(waiter.BackoffPolicyExponential),
)
```
</details>

<details>
<summary><b>🌟 Example: Custom Checker Implementation</b></summary>

```go
// Define your custom checker
type FileChecker struct {
    filePath string
    minSize  int64
}

// Implement Checker interface
func (f *FileChecker) Identity() (string, error) {
    return fmt.Sprintf("file(%s)", f.filePath), nil
}

func (f *FileChecker) Check(ctx context.Context) error {
    // Check if context is done
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue checking
    }

    fileInfo, err := os.Stat(f.filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return checker.NewExpectedError(
                "file does not exist",
                err,
                "path", f.filePath,
            )
        }
        return err
    }

    if fileInfo.Size() < f.minSize {
        return checker.NewExpectedError(
            "file is smaller than expected",
            nil,
            "path", f.filePath,
            "actual_size", fileInfo.Size(),
            "expected_min_size", f.minSize,
        )
    }

    return nil
}
```
</details>

For more detailed examples with complete code, see the [examples/pkg](examples/pkg) directory. Each example is in its own directory with a runnable `main.go` file.

## 🤝 Contributing

<details>
<summary><b>🐛 Reporting Issues</b></summary>

If you encounter a bug or have a feature request, please open an issue:
- **[Report a bug](https://github.com/wait4x/wait4x/issues/new?template=bug_report.md)**
- **[Request a feature](https://github.com/wait4x/wait4x/issues/new?template=feature_request.md)**

Please include as much information as possible, including:
- Wait4X version
- Command-line arguments
- Expected vs. actual behavior
- Any error messages
</details>

<details>
<summary><b>💻 Code Contributions</b></summary>

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Add tests for your changes
5. Run the tests: `make test`
6. Commit your changes: `git commit -am 'Add awesome feature'`
7. Push the branch: `git push origin feature/your-feature-name`
8. Create a Pull Request
</details>

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

```
Copyright 2019-2025 The Wait4X Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

### Credits

The project logo is based on the "Waiting Man" character (Zhdun) and is used with attribution to the original creator.