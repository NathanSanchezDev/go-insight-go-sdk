[![Go Reference](https://pkg.go.dev/badge/github.com/NathanSanchezDev/go-insight-go-sdk.svg)](https://pkg.go.dev/github.com/NathanSanchezDev/go-insight-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/NathanSanchezDev/go-insight-go-sdk)](https://goreportcard.com/report/github.com/NathanSanchezDev/go-insight-go-sdk)

<p align="center">
  <img src="https://github.com/user-attachments/assets/c193fadf-1508-4fde-881e-b8753380cbc7" alt="Go-Insight Logo" width="400" height="400">
</p>

Official Go SDK for [Go-Insight](https://github.com/NathanSanchezDev/go-insight) observability platform.

## Features

- 🚀 **Zero Configuration** - Add one line, get full observability
- 🔗 **Automatic Correlation** - Traces, spans, and logs automatically linked  
- 📊 **Smart Metrics** - HTTP performance automatically tracked
- 🛡️ **Error Handling** - Graceful failures, non-blocking operations
- ⚡ **High Performance** - Async operations, minimal overhead
- 🎯 **Framework Support** - Gin, Echo middleware included

## Installation

```bash
go get github.com/NathanSanchezDev/go-insight-go-sdk
```

## Quick Start

### Basic Setup

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"
)

func main() {
    // Initialize Go-Insight client
    client := goinsight.New(goinsight.Config{
        APIKey:      "your-api-key",
        Endpoint:    "http://localhost:8080",
        ServiceName: "my-go-service",
    })

    // Setup Gin with auto-instrumentation
    r := gin.Default()
    r.Use(client.GinMiddleware())

    // Your routes are automatically instrumented!
    r.GET("/users", func(c *gin.Context) {
        // Manual logging with automatic trace correlation
        client.LogInfo(c.Request.Context(), "Fetching users", map[string]interface{}{
            "user_count": 42,
        })

        c.JSON(200, gin.H{"users": []string{}})
    })

    r.Run(":8080")
}
```

## Framework Support

### Gin Framework

```go
r := gin.Default()
r.Use(client.GinMiddleware())
```

### Echo Framework

```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.Use(client.EchoMiddleware())
```

## Manual Instrumentation

### Custom Spans

```go
func processUser(ctx context.Context, client *goinsight.Client, userID string) error {
    // Start a span for this operation
    spanCtx, err := client.StartSpan(ctx, "process_user")
    if err != nil {
        return err
    }
    defer client.FinishSpan(spanCtx)

    // Your business logic here
    client.LogInfo(spanCtx, "Processing user", map[string]interface{}{
        "user_id": userID,
    })

    return nil
}
```

### Function Decoration

```go
// Wrap function with automatic instrumentation
instrumentedFn := client.Instrument("database_query", func(ctx context.Context) error {
    // Your database logic here
    time.Sleep(50 * time.Millisecond)
    return nil
})

// Call instrumented function
err := instrumentedFn(ctx)
```

## Configuration

### Environment Variables

```bash
export GO_INSIGHT_API_KEY="your-api-key"
export GO_INSIGHT_ENDPOINT="http://localhost:8080"
export GO_INSIGHT_SERVICE_NAME="my-service"
```

### Programmatic Configuration

```go
config := goinsight.Config{
    APIKey:      "your-api-key",           // Required
    Endpoint:    "http://localhost:8080",  // Required
    ServiceName: "my-service",             // Required
    Timeout:     10 * time.Second,         // Optional, default: 5s
}

client := goinsight.New(config)
```

## Examples

Check out the [examples directory](examples/) for complete working examples:

- [Gin Example](examples/gin-example/) - Complete Gin web server with auto-instrumentation
- [Echo Example](examples/echo-example/) - Echo framework integration
- [Manual Instrumentation](examples/manual-instrumentation/) - Custom tracing and logging

## Documentation

📖 **[Complete Documentation](docs/)**

### Quick Links
- [📚 Quick Start Guide](docs/quick-start.md) - Get up and running in 5 minutes
- [🔧 API Reference](docs/api-reference.md) - Complete method documentation
- [🎯 Framework Integration](docs/framework-integration.md) - Gin & Echo setup guides
- [⚡ Advanced Usage](docs/advanced-usage.md) - Complex patterns & techniques
- [🏆 Best Practices](docs/best-practices.md) - Production deployment guidelines
- [🚀 Performance Guide](docs/performance.md) - Optimization & benchmarking
- [🔍 Troubleshooting](docs/troubleshooting.md) - Common issues & solutions
- [📋 Migration Guide](docs/migration.md) - Migrating from other tools

## API Reference

### Logging

```go
// Different log levels
client.LogInfo(ctx, "Info message", metadata)
client.LogWarn(ctx, "Warning message", metadata)
client.LogError(ctx, "Error message", err, metadata)
client.LogDebug(ctx, "Debug message", metadata)

// Generic log method
client.Log(ctx, "INFO", "Custom message", metadata)
```

### Metrics

```go
metric := goinsight.Metric{
    Path:        "/api/users",
    Method:      "GET",
    StatusCode:  200,
    Duration:    45.7,
    Source: goinsight.MetricSource{
        Language:  "go",
        Framework: "gin",
        Version:   "1.9.1",
    },
    Metadata: map[string]interface{}{
        "cache_hit": true,
    },
}

client.SendMetric(metric)
```

### Distributed Tracing

```go
// Start a new trace
ctx, traceCtx, err := client.StartTrace(context.Background(), "user_request")
if err == nil {
    defer client.FinishTrace(ctx)
    
    // Start spans within the trace
    spanCtx, err := client.StartSpan(ctx, "database_query")
    if err == nil {
        defer client.FinishSpan(spanCtx)
        
        // Your instrumented code here
    }
}
```

## Performance

- **< 1ms overhead** per instrumented request
- **Non-blocking**: All operations are asynchronous and won't block your application
- **Minimal Overhead**: < 1ms overhead per instrumented request
- **Connection Pooling**: Efficient HTTP client with connection reuse
- **Graceful Failures**: SDK errors never affect your application logic

## Best Practices

1. **Service Naming**: Use consistent service names across environments
2. **Error Handling**: Always handle SDK errors gracefully - they should never break your app
3. **Async Operations**: SDK operations are async by default for performance
4. **Trace Context**: Pass context between functions to maintain trace correlation
5. **Metadata**: Include relevant metadata for better observability

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](docs/) - Comprehensive guides and references
- 🐛 [Issue Tracker](https://github.com/NathanSanchezDev/go-insight-go-sdk/issues) - Report bugs or request features
- 💬 [Discussions](https://github.com/NathanSanchezDev/go-insight-go-sdk/discussions) - Community support and questions
- 🏠 [Main Go-Insight Project](https://github.com/NathanSanchezDev/go-insight) - The observability platform
