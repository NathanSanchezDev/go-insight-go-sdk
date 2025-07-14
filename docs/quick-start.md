# Quick Start Guide

Get up and running with Go-Insight Go SDK in under 5 minutes.

## Prerequisites

- Go 1.21 or later
- A running Go-Insight instance
- API key for authentication

## Installation

```bash
go get github.com/NathanSanchezDev/go-insight-go-sdk
```

## Basic Setup

### 1. Import the SDK

```go
import "github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"
```

### 2. Initialize the Client

```go
client := goinsight.New(goinsight.Config{
    APIKey:      "your-api-key",
    Endpoint:    "http://localhost:8080", // Your Go-Insight endpoint
    ServiceName: "my-awesome-service",
    Timeout:     5 * time.Second, // Optional, defaults to 5s
})
```

### 3. Add Framework Middleware

#### For Gin Framework

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"
)

func main() {
    client := goinsight.New(goinsight.Config{
        APIKey:      "your-api-key",
        Endpoint:    "http://localhost:8080",
        ServiceName: "gin-service",
    })

    r := gin.Default()
    
    // Add Go-Insight middleware - this one line gives you full observability!
    r.Use(client.GinMiddleware())

    r.GET("/ping", func(c *gin.Context) {
        // This request is automatically traced, logged, and metrics are collected
        c.JSON(200, gin.H{"message": "pong"})
    })

    r.Run(":8080")
}
```

#### For Echo Framework

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"
)

func main() {
    client := goinsight.New(goinsight.Config{
        APIKey:      "your-api-key",
        Endpoint:    "http://localhost:8080",
        ServiceName: "echo-service",
    })

    e := echo.New()
    
    // Add Go-Insight middleware
    e.Use(client.EchoMiddleware())

    e.GET("/ping", func(c echo.Context) error {
        return c.JSON(200, map[string]string{"message": "pong"})
    })

    e.Logger.Fatal(e.Start(":8080"))
}
```

### 4. Manual Logging (Optional)

Add custom logs with automatic trace correlation:

```go
r.GET("/users", func(c *gin.Context) {
    // Automatic trace correlation - no need to pass trace IDs manually!
    client.LogInfo(c.Request.Context(), "Fetching users", map[string]interface{}{
        "page": 1,
        "limit": 10,
    })

    // Your business logic here
    users := fetchUsers()

    client.LogInfo(c.Request.Context(), "Users fetched successfully", map[string]interface{}{
        "count": len(users),
    })

    c.JSON(200, users)
})
```

## What You Get Automatically

Once you add the middleware, you automatically get:

### 🔍 Distributed Tracing
- Every HTTP request creates a trace
- Nested spans for complex operations
- Automatic trace correlation across services

### 📊 Performance Metrics
- Request duration
- HTTP status codes
- Endpoint performance
- Framework and language metadata

### 📝 Request Logging
- Request start/completion logs
- Automatic log levels based on status codes
- Request metadata (method, path, user agent, etc.)

### 🔗 Trace Correlation
- All logs include trace and span IDs
- Easy correlation between logs, metrics, and traces
- Context propagation across function calls

## Testing Your Setup

1. **Start your application** with the middleware enabled

2. **Make some requests** to your endpoints:
   ```bash
   curl http://localhost:8080/ping
   curl http://localhost:8080/users
   ```

3. **Check Go-Insight** for your data:
   ```bash
   # View logs
   curl -H "X-API-Key: your-api-key" \
     "http://localhost:8080/logs?service=my-awesome-service"

   # View metrics
   curl -H "X-API-Key: your-api-key" \
     "http://localhost:8080/metrics?service=my-awesome-service"

   # View traces
   curl -H "X-API-Key: your-api-key" \
     "http://localhost:8080/traces?service=my-awesome-service"
   ```

## Configuration Options

### Environment Variables

You can configure the SDK using environment variables:

```bash
export GO_INSIGHT_API_KEY="your-api-key"
export GO_INSIGHT_ENDPOINT="http://localhost:8080"
export GO_INSIGHT_SERVICE_NAME="my-service"
export GO_INSIGHT_TIMEOUT="10s"
```

Then initialize with minimal config:

```go
client := goinsight.New(goinsight.Config{
    APIKey:      os.Getenv("GO_INSIGHT_API_KEY"),
    Endpoint:    os.Getenv("GO_INSIGHT_ENDPOINT"),
    ServiceName: os.Getenv("GO_INSIGHT_SERVICE_NAME"),
})
```

### Configuration Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `APIKey` | string | Yes | - | Authentication key for Go-Insight |
| `Endpoint` | string | Yes | - | Go-Insight server URL |
| `ServiceName` | string | Yes | - | Name of your service |
| `Timeout` | time.Duration | No | 5s | HTTP request timeout |

## Common Patterns

### Adding Custom Context

```go
r.GET("/process", func(c *gin.Context) {
    // Start a custom span for a specific operation
    spanCtx, err := client.StartSpan(c.Request.Context(), "data_processing")
    if err == nil {
        defer client.FinishSpan(spanCtx)
        
        // Use spanCtx for operations within this span
        client.LogInfo(spanCtx, "Starting data processing")
        
        // Your processing logic here
        processData()
        
        client.LogInfo(spanCtx, "Data processing completed")
    }
    
    c.JSON(200, gin.H{"status": "processed"})
})
```

### Error Handling

```go
r.GET("/risky", func(c *gin.Context) {
    err := riskyOperation()
    if err != nil {
        // Log the error with automatic trace correlation
        client.LogError(c.Request.Context(), "Risky operation failed", err, map[string]interface{}{
            "operation": "risky_operation",
            "user_id": c.GetHeader("User-ID"),
        })
        
        c.JSON(500, gin.H{"error": "Operation failed"})
        return
    }
    
    c.JSON(200, gin.H{"status": "success"})
})
```

### Function Instrumentation

```go
func processUserData(ctx context.Context, client *goinsight.Client, userID string) error {
    // Wrap function with automatic instrumentation
    instrumentedFn := client.Instrument("process_user_data", func(ctx context.Context) error {
        // Your business logic here
        time.Sleep(100 * time.Millisecond) // Simulate work
        
        client.LogInfo(ctx, "User data processed", map[string]interface{}{
            "user_id": userID,
        })
        
        return nil
    })
    
    return instrumentedFn(ctx)
}
```

## Next Steps

### 🎯 Learn More
- [Framework Integration](framework-integration.md) - Deep dive into Gin and Echo integration
- [Advanced Usage](advanced-usage.md) - Custom spans, traces, and instrumentation
- [Best Practices](best-practices.md) - Production deployment guidelines

### 🔧 Customize
- Add custom metadata to logs and metrics
- Create manual spans for specific operations
- Implement custom instrumentation patterns

### 📊 Monitor
- Set up dashboards in Go-Insight
- Configure alerts based on error rates
- Monitor performance trends over time

## Troubleshooting

### Common Issues

**SDK not sending data:**
- Verify API key is correct
- Check Go-Insight endpoint is reachable
- Ensure network connectivity

**High latency:**
- Check if operations are being sent synchronously
- Verify Go-Insight server performance
- Consider adjusting timeout settings

**Missing traces:**
- Ensure middleware is properly registered
- Check for context propagation issues
- Verify trace context is being passed correctly

See [Troubleshooting Guide](troubleshooting.md) for detailed solutions.

## Performance Impact

The Go-Insight SDK is designed for production use with minimal performance impact:

- **< 1ms overhead** per instrumented request
- **Asynchronous operations** don't block your application
- **Efficient batching** and connection pooling
- **Graceful degradation** if Go-Insight is unavailable

## Support

Need help? Here are your options:

- 📖 [Documentation](README.md) - Comprehensive guides and references
- 🐛 [Issue Tracker](https://github.com/NathanSanchezDev/go-insight-go-sdk/issues) - Report bugs or request features
- 💬 [Discussions](https://github.com/NathanSanchezDev/go-insight-go-sdk/discussions) - Community support and questions

---

**You're ready to go!** 🚀 Your application now has comprehensive observability with just a few lines of code.