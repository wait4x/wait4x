<div align="center">
  <img src="logo.png" alt="Wait4X Logo" width="120">
  <h1>Wait4X</h1>
  <p><b>Wait4X</b> is a lightweight, zero-dependency tool to wait for services to be ready. Perfect for CI/CD, containers, and local development.</p>

  <a href="https://github.com/wait4x/wait4x/actions/workflows/ci.yaml"><img src="https://img.shields.io/github/actions/workflow/status/wait4x/wait4x/ci.yaml?branch=main&style=flat-square" alt="CI Status"></a>
  <a href="https://coveralls.io/github/wait4x/wait4x?branch=main"><img src="https://img.shields.io/coverallsCoverage/github/wait4x/wait4x?branch=main&style=flat-square" alt="Coverage"></a>
  <a href="https://goreportcard.com/report/wait4x.dev/v3"><img src="https://goreportcard.com/badge/wait4x.dev/v3?style=flat-square" alt="Go Report"></a>
  <a href="https://hub.docker.com/r/wait4x/wait4x"><img src="https://img.shields.io/docker/pulls/wait4x/wait4x?logo=docker&style=flat-square" alt="Docker Pulls"></a>
  <a href="https://github.com/wait4x/wait4x/releases"><img src="https://img.shields.io/github/downloads/wait4x/wait4x/total?logo=github&style=flat-square" alt="Downloads"></a>
  <a href="https://repology.org/project/wait4x/versions"><img src="https://img.shields.io/repology/repositories/wait4x?style=flat-square" alt="Packaging"></a>
  <a href="https://pkg.go.dev/wait4x.dev/v3"><img src="https://img.shields.io/badge/reference-007D9C.svg?style=flat-square&logo=go&logoColor=white&labelColor=5C5C5C" alt="Go Reference"></a>
</div>

---

## 📑 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage Examples](#usage-examples)
- [Advanced Features](#advanced-features)
- [Go Package Usage](#go-package-usage)
- [CLI Reference](#cli-reference)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

**Wait4X** helps you wait for services (databases, APIs, message queues, etc.) to be ready before your app or script continues. It's ideal for:

- **CI/CD pipelines**: Ensure dependencies are up before tests run
- **Containers & orchestration**: Health check services before startup
- **Deployments**: Verify readiness before rollout
- **Local development**: Simplify service readiness checks

## Features

| Feature | Description |
|---------|-------------|
| **Multi-Protocol** | TCP, HTTP, DNS, and more |
| **Service Integrations** | Redis, MySQL, PostgreSQL, MongoDB, Kafka, RabbitMQ, InfluxDB, Temporal |
| **Reverse/Parallel Checking** | Invert checks or check multiple services at once |
| **Exponential Backoff** | Smarter retries |
| **Cross-Platform** | Single binary for Linux, macOS, Windows |
| **Go Package** | Use as a Go library |
| **Command Execution** | Run commands after checks |

## 📥 Installation

*After installing, jump to [Quick Start](#quick-start) to try it out!*

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

<details>
<summary><b>🐹 Go Install (for Go users)</b></summary>

You can install Wait4X directly from source using Go (requires Go 1.16+):

```bash
go install wait4x.dev/v3/cmd/wait4x@latest
```

This will place the `wait4x` binary in your `$GOPATH/bin` or `$HOME/go/bin` directory.
</details>

## 🚀 Quick Start

Get started in seconds! After [installing](#installation), try these common checks:

### Wait for a TCP Port
```bash
wait4x tcp localhost:3306
```

### HTTP Health Check
```bash
wait4x http https://example.com/health --expect-status-code 200
```

### Wait for Multiple Services (Parallel)
```bash
wait4x tcp 127.0.0.1:5432 127.0.0.1:6379 127.0.0.1:27017
```

### Database Readiness
```bash
wait4x postgresql 'postgres://user:pass@localhost:5432/mydb?sslmode=disable'
```

For more, see [Usage Examples](#usage-examples) or [Detailed Usage](#detailed-usage).

## Usage Examples

Here are some of the most useful Wait4X commands. Click the links for more details!

- **TCP:** Wait for a port to be available  
  [`wait4x tcp localhost:8080`](#main-commands)
- **HTTP:** Wait for a web endpoint with status code and body check  
  [`wait4x http https://api.example.com/health --expect-status-code 200 --expect-body-regex '"status":"UP"'`](#main-commands)
- **DNS:** Wait for DNS A record  
  [`wait4x dns A example.com`](#main-commands)
- **MySQL:** Wait for MySQL DB  
  [`wait4x mysql 'user:password@tcp(localhost:3306)/mydb'`](#main-commands)
- **Redis:** Wait for Redis and check for a key  
  [`wait4x redis redis://localhost:6379 --expect-key "session:active"`](#main-commands)
- **Run a command after check:**  
  [`wait4x tcp localhost:8080 -- ./start-app.sh`](#main-commands)
- **Reverse check (wait for port to be free):**  
  [`wait4x tcp localhost:8080 --invert-check`](#main-commands)
- **Parallel check:**  
  [`wait4x tcp localhost:3306 localhost:6379 localhost:27017`](#main-commands)

See [Detailed Usage](#detailed-usage) for advanced options and more protocols.

## 📖 Detailed Usage

Jump to:
- [HTTP Checking](#http-checking)
- [DNS Checking](#dns-checking)
- [Database Checking](#database-checking)
- [Message Queue Checking](#message-queue-checking)
- [Shell Command](#shell-command)

---

### HTTP Checking

Wait for an HTTP(S) endpoint to be ready, with flexible validation options.

- **Status code check:**
  ```bash
  wait4x http https://api.example.com/health --expect-status-code 200
  ```
- **Response body regex:**
  ```bash
  wait4x http https://api.example.com/status --expect-body-regex '"status":\s*"healthy"'
  ```
- **JSON path check:**
  ```bash
  wait4x http https://api.example.com/status --expect-body-json "services.database.status"
  ```
  Uses [GJSON Path Syntax](https://github.com/tidwall/gjson#path-syntax).
- **XPath check:**
  ```bash
  wait4x http https://example.com --expect-body-xpath "//div[@id='status']"
  ```
- **Custom request headers:**
  ```bash
  wait4x http https://api.example.com \
    --request-header "Authorization: Bearer token123" \
    --request-header "Content-Type: application/json"
  ```
- **Response header check:**
  ```bash
  wait4x http https://api.example.com --expect-header "Content-Type=application/json"
  ```
- **TLS options:**
  ```bash
  wait4x http https://www.wait4x.dev --cert-file /path/to/certfile --key-file /path/to/keyfile
  wait4x http https://www.wait4x.dev --ca-file /path/to/cafile
  ```

---

### DNS Checking

Check for various DNS record types and values.

- **A record:**
  ```bash
  wait4x dns A example.com
  wait4x dns A example.com --expected-ip 93.184.216.34
  wait4x dns A example.com --expected-ip 93.184.216.34 -n 8.8.8.8
  ```
- **AAAA record (IPv6):**
  ```bash
  wait4x dns AAAA example.com --expected-ip "2606:2800:220:1:248:1893:25c8:1946"
  ```
- **CNAME record:**
  ```bash
  wait4x dns CNAME www.example.com --expected-domain example.com
  ```
- **MX record:**
  ```bash
  wait4x dns MX example.com --expected-domain "mail.example.com"
  ```
- **NS record:**
  ```bash
  wait4x dns NS example.com --expected-nameserver "ns1.example.com"
  ```
- **TXT record:**
  ```bash
  wait4x dns TXT example.com --expected-value "v=spf1 include:_spf.example.com ~all"
  ```

---

### Database Checking

Check readiness for popular databases.

#### MySQL
- **TCP connection:**
  ```bash
  wait4x mysql 'user:password@tcp(localhost:3306)/mydb'
  ```
- **Unix socket:**
  ```bash
  wait4x mysql 'user:password@unix(/var/run/mysqld/mysqld.sock)/mydb'
  ```

#### PostgreSQL
- **TCP connection:**
  ```bash
  wait4x postgresql 'postgres://user:password@localhost:5432/mydb?sslmode=disable'
  ```
- **Unix socket:**
  ```bash
  wait4x postgresql 'postgres://user:password@/mydb?host=/var/run/postgresql'
  ```

#### MongoDB
  ```bash
  wait4x mongodb 'mongodb://user:password@localhost:27017/mydb?maxPoolSize=20'
  ```

#### Redis
- **Basic connection:**
  ```bash
  wait4x redis redis://localhost:6379
  ```
- **With authentication and DB selection:**
  ```bash
  wait4x redis redis://user:password@localhost:6379/0
  ```
- **Check for key existence:**
  ```bash
  wait4x redis redis://localhost:6379 --expect-key "session:active"
  ```
- **Check for key with value (regex):**
  ```bash
  wait4x redis redis://localhost:6379 --expect-key "status=^ready$"
  ```

#### InfluxDB
  ```bash
  wait4x influxdb http://localhost:8086
  ```

---

### Message Queue Checking

#### RabbitMQ
  ```bash
  wait4x rabbitmq 'amqp://guest:guest@localhost:5672/myvhost'
  ```

#### Temporal
- **Server check:**
  ```bash
  wait4x temporal server localhost:7233
  ```
- **Worker check (namespace & task queue):**
  ```bash
  wait4x temporal worker localhost:7233 \
    --namespace my-namespace \
    --task-queue my-queue
  ```
- **Check for specific worker identity:**
  ```bash
  wait4x temporal worker localhost:7233 \
    --namespace my-namespace \
    --task-queue my-queue \
    --expect-worker-identity-regex "worker-.*"
  ```

#### Kafka
- **Basic Kafka broker readiness check:**
  ```bash
  wait4x kafka kafka://localhost:9092
  ```

- **Check Kafka broker with SCRAM authentication:**
  ```bash
  wait4x kafka kafka://user:pass@localhost:9092?authMechanism=scram-sha-256
  ```

- **Wait for multiple Kafka brokers (cluster) to be ready:**
  ```bash
  wait4x kafka kafka://broker1:9092 kafka://broker2:9092 kafka://broker3:9092
  ```

> **Notes:**
> - The connection string format is: kafka://[user:pass@]host:port[?option=value&...]
> - Supported options: authMechanism (scram-sha-256, scram-sha-512)
---

### Shell Command

Wait for a shell command to succeed or return a specific exit code.

- **Check connection:**
  ```bash
  wait4x exec 'ping wait4x.dev -c 2'
  ```
- **Check file existence:**
  ```bash
  wait4x exec 'ls target/debug/main' --exit-code 2
  ```

---

See [Advanced Features](#advanced-features) for timeout, retry, backoff, and parallel/reverse checking options.

## ⚙️ Advanced Features

Jump to:
- [Timeout & Retry Control](#timeout--retry-control)
- [Exponential Backoff](#exponential-backoff)
- [Reverse Checking](#reverse-checking)
- [Command Execution](#command-execution)
- [Parallel Checking](#parallel-checking)

---

### Timeout & Retry Control

Control how long Wait4X waits and how often it checks.

- **Set a timeout:**
  ```bash
  wait4x tcp localhost:8080 --timeout 30s
  ```
  *Waits up to 30 seconds before giving up.*

- **Set check interval:**
  ```bash
  wait4x tcp localhost:8080 --interval 2s
  ```
  *Checks every 2 seconds (default: 1s).* 

---

### Exponential Backoff

Retry with increasing delays for more efficient waiting (useful for slow-starting services).

- **Enable exponential backoff:**
  ```bash
  wait4x http https://api.example.com \
    --backoff-policy exponential \
    --backoff-exponential-coefficient 2.0 \
    --backoff-exponential-max-interval 30s
  ```
  *Doubles the wait time between retries, up to 30 seconds.*

---

### Reverse Checking

Wait for a service to become unavailable (e.g., port to be free, service to stop).

- **Wait for a port to become free:**
  ```bash
  wait4x tcp localhost:8080 --invert-check
  ```
- **Wait for a service to stop:**
  ```bash
  wait4x http https://service.local/health --expect-status-code 200 --invert-check
  ```
  *Use for shutdown/cleanup workflows or to ensure a port is not in use.*

---

### Command Execution

Run a command after a successful check (great for CI/CD or startup scripts).

- **Run a script after waiting:**
  ```bash
  wait4x tcp localhost:3306 -- ./deploy.sh
  ```
- **Chain multiple commands:**
  ```bash
  wait4x redis redis://localhost:6379 -- echo "Redis is ready" && ./init-redis.sh
  ```
  *Automate your workflow after dependencies are ready.*

---

### Parallel Checking

Wait for multiple services at once (all must be ready to continue).

- **Check several services in parallel:**
  ```bash
  wait4x tcp localhost:3306 localhost:6379 localhost:27017
  ```
  *Use for microservices, integration tests, or complex startup dependencies.*

---

See [CLI Reference](#cli-reference) for all available flags and options.

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

## 📝 CLI Reference

Wait4X provides a flexible CLI with many commands and options. Here is a summary of the main commands and global flags. For the most up-to-date and detailed information, use the built-in help:

```bash
wait4x --help
wait4x <command> --help
```

### Main Commands

| Command        | Description                                       |
|----------------|---------------------------------------------------|
| `tcp`          | Wait for a TCP port to become available           |
| `http`         | Wait for an HTTP(S) endpoint with advanced checks |
| `dns`          | Wait for DNS records (A, AAAA, CNAME, MX, etc.)   |
| `kafka`        | Wait for Kafka server                             |
| `mysql`        | Wait for a MySQL database to be ready             |
| `postgresql`   | Wait for a PostgreSQL database to be ready        |
| `mongodb`      | Wait for a MongoDB database to be ready           |
| `redis`        | Wait for a Redis server or key                    |
| `influxdb`     | Wait for an InfluxDB server                       |
| `rabbitmq`     | Wait for a RabbitMQ server                        |
| `temporal`     | Wait for a Temporal server or worker              |
| `exec`         | Wait for a shell command to succeed               |

Each command supports its own set of flags. See examples above or run `wait4x <command> --help` for details.

### Global Flags

| Flag                                 | Description                                      |
|--------------------------------------|--------------------------------------------------|
| `--timeout`, `-t`                    | Set the maximum wait time (e.g., `30s`, `2m`)     |
| `--interval`, `-i`                   | Set the interval between checks (default: 1s)     |
| `--invert-check`                     | Invert the check (wait for NOT ready)             |
| `--backoff-policy`                   | Retry policy: `linear` or `exponential`           |
| `--backoff-exponential-coefficient`  | Exponential backoff multiplier (default: 2.0)     |
| `--backoff-exponential-max-interval` | Max interval for exponential backoff              |
| `--quiet`                            | Suppress output except errors                     |
| `--no-color`                         | Disable colored output                            |

### Getting Help

For a full list of commands and options, run:

```bash
wait4x --help
wait4x <command> --help
```

## 🤝 Contributing

We welcome contributions of all kinds! Whether you want to fix a bug, add a feature, improve documentation, or help others, you're in the right place.

**How to contribute:**
1. [Fork the repository](https://github.com/wait4x/wait4x/fork)
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes (add tests if possible)
4. Run tests: `make test`
5. Commit: `git commit -am 'Describe your change'`
6. Push: `git push origin feature/your-feature-name`
7. [Open a Pull Request](https://github.com/wait4x/wait4x/pulls)

**Found a bug or have a feature request?**
- [Report an issue](https://github.com/wait4x/wait4x/issues/new/choose)

For more details, see [CONTRIBUTING.md](CONTRIBUTING.md) (if available).

## 💬 Community & Support

- 💡 **Questions or ideas?** Use [GitHub Discussions](https://github.com/wait4x/wait4x/discussions)
- 🐞 **Bugs or feature requests?** [Open an issue](https://github.com/wait4x/wait4x/issues/new/choose)
- ⭐ **Star the repo** to support the project!

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