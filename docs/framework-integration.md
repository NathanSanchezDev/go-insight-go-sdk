# Framework Integration

Detailed guide for integrating Go-Insight SDK with popular Go web frameworks.

## Gin Framework

### Basic Setup

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
    r.Use(client.GinMiddleware()) // Add Go-Insight instrumentation

    r.GET("/users", handleUsers)
    r.Run(":8080")
}
```

### What Gin Middleware Provides

- **Automatic request tracing** for every HTTP request
- **Performance metrics** (duration, status code, method, path)
- **Request logging** with configurable log levels based on status codes
- **Trace context injection** into Gin context for manual instrumentation

### Accessing Go-Insight in Handlers

```go
func handleUsers(c *gin.Context) {
    // Access the Go-Insight client from middleware
    if traceCtx, exists := c.Get("go-insight-trace"); exists {
        // Trace context is available
        client.LogInfo(c.Request.Context(), "Processing users request")
    }
    
    // Manual span creation
    spanCtx, err := client.StartSpan(c.Request.Context(), "fetch_users_from_db")
    if err == nil {
        defer client.FinishSpan(spanCtx)
        // Database operations here
    }
    
    c.JSON(200, gin.H{"users": []string{}})
}
```

### Gin Configuration Options

The Gin middleware automatically handles:

- **Route path extraction** using `ginCtx.FullPath()`
- **Status code detection** from `ginCtx.Writer.Status()`
- **Request metadata** including User-Agent, Request-ID headers
- **Async operations** to avoid blocking request processing

## Echo Framework

### Basic Setup

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
    e.Use(client.EchoMiddleware()) // Add Go-Insight instrumentation

    e.GET("/users", handleUsers)
    e.Logger.Fatal(e.Start(":8080"))
}
```

### What Echo Middleware Provides

- **Automatic request tracing** with Echo-specific context handling
- **Route-aware metrics** using Echo's route patterns
- **Error handling** that respects Echo's error handling patterns
- **Context propagation** through Echo's request context

### Accessing Go-Insight in Handlers

```go
func handleUsers(c echo.Context) error {
    // Access trace context
    if traceCtx, ok := c.Get("go-insight-trace").(*goinsight.TraceContext); ok {
        client.LogInfo(c.Request().Context(), "Processing Echo request")
    }
    
    // Manual instrumentation
    spanCtx, err := client.StartSpan(c.Request().Context(), "business_logic")
    if err == nil {
        defer client.FinishSpan(spanCtx)
        // Your business logic here
    }
    
    return c.JSON(200, map[string]interface{}{"users": []string{}})
}
```

### Echo Configuration Options

The Echo middleware handles:

- **Route path extraction** using `echoCtx.Path()`
- **Status code detection** from `echoCtx.Response().Status`
- **Error propagation** without interfering with Echo's error handling
- **Request ID extraction** from headers

## Custom HTTP Handlers

For applications not using Gin or Echo, you can create custom middleware:

```go
func goInsightMiddleware(client *goinsight.Client) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Start trace
            ctx, traceCtx, err := client.StartTrace(r.Context(), 
                fmt.Sprintf("%s %s", r.Method, r.URL.Path))
            if err == nil {
                r = r.WithContext(ctx)
            }
            
            // Capture response
            recorder := &responseRecorder{ResponseWriter: w, statusCode: 200}
            
            // Process request
            next.ServeHTTP(recorder, r)
            
            // Calculate duration
            duration := time.Since(start)
            
            // Send metric
            go func() {
                metric := goinsight.Metric{
                    Path:       r.URL.Path,
                    Method:     r.Method,
                    StatusCode: recorder.statusCode,
                    Duration:   float64(duration.Nanoseconds()) / 1e6,
                    Source: goinsight.MetricSource{
                        Language:  "go",
                        Framework: "http",
                        Version:   "1.21",
                    },
                }
                client.SendMetric(metric)
            }()
            
            // Log completion
            go func() {
                level := "INFO"
                if recorder.statusCode >= 400 {
                    level = "ERROR"
                }
                
                client.Log(r.Context(), level, "Request completed", map[string]interface{}{
                    "method":      r.Method,
                    "path":        r.URL.Path,
                    "status_code": recorder.statusCode,
                    "duration_ms": duration.Milliseconds(),
                })
            }()
            
            // Finish trace
            if traceCtx != nil {
                go func() {
                    client.FinishSpan(r.Context())
                    client.FinishTrace(r.Context())
                }()
            }
        })
    }
}

type responseRecorder struct {
    http.ResponseWriter
    statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
    r.statusCode = statusCode
    r.ResponseWriter.WriteHeader(statusCode)
}
```

## Advanced Framework Integration

### Custom Metadata

Add framework-specific metadata to all requests:

```go
// For Gin
r.Use(func(c *gin.Context) {
    c.Set("service_version", "1.0.0")
    c.Set("deployment_env", "production")
    c.Next()
})

// In your handler
func handleRequest(c *gin.Context) {
    metadata := map[string]interface{}{
        "version":     c.GetString("service_version"),
        "environment": c.GetString("deployment_env"),
        "user_id":     c.GetHeader("User-ID"),
    }
    
    client.LogInfo(c.Request.Context(), "Request processed", metadata)
}
```

### Error Handling Integration

```go
// Gin error handling
r.Use(gin.Recovery())
r.Use(func(c *gin.Context) {
    defer func() {
        if err := recover(); err != nil {
            client.LogError(c.Request.Context(), "Panic recovered", 
                fmt.Errorf("panic: %v", err), map[string]interface{}{
                    "stack": string(debug.Stack()),
                })
            c.JSON(500, gin.H{"error": "Internal server error"})
        }
    }()
    c.Next()
})

// Echo error handling
e.HTTPErrorHandler = func(err error, c echo.Context) {
    client.LogError(c.Request().Context(), "HTTP error", err, map[string]interface{}{
        "path":   c.Request().URL.Path,
        "method": c.Request().Method,
    })
    
    // Default Echo error handling
    e.DefaultHTTPErrorHandler(err, c)
}
```

### Route Grouping

```go
// Gin route groups with shared context
api := r.Group("/api/v1")
api.Use(func(c *gin.Context) {
    c.Set("api_version", "v1")
    c.Next()
})

users := api.Group("/users")
users.Use(func(c *gin.Context) {
    client.LogDebug(c.Request.Context(), "Accessing users API", map[string]interface{}{
        "version": c.GetString("api_version"),
    })
    c.Next()
})
```

## Performance Considerations

### Middleware Ordering

Place Go-Insight middleware after recovery but before your business logic:

```go
// Gin
r.Use(gin.Recovery())        // 1. Panic recovery
r.Use(client.GinMiddleware()) // 2. Go-Insight instrumentation
r.Use(cors.Default())        // 3. CORS
r.Use(yourCustomMiddleware)  // 4. Your middleware

// Echo  
e.Use(middleware.Recover())   // 1. Panic recovery
e.Use(client.EchoMiddleware()) // 2. Go-Insight instrumentation
e.Use(middleware.CORS())      // 3. CORS
e.Use(yourCustomMiddleware)   // 4. Your middleware
```

### Async Operations

All SDK operations are asynchronous to minimize performance impact:

```go
// These operations don't block your request processing
go client.SendMetric(metric)
go client.Log(ctx, level, message, metadata)
go client.FinishSpan(ctx)
```

### Memory Usage

The middleware has minimal memory overhead:

- ~100 bytes per request for trace context
- Temporary allocations for JSON serialization
- Connection pooling reuses HTTP connections

## Testing

### Testing with Middleware

```go
func TestHandlerWithInstrumentation(t *testing.T) {
    // Create test client
    client := goinsight.New(goinsight.Config{
        APIKey:      "test-key",
        Endpoint:    "http://test-server",
        ServiceName: "test-service",
    })
    
    // Create test router
    r := gin.New()
    r.Use(client.GinMiddleware())
    r.GET("/test", yourHandler)
    
    // Test request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/test", nil)
    r.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

### Mocking for Tests

```go
// Create a no-op client for testing
type MockClient struct{}

func (m *MockClient) LogInfo(ctx context.Context, message string, metadata ...map[string]interface{}) error {
    return nil
}

func (m *MockClient) GinMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Next()
    })
}
```

## Troubleshooting

### Common Issues

**Middleware not working:**
- Ensure middleware is registered before route handlers
- Check API key and endpoint configuration
- Verify Go-Insight server is reachable

**Missing trace context:**
- Ensure you're using the context from the request
- Check middleware ordering
- Verify context propagation in nested function calls

**Performance issues:**
- Ensure async operations are being used
- Check for network connectivity to Go-Insight
- Monitor goroutine creation

See [Troubleshooting Guide](troubleshooting.md) for detailed solutions.