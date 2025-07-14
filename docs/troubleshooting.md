# Troubleshooting Guide

Common issues and solutions when using the Go-Insight SDK.

## Installation Issues

### Import Path Errors

**Problem:** Import path not found or module not recognized
```
go: module github.com/NathanSanchezDev/go-insight-go-sdk not found
```

**Solutions:**
1. Check the exact repository URL:
   ```bash
   go get github.com/NathanSanchezDev/go-insight-go-sdk
   ```

2. For local development, use replace directive:
   ```go
   // go.mod
   replace github.com/NathanSanchezDev/go-insight-go-sdk => ../path/to/local/sdk
   ```

3. Verify Go module is enabled:
   ```bash
   export GO111MODULE=on
   go mod tidy
   ```

### Version Conflicts

**Problem:** Dependency version conflicts
```
go: inconsistent vendoring: module requires different version
```

**Solutions:**
1. Clean module cache:
   ```bash
   go clean -modcache
   go mod download
   ```

2. Update to latest version:
   ```bash
   go get -u github.com/NathanSanchezDev/go-insight-go-sdk
   ```

## Configuration Issues

### Authentication Failures

**Problem:** API requests failing with 401 Unauthorized
```
API request failed with status 401
```

**Solutions:**
1. Verify API key is correct:
   ```go
   // Check environment variable
   apiKey := os.Getenv("GO_INSIGHT_API_KEY")
   if apiKey == "" {
       log.Fatal("GO_INSIGHT_API_KEY environment variable not set")
   }
   ```

2. Test API key manually:
   ```bash
   curl -H "X-API-Key: your-api-key" http://localhost:8080/health
   ```

3. Check for trailing spaces in environment variables:
   ```bash
   # Remove any whitespace
   export GO_INSIGHT_API_KEY=$(echo $GO_INSIGHT_API_KEY | tr -d '[:space:]')
   ```

### Connection Issues

**Problem:** Cannot reach Go-Insight server
```
failed to send request: dial tcp: connection refused
```

**Solutions:**
1. Verify Go-Insight server is running:
   ```bash
   curl http://localhost:8080/health
   ```

2. Check endpoint URL format:
   ```go
   // ✅ Correct
   Endpoint: "http://localhost:8080"
   
   // ❌ Incorrect
   Endpoint: "localhost:8080"        // Missing protocol
   Endpoint: "http://localhost:8080/" // Trailing slash
   ```

3. Test network connectivity:
   ```bash
   telnet localhost 8080
   ```

4. Check firewall and network policies:
   ```bash
   # For Docker environments
   docker network ls
   docker network inspect bridge
   ```

## Runtime Issues

### Missing Trace Context

**Problem:** Logs not correlated with traces
```
Trace context is nil in handler
```

**Solutions:**
1. Ensure middleware is properly registered:
   ```go
   r := gin.Default()
   r.Use(client.GinMiddleware()) // Must be before route handlers
   r.GET("/users", handleUsers)
   ```

2. Check middleware order:
   ```go
   // ✅ Correct order
   r.Use(gin.Recovery())
   r.Use(client.GinMiddleware())
   r.Use(cors.Default())
   
   // ❌ Wrong order
   r.Use(cors.Default())
   r.Use(client.GinMiddleware()) // Too late
   ```

3. Verify context propagation:
   ```go
   func handleRequest(c *gin.Context) {
       ctx := c.Request.Context()
       traceCtx := goinsight.GetTraceFromContext(ctx)
       
       if traceCtx == nil {
           log.Println("No trace context found!")
           return
       }
       
       // Context is available
       client.LogInfo(ctx, "Processing request")
   }
   ```

### Memory Leaks

**Problem:** Increasing memory usage over time
```
runtime: VirtualAlloc of 1048576 bytes failed
```

**Solutions:**
1. Check for unfinished spans:
   ```go
   func processRequest(ctx context.Context) error {
       spanCtx, err := client.StartSpan(ctx, "process")
       if err == nil {
           defer client.FinishSpan(spanCtx) // ✅ Always finish spans
       }
       
       return doWork(spanCtx)
   }
   ```

2. Monitor goroutine count:
   ```go
   import _ "net/http/pprof"
   
   // Access http://localhost:6060/debug/pprof/goroutine
   go func() {
       log.Println(http.ListenAndServe(":6060", nil))
   }()
   ```

3. Use context with timeout:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   
   client.LogInfo(ctx, "Timed operation")
   ```

### Performance Issues

**Problem:** High latency or CPU usage
```
Request processing taking too long
```

**Solutions:**
1. Verify async operations:
   ```go
   // ✅ Async (default behavior)
   client.LogInfo(ctx, "Message") // Returns immediately
   
   // ❌ Don't block on SDK operations
   if err := client.LogInfo(ctx, "Message"); err != nil {
       return err // Don't fail business logic due to SDK
   }
   ```

2. Check timeout settings:
   ```go
   client := goinsight.New(goinsight.Config{
       Timeout: 3 * time.Second, // Shorter timeout for production
   })
   ```

3. Profile the application:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/profile
   ```

## Framework-Specific Issues

### Gin Framework

**Problem:** Middleware not working with Gin
```
Routes not being traced
```

**Solutions:**
1. Check Gin version compatibility:
   ```bash
   go list -m github.com/gin-gonic/gin
   # Should be v1.9.1 or later
   ```

2. Verify middleware registration:
   ```go
   func main() {
       client := goinsight.New(config)
       
       r := gin.Default()
       r.Use(client.GinMiddleware()) // Must call GinMiddleware()
       
       r.GET("/test", func(c *gin.Context) {
           c.JSON(200, gin.H{"status": "ok"})
       })
       
       r.Run(":8080")
   }
   ```

3. Test middleware in isolation:
   ```go
   func TestGinMiddleware(t *testing.T) {
       client := goinsight.New(testConfig)
       
       r := gin.New()
       r.Use(client.GinMiddleware())
       r.GET("/test", func(c *gin.Context) {
           c.String(200, "ok")
       })
       
       w := httptest.NewRecorder()
       req, _ := http.NewRequest("GET", "/test", nil)
       r.ServeHTTP(w, req)
       
       assert.Equal(t, 200, w.Code)
   }
   ```

### Echo Framework

**Problem:** Echo middleware not instrumenting requests
```
Echo routes not generating traces
```

**Solutions:**
1. Check Echo version:
   ```bash
   go list -m github.com/labstack/echo/v4
   # Should be v4.11.3 or later
   ```

2. Proper middleware usage:
   ```go
   func main() {
       client := goinsight.New(config)
       
       e := echo.New()
       e.Use(client.EchoMiddleware()) // Must call EchoMiddleware()
       
       e.GET("/test", func(c echo.Context) error {
           return c.JSON(200, map[string]string{"status": "ok"})
       })
       
       e.Logger.Fatal(e.Start(":8080"))
   }
   ```

3. Error handling integration:
   ```go
   e.HTTPErrorHandler = func(err error, c echo.Context) {
       // Custom error handling that doesn't interfere with SDK
       e.DefaultHTTPErrorHandler(err, c)
   }
   ```

## Data Issues

### Missing Logs

**Problem:** Logs not appearing in Go-Insight
```
Logs sent but not visible in dashboard
```

**Solutions:**
1. Check log level filtering:
   ```bash
   # Query specific log levels
   curl -H "X-API-Key: your-key" \
     "http://localhost:8080/logs?level=INFO&service=your-service"
   ```

2. Verify service name:
   ```bash
   # Check what services exist
   curl -H "X-API-Key: your-key" \
     "http://localhost:8080/logs" | jq '.[].service_name' | sort | uniq
   ```

3. Test log sending directly:
   ```go
   func testLogging() {
       client := goinsight.New(config)
       
       err := client.LogInfo(context.Background(), "Test log message", map[string]interface{}{
           "test": true,
           "timestamp": time.Now().Unix(),
       })
       
       if err != nil {
           log.Printf("Failed to send log: %v", err)
       } else {
           log.Println("Log sent successfully")
       }
   }
   ```

### Trace Correlation Issues

**Problem:** Logs and traces not correlated
```
Logs and traces appear separately in Go-Insight
```

**Solutions:**
1. Verify trace context usage:
   ```go
   func handleRequest(c *gin.Context) {
       // Use the context from the request (has trace info)
       ctx := c.Request.Context()
       
       client.LogInfo(ctx, "Processing") // ✅ Will be correlated
       
       // ❌ Don't use background context
       // client.LogInfo(context.Background(), "Processing")
   }
   ```

2. Check trace ID propagation:
   ```go
   func debugTraceContext(ctx context.Context) {
       traceCtx := goinsight.GetTraceFromContext(ctx)
       if traceCtx != nil {
           log.Printf("Trace ID: %s, Span ID: %s", 
               traceCtx.TraceID, traceCtx.SpanID)
       } else {
           log.Println("No trace context found")
       }
   }
   ```

3. Manual trace creation:
   ```go
   func createManualTrace() {
       ctx, traceCtx, err := client.StartTrace(context.Background(), "manual_operation")
       if err != nil {
           log.Printf("Failed to start trace: %v", err)
           return
       }
       defer client.FinishTrace(ctx)
       
       client.LogInfo(ctx, "Manual trace test")
       log.Printf("Created trace: %s", traceCtx.TraceID)
   }
   ```

## Testing Issues

### Unit Testing

**Problem:** Tests failing due to SDK calls
```
Tests timing out or failing on SDK operations
```

**Solutions:**
1. Use test configuration:
   ```go
   func getTestClient() *goinsight.Client {
       return goinsight.New(goinsight.Config{
           APIKey:      "test-key",
           Endpoint:    "http://localhost:8080",
           ServiceName: "test-service",
           Timeout:     1 * time.Second, // Shorter timeout for tests
       })
   }
   ```

2. Mock the SDK for unit tests:
   ```go
   type MockSDK struct{}
   
   func (m *MockSDK) LogInfo(ctx context.Context, message string, metadata ...map[string]interface{}) error {
       return nil // No-op for tests
   }
   
   func (m *MockSDK) GinMiddleware() gin.HandlerFunc {
       return gin.HandlerFunc(func(c *gin.Context) {
           c.Next() // Pass through without instrumentation
       })
   }
   ```

3. Use build tags:
   ```go
   // +build integration
   
   func TestWithRealSDK(t *testing.T) {
       // Only run with: go test -tags=integration
   }
   ```

### Integration Testing

**Problem:** Integration tests unreliable
```
Tests pass locally but fail in CI
```

**Solutions:**
1. Use test containers:
   ```go
   func setupTestEnvironment(t *testing.T) *goinsight.Client {
       // Start Go-Insight test container
       container := startGoInsightContainer(t)
       t.Cleanup(func() { container.Stop() })
       
       endpoint := fmt.Sprintf("http://localhost:%s", container.Port())
       
       return goinsight.New(goinsight.Config{
           APIKey:      "test-key",
           Endpoint:    endpoint,
           ServiceName: "integration-test",
       })
   }
   ```

2. Wait for service readiness:
   ```go
   func waitForService(endpoint string, timeout time.Duration) error {
       deadline := time.Now().Add(timeout)
       
       for time.Now().Before(deadline) {
           resp, err := http.Get(endpoint + "/health")
           if err == nil && resp.StatusCode == 200 {
               resp.Body.Close()
               return nil
           }
           time.Sleep(100 * time.Millisecond)
       }
       
       return fmt.Errorf("service not ready after %v", timeout)
   }
   ```

## Debugging Tools

### Enable Debug Logging

```go
import "log"

func enableDebugLogging() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.Println("Debug logging enabled")
}

// Custom logger for SDK operations
type DebugLogger struct {
    client *goinsight.Client
}

func (dl *DebugLogger) LogInfo(ctx context.Context, message string, metadata ...map[string]interface{}) error {
    log.Printf("[SDK] LogInfo: %s, metadata: %+v", message, metadata)
    return dl.client.LogInfo(ctx, message, metadata...)
}
```

### Network Debugging

```bash
# Monitor HTTP traffic
tcpdump -i lo -A -s 0 'port 8080'

# Check DNS resolution
nslookup your-goinsight-server.com

# Test with curl
curl -v -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"service_name":"test","message":"debug"}' \
  http://localhost:8080/logs
```

### Performance Debugging

```go
// Add request timing
func debugMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        log.Printf("[DEBUG] %s %s took %v", 
            c.Request.Method, c.Request.URL.Path, duration)
    }
}
```

## Getting Help

### Information to Provide

When reporting issues, include:

1. **Environment information:**
   ```bash
   go version
   echo $GOOS $GOARCH
   go list -m github.com/NathanSanchezDev/go-insight-go-sdk
   ```

2. **Configuration (sanitized):**
   ```go
   // Remove sensitive data like API keys
   config := goinsight.Config{
       APIKey:      "[REDACTED]",
       Endpoint:    "http://localhost:8080",
       ServiceName: "my-service",
       Timeout:     5 * time.Second,
   }
   ```

3. **Error messages and stack traces**
4. **Minimal reproducible example**
5. **Expected vs actual behavior**

### Support Channels

- 🐛 [GitHub Issues](https://github.com/NathanSanchezDev/go-insight-go-sdk/issues) - Bug reports and feature requests
- 💬 [GitHub Discussions](https://github.com/NathanSanchezDev/go-insight-go-sdk/discussions) - Questions and community support
- 📖 [Documentation](README.md) - Comprehensive guides and examples

### Before Opening an Issue

1. Search existing issues and discussions
2. Check the documentation and examples
3. Test with the latest version
4. Provide a minimal reproducible example
5. Include relevant logs and error messages