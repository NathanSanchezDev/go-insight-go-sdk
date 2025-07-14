# API Reference

Complete reference for all Go-Insight Go SDK methods and types.

## Client

### New

Creates a new Go-Insight client instance.

```go
func New(config Config) *Client
```

**Parameters:**
- `config` - Configuration for the client

**Returns:**
- `*Client` - Configured Go-Insight client

**Example:**
```go
client := goinsight.New(goinsight.Config{
    APIKey:      "your-api-key",
    Endpoint:    "http://localhost:8080",
    ServiceName: "my-service",
    Timeout:     10 * time.Second,
})
```

## Configuration

### Config

Configuration structure for initializing the Go-Insight client.

```go
type Config struct {
    APIKey      string        // Required: API key for authentication
    Endpoint    string        // Required: Go-Insight server URL
    ServiceName string        // Required: Name of your service
    Timeout     time.Duration // Optional: HTTP timeout (default: 5s)
}
```

## Logging Methods

### Log

Generic logging method that accepts any log level.

```go
func (c *Client) Log(ctx context.Context, level, message string, metadata map[string]interface{}) error
```

### LogInfo

Sends an info-level log entry.

```go
func (c *Client) LogInfo(ctx context.Context, message string, metadata ...map[string]interface{}) error
```

### LogWarn

Sends a warning-level log entry.

```go
func (c *Client) LogWarn(ctx context.Context, message string, metadata ...map[string]interface{}) error
```

### LogError

Sends an error-level log entry.

```go
func (c *Client) LogError(ctx context.Context, message string, errAndMetadata ...interface{}) error
```

**Parameters for LogError:**
- `ctx` - Context for trace correlation
- `message` - Log message
- `errAndMetadata` - Variable arguments that can include:
  - `error` - Go error object
  - `map[string]interface{}` - Metadata

### LogDebug

Sends a debug-level log entry.

```go
func (c *Client) LogDebug(ctx context.Context, message string, metadata ...map[string]interface{}) error
```

## Metrics

### SendMetric

Sends a performance metric to Go-Insight.

```go
func (c *Client) SendMetric(metric Metric) error
```

### Metric Type

```go
type Metric struct {
    ServiceName string                 `json:"service_name"`
    Path        string                 `json:"path"`
    Method      string                 `json:"method"`
    StatusCode  int                    `json:"status_code"`
    Duration    float64                `json:"duration_ms"`
    Source      MetricSource           `json:"source"`
    Environment string                 `json:"environment,omitempty"`
    RequestID   string                 `json:"request_id,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

### MetricSource Type

```go
type MetricSource struct {
    Language  string `json:"language"`
    Framework string `json:"framework"`
    Version   string `json:"version"`
}
```

## Distributed Tracing

### StartTrace

Creates a new distributed trace.

```go
func (c *Client) StartTrace(ctx context.Context, operation string) (context.Context, *TraceContext, error)
```

**Returns:**
- `context.Context` - New context with trace information
- `*TraceContext` - Trace context information
- `error` - Error if trace creation fails

### StartSpan

Creates a new span within an existing trace.

```go
func (c *Client) StartSpan(ctx context.Context, operation string) (context.Context, error)
```

### FinishSpan

Ends the current span.

```go
func (c *Client) FinishSpan(ctx context.Context) error
```

### FinishTrace

Ends the current trace.

```go
func (c *Client) FinishTrace(ctx context.Context) error
```

### GetTraceFromContext

Extracts trace context from a Go context.

```go
func GetTraceFromContext(ctx context.Context) *TraceContext
```

## Middleware

### GinMiddleware

Returns Gin framework middleware for automatic instrumentation.

```go
func (c *Client) GinMiddleware() gin.HandlerFunc
```

**Example:**
```go
r := gin.Default()
r.Use(client.GinMiddleware())
```

### EchoMiddleware

Returns Echo framework middleware for automatic instrumentation.

```go
func (c *Client) EchoMiddleware() echo.MiddlewareFunc
```

**Example:**
```go
e := echo.New()
e.Use(client.EchoMiddleware())
```

## Function Instrumentation

### Instrument

Wraps a function with automatic instrumentation.

```go
func (c *Client) Instrument(operation string, fn func(ctx context.Context) error) func(ctx context.Context) error
```

**Parameters:**
- `operation` - Name of the operation for tracing
- `fn` - Function to instrument

**Returns:**
- Instrumented function that automatically creates spans and logs

**Example:**
```go
processData := client.Instrument("process_user_data", func(ctx context.Context) error {
    // Your business logic here
    time.Sleep(100 * time.Millisecond)
    return nil
})

err := processData(ctx)
```

## Data Types

### TraceContext

Holds trace information for context propagation.

```go
type TraceContext struct {
    TraceID string
    SpanID  string
}
```

### LogEntry

Structure for log entries sent to Go-Insight.

```go
type LogEntry struct {
    ServiceName string                 `json:"service_name"`
    LogLevel    string                 `json:"log_level"`
    Message     string                 `json:"message"`
    TraceID     string                 `json:"trace_id,omitempty"`
    SpanID      string                 `json:"span_id,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

### Trace

Structure for distributed traces.

```go
type Trace struct {
    ID          string `json:"id,omitempty"`
    ServiceName string `json:"service_name"`
}
```

### Span

Structure for spans within traces.

```go
type Span struct {
    ID        string `json:"id,omitempty"`
    TraceID   string `json:"trace_id"`
    ParentID  string `json:"parent_id,omitempty"`
    Service   string `json:"service"`
    Operation string `json:"operation"`
}
```

## Error Handling

All SDK methods that communicate with Go-Insight return errors. These errors should be handled gracefully:

```go
if err := client.LogInfo(ctx, "Operation completed"); err != nil {
    // Log locally or handle as appropriate
    // Never let SDK errors break your application flow
    log.Printf("Failed to send log to Go-Insight: %v", err)
}
```

## Thread Safety

The Go-Insight client is thread-safe and can be used concurrently from multiple goroutines. It's recommended to create a single client instance and reuse it throughout your application.

## Performance Considerations

- All operations are asynchronous by default
- Failed API calls have minimal impact on application performance
- Connection pooling is used for HTTP requests
- Automatic retry logic with exponential backoff (future enhancement)

## Examples

See the [examples directory](../examples/) for complete working examples:

- [Gin Example](../examples/gin-example/) - Complete Gin web server
- [Echo Example](../examples/echo-example/) - Echo framework integration  
- [Manual Instrumentation](../examples/manual-instrumentation/) - Custom tracing patterns