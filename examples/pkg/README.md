# Wait4X as an Importable Package

These examples demonstrate how to use Wait4X as an importable package in your Go applications. Wait4X isn't just a CLI tool - it provides a powerful library that you can integrate directly into your Go code.

## Examples Overview

1. **Basic TCP Checker** (`tcp_basic/main.go`): Simple example of waiting for a TCP port to become available.

2. **Advanced HTTP Checker** (`http_advanced/main.go`): Demonstrates complex HTTP checking with custom headers, body validations, status code checks, and more.

3. **gRPC Health Check** (`grpc_health/main.go`): Shows how to check gRPC server health using the standard health checking protocol.

4. **Parallel Service Checking** (`parallel_services/main.go`): Shows how to check multiple services in parallel, waiting for all of them to be ready before proceeding.

5. **Reverse Checking** (`reverse_checking/main.go`): Example of using the inverse check to wait for a port to become free.

6. **Custom Checker** (`custom_checker/main.go`): Shows how to create your own custom checker by implementing the Checker interface.

## Using Wait4X in Your Go Projects

To use Wait4X in your Go project, add it as a dependency:

```bash
go get wait4x.dev/v4
```

Then import the packages you need:

```go
import (
    "wait4x.dev/v4/checker/tcp"      // TCP checker
    "wait4x.dev/v4/checker/http"     // HTTP checker
    "wait4x.dev/v4/checker/grpc"     // gRPC health checker
    "wait4x.dev/v4/checker/redis"    // Redis checker
    // ...other checkers
    "wait4x.dev/v4/waiter"           // Waiter functionality
)
```

### Core Components

1. **Checkers**: Implements the `checker.Checker` interface:
   ```go
   type Checker interface {
       Identity() (string, error)
       Check(ctx context.Context) error
   }
   ```

2. **Waiter**: Provides waiting functionality with options like timeout, interval, backoff, etc.

3. **Context Usage**: All checkers and waiters support context for cancellation and timeouts.

### Common Patterns

1. **Option Pattern**: All checkers use the functional options pattern for configuration.

2. **Error Handling**: Use the `ExpectedError` type for expected failures vs. unexpected errors.

3. **Parallel Execution**: Use `WaitParallelContext` to check multiple services simultaneously.

4. **Context Propagation**: Always pass context to allow for proper cancellation and timeouts.

## Extending Wait4X

To create your own checker:

1. Define a type that implements the `checker.Checker` interface
2. Implement the `Identity()` and `Check()` methods
3. Use the `checker.NewExpectedError()` function for creating appropriate error types

See `custom_checker.go` for a complete example of implementing a custom checker.

## Best Practices

1. Always use contexts with timeouts to prevent indefinite waiting
2. Consider using exponential backoff for services that might take a while to start
3. Use parallel checking when waiting for multiple independent services
4. Handle errors appropriately - distinguish between timeout errors and other errors
5. Add logging where appropriate to understand what's happening during waiting

## Additional Resources

- Go Reference Documentation: https://pkg.go.dev/wait4x.dev/v4
- GitHub Repository: https://github.com/wait4x/wait4x
- Report Issues: https://github.com/wait4x/wait4x/issues