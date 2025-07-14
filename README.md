# Go-Insight Go SDK Documentation

Welcome to the comprehensive documentation for the Go-Insight Go SDK.

## Table of Contents

- [Quick Start Guide](quick-start.md)
- [API Reference](api-reference.md)
- [Framework Integration](framework-integration.md)
- [Advanced Usage](advanced-usage.md)
- [Best Practices](best-practices.md)
- [Performance Guide](performance.md)
- [Troubleshooting](troubleshooting.md)
- [Migration Guide](migration.md)

## Overview

The Go-Insight Go SDK provides zero-configuration observability for Go applications. With just one line of middleware, you get:

- **Automatic request tracing** with distributed trace correlation
- **Performance metrics collection** for all HTTP endpoints
- **Structured logging** with trace context
- **Custom instrumentation** for business logic
- **Framework integration** for Gin and Echo

## Quick Examples

### Gin Framework
```go
import "github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"

client := goinsight.New(goinsight.Config{
    APIKey:      "your-api-key",
    Endpoint:    "http://localhost:8080",
    ServiceName: "my-service",
})

r := gin.Default()
r.Use(client.GinMiddleware()) // One line for full observability!
```

### Manual Instrumentation
```go
// Wrap any function with automatic instrumentation
processData := client.Instrument("process_data", func(ctx context.Context) error {
    // Your business logic here
    client.LogInfo(ctx, "Processing completed")
    return nil
})

err := processData(ctx)
```

## Key Features

### 🚀 Zero Configuration
Add one line of middleware and get complete observability without any additional setup.

### 🔗 Automatic Correlation
All logs, metrics, and traces are automatically correlated using distributed tracing context.

### ⚡ High Performance
- Sub-millisecond overhead per request
- Asynchronous operations that don't block your application
- Efficient connection pooling and batching

### 🛡️ Production Ready
- Graceful error handling that never breaks your app
- Configurable timeouts and retry logic
- Comprehensive test coverage

### 🎯 Framework Support
Built-in middleware for popular Go web frameworks:
- Gin (v1.9.1+)
- Echo (v4.11.3+)
- Custom HTTP handlers

## Getting Started

1. **Install the SDK**
   ```bash
   go get github.com/NathanSanchezDev/go-insight-go-sdk
   ```

2. **Initialize the client**
   ```go
   client := goinsight.New(goinsight.Config{
       APIKey:      "your-api-key",
       Endpoint:    "http://localhost:8080",
       ServiceName: "my-service",
   })
   ```

3. **Add middleware** (for web frameworks)
   ```go
   r.Use(client.GinMiddleware())
   ```

4. **Start observing** your application automatically!

## Documentation Structure

### For Beginners
- Start with the [Quick Start Guide](quick-start.md)
- Check out [Framework Integration](framework-integration.md) for your web framework

### For Advanced Users
- [Advanced Usage](advanced-usage.md) for custom instrumentation
- [Performance Guide](performance.md) for optimization tips
- [Best Practices](best-practices.md) for production deployments

### For Contributors
- [API Reference](api-reference.md) for complete method documentation
- [Troubleshooting](troubleshooting.md) for common issues

## Support

- 🐛 [Issue Tracker](https://github.com/NathanSanchezDev/go-insight-go-sdk/issues)
- 💬 [Discussions](https://github.com/NathanSanchezDev/go-insight-go-sdk/discussions)
- 📖 [Main Go-Insight Documentation](https://github.com/NathanSanchezDev/go-insight)

## Contributing

We welcome contributions! Please see our [Contributing Guide](../CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.