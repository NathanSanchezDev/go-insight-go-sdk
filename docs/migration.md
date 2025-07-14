# Migration Guide

Guide for migrating between versions and from other observability solutions.

## Version Migrations

### From 0.0.x to 0.1.0

This is the first stable release. Key changes include:

**Breaking Changes:**
- None - this is the initial stable release

**New Features:**
- Complete middleware support for Gin and Echo
- Automatic trace correlation
- Function instrumentation with `Instrument()` method
- Improved error handling with variadic parameters

**Migration Steps:**
```bash
# Update to latest version
go get github.com/NathanSanchezDev/go-insight-go-sdk@v0.1.0
go mod tidy
```

## Migrating from Other Solutions

### From Manual HTTP Calls

**Before:**
```go
// Manual HTTP calls to Go-Insight
func sendLog(message string) error {
    data := map[string]interface{}{
        "service_name": "my-service",
        "log_level":    "INFO",
        "message":      message,
    }
    
    jsonData, _ := json.Marshal(data)
    req, _ := http.NewRequest("POST", "http://localhost:8080/logs", 
        bytes.NewBuffer(jsonData))
    req.Header.Set("X-API-Key", "your-api-key")
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

**After:**
```go
// Using Go-Insight SDK
client := goinsight.New(goinsight.Config{
    APIKey:      "your-api-key",
    Endpoint:    "http://localhost:8080",
    ServiceName: "my-service",
})

// Simple logging with automatic trace correlation
client.LogInfo(ctx, "User action completed")
```

**Benefits:**
- Automatic trace correlation
- Connection pooling and retry logic
- Type safety and validation
- Async operations by default
- Comprehensive error handling

### From OpenTelemetry

**Before (OpenTelemetry):**
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/attribute"
)

func instrumentedFunction(ctx context.Context) error {
    tracer := otel.Tracer("my-service")
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("user.id", "12345"),
        attribute.String("operation", "process_data"),
    )
    
    // Your business logic
    return processData()
}
```

**After (Go-Insight SDK):**
```go
func instrumentedFunction(ctx context.Context) error {
    // Automatic instrumentation
    processData := client.Instrument("process_data", func(ctx context.Context) error {
        client.LogInfo(ctx, "Processing data", map[string]interface{}{
            "user_id": "12345",
            "operation": "process_data",
        })
        
        // Your business logic
        return processData()
    })
    
    return processData(ctx)
}
```

**Migration Benefits:**
- Simpler API with less boilerplate
- Automatic correlation between logs, metrics, and traces
- Built-in middleware for popular frameworks
- No need to configure exporters or collectors

### From Prometheus + Jaeger

**Before (Prometheus metrics + Jaeger tracing):**
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
)

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"method", "path", "status"},
    )
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Start Jaeger span
    span := opentracing.StartSpan("handle_request")
    defer span.Finish()
    
    // Your business logic
    processRequest()
    
    // Record Prometheus metric
    duration := time.Since(start)
    requestDuration.WithLabelValues(
        r.Method, r.URL.Path, "200",
    ).Observe(duration.Seconds())
}
```

**After (Go-Insight SDK):**
```go
func main() {
    client := goinsight.New(goinsight.Config{
        APIKey:      "your-api-key",
        Endpoint:    "http://localhost:8080",
        ServiceName: "my-service",
    })
    
    r := gin.Default()
    r.Use(client.GinMiddleware()) // Automatic metrics + tracing + logging
    
    r.GET("/api", func(c *gin.Context) {
        // Your business logic - everything is automatically instrumented
        processRequest()
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.Run(":8080")
}
```

**Migration Benefits:**
- Single tool instead of multiple solutions
- No need to manually instrument every endpoint
- Automatic correlation between all observability data
- Simplified configuration and deployment

### From Custom Logging Libraries

**Before (logrus/zap):**
```go
import (
    "github.com/sirupsen/logrus"
    "go.uber.org/zap"
)

// Logrus example
func processUser(userID string) {
    logrus.WithFields(logrus.Fields{
        "user_id": userID,
        "service": "user-service",
    }).Info("Processing user")
    
    // Business logic
    
    logrus.WithFields(logrus.Fields{
        "user_id": userID,
        "duration": "150ms",
    }).Info("User processed")
}

// Zap example
func processUserZap(userID string) {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    logger.Info("Processing user",
        zap.String("user_id", userID),
        zap.String("service", "user-service"),
    )
    
    // Business logic
    
    logger.Info("User processed",
        zap.String("user_id", userID),
        zap.String("duration", "150ms"),
    )
}
```

**After (Go-Insight SDK):**
```go
func processUser(ctx context.Context, userID string) {
    client.LogInfo(ctx, "Processing user", map[string]interface{}{
        "user_id": userID,
    })
    
    // Business logic with automatic span tracking
    processUserData := client.Instrument("process_user_data", func(ctx context.Context) error {
        // Your business logic here
        time.Sleep(150 * time.Millisecond)
        return nil
    })
    
    if err := processUserData(ctx); err != nil {
        client.LogError(ctx, "User processing failed", err, map[string]interface{}{
            "user_id": userID,
        })
        return
    }
    
    client.LogInfo(ctx, "User processed", map[string]interface{}{
        "user_id": userID,
    })
}
```

**Migration Benefits:**
- Automatic trace correlation across all logs
- Built-in performance metrics
- Centralized observability data
- Distributed tracing without additional setup

## Migration Strategies

### Gradual Migration

**Phase 1: Parallel Logging**
```go
// Run both systems in parallel during migration
func dualLogging(ctx context.Context, message string, fields map[string]interface{}) {
    // Existing logging (keep during migration)
    logrus.WithFields(logrus.Fields(fields)).Info(message)
    
    // New Go-Insight logging
    client.LogInfo(ctx, message, fields)
}
```

**Phase 2: Framework Integration**
```go
// Add Go-Insight middleware while keeping existing instrumentation
r := gin.Default()

// Keep existing middleware during transition
r.Use(existingLoggingMiddleware())
r.Use(existingMetricsMiddleware())

// Add Go-Insight middleware
r.Use(client.GinMiddleware())

r.GET("/api", handleAPI)
```

**Phase 3: Replace Custom Instrumentation**
```go
// Replace custom instrumentation gradually
func migrateCustomSpans(ctx context.Context) error {
    // Old way (remove after migration)
    // span := tracer.StartSpan("operation")
    // defer span.Finish()
    
    // New way
    spanCtx, err := client.StartSpan(ctx, "operation")
    if err == nil {
        defer client.FinishSpan(spanCtx)
    } else {
        spanCtx = ctx
    }
    
    return doWork(spanCtx)
}
```

**Phase 4: Remove Legacy Systems**
```go
// Final phase - clean up old dependencies
// Remove from go.mod:
// - github.com/sirupsen/logrus
// - go.opentelemetry.io/otel
// - github.com/prometheus/client_golang
```

### Big Bang Migration

For smaller applications, complete migration in one step:

```go
// Before: Multiple observability tools
func setupObservability() {
    // Prometheus setup
    prometheus.MustRegister(requestDuration)
    
    // Jaeger setup
    tracer, closer := jaeger.NewTracer(...)
    defer closer.Close()
    
    // Logrus setup
    logrus.SetFormatter(&logrus.JSONFormatter{})
}

// After: Single SDK setup
func setupObservability() {
    client := goinsight.New(goinsight.Config{
        APIKey:      os.Getenv("GO_INSIGHT_API_KEY"),
        Endpoint:    os.Getenv("GO_INSIGHT_ENDPOINT"),
        ServiceName: os.Getenv("GO_INSIGHT_SERVICE_NAME"),
    })
    
    return client
}
```

## Configuration Migration

### Environment Variables

**Before (Multiple tools):**
```bash
# Jaeger
export JAEGER_AGENT_HOST=localhost
export JAEGER_AGENT_PORT=6831
export JAEGER_SERVICE_NAME=my-service

# Prometheus
export PROMETHEUS_GATEWAY=localhost:9091

# Custom logging
export LOG_LEVEL=info
export LOG_FORMAT=json
```

**After (Go-Insight SDK):**
```bash
# Simplified configuration
export GO_INSIGHT_API_KEY=your-api-key
export GO_INSIGHT_ENDPOINT=http://localhost:8080
export GO_INSIGHT_SERVICE_NAME=my-service
export GO_INSIGHT_TIMEOUT=5s
```

### Docker Configuration

**Before:**
```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    environment:
      - JAEGER_AGENT_HOST=jaeger
      - PROMETHEUS_GATEWAY=prometheus:9091
    depends_on:
      - jaeger
      - prometheus
      
  jaeger:
    image: jaegertracing/all-in-one:latest
    
  prometheus:
    image: prom/prometheus:latest
```

**After:**
```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    environment:
      - GO_INSIGHT_API_KEY=${GO_INSIGHT_API_KEY}
      - GO_INSIGHT_ENDPOINT=http://go-insight:8080
      - GO_INSIGHT_SERVICE_NAME=my-service
    depends_on:
      - go-insight
      
  go-insight:
    image: go-insight:latest
```

## Testing Migration

### Before and After Comparison

```go
func TestMigrationEquivalence(t *testing.T) {
    // Test that new SDK produces equivalent observability data
    
    // Setup both old and new systems
    oldTracer := setupJaeger()
    newClient := setupGoInsight()
    
    // Run same operation with both
    ctx := context.Background()
    
    // Old way
    oldSpan := oldTracer.StartSpan("test_operation")
    oldSpan.SetTag("user_id", "123")
    oldSpan.Finish()
    
    // New way
    spanCtx, err := newClient.StartSpan(ctx, "test_operation")
    require.NoError(t, err)
    newClient.LogInfo(spanCtx, "Operation completed", map[string]interface{}{
        "user_id": "123",
    })
    newClient.FinishSpan(spanCtx)
    
    // Verify both systems captured the data
    // (Implementation depends on your verification strategy)
}
```

### Performance Testing

```go
func BenchmarkOldVsNew(b *testing.B) {
    oldLogger := logrus.New()
    newClient := goinsight.New(testConfig)
    
    b.Run("Old", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            oldLogger.WithFields(logrus.Fields{
                "iteration": i,
            }).Info("Test message")
        }
    })
    
    b.Run("New", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            newClient.LogInfo(context.Background(), "Test message", map[string]interface{}{
                "iteration": i,
            })
        }
    })
}
```

## Rollback Plan

### Safe Migration with Feature Flags

```go
type ObservabilityConfig struct {
    UseGoInsight bool `env:"USE_GO_INSIGHT" default:"false"`
    UseLogrus    bool `env:"USE_LOGRUS" default:"true"`
    UseJaeger    bool `env:"USE_JAEGER" default:"true"`
}

func setupObservability(config ObservabilityConfig) {
    if config.UseGoInsight {
        client := goinsight.New(goInsightConfig)
        // Use Go-Insight SDK
    }
    
    if config.UseLogrus {
        logrus.SetFormatter(&logrus.JSONFormatter{})
        // Keep existing logrus setup
    }
    
    if config.UseJaeger {
        tracer, _ := jaeger.NewTracer(jaegerConfig)
        // Keep existing Jaeger setup
    }
}
```

### Gradual Rollout

```go
func shouldUseNewSDK(userID string) bool {
    // Gradual rollout based on user ID hash
    hash := fnv.New32a()
    hash.Write([]byte(userID))
    return hash.Sum32()%100 < 10 // 10% of users
}

func conditionalLogging(ctx context.Context, userID, message string) {
    if shouldUseNewSDK(userID) {
        client.LogInfo(ctx, message, map[string]interface{}{
            "user_id": userID,
        })
    } else {
        logrus.WithFields(logrus.Fields{
            "user_id": userID,
        }).Info(message)
    }
}
```

## Migration Checklist

### Pre-Migration
- [ ] Identify all current observability tools
- [ ] Document current logging patterns
- [ ] Set up Go-Insight instance
- [ ] Create migration timeline
- [ ] Plan rollback strategy

### During Migration
- [ ] Install Go-Insight SDK
- [ ] Update dependencies
- [ ] Migrate configuration
- [ ] Update middleware
- [ ] Replace manual instrumentation
- [ ] Update tests

### Post-Migration
- [ ] Verify data completeness
- [ ] Compare performance metrics
- [ ] Remove old dependencies
- [ ] Update documentation
- [ ] Train team on new SDK

### Validation
- [ ] All logs are being captured
- [ ] Traces are properly correlated
- [ ] Metrics are accurate
- [ ] Performance is acceptable
- [ ] Error handling works correctly

## Common Migration Issues

### Data Format Differences

**Issue:** Log format changes between systems
**Solution:** Use metadata mapping

```go
func mapLogrusToGoInsight(entry *logrus.Entry) map[string]interface{} {
    metadata := make(map[string]interface{})
    
    for key, value := range entry.Data {
        metadata[key] = value
    }
    
    // Map logrus levels to Go-Insight levels
    level := strings.ToUpper(entry.Level.String())
    if level == "WARNING" {
        level = "WARN"
    }
    
    return metadata
}
```

### Context Propagation

**Issue:** Different context handling patterns
**Solution:** Create adapter functions

```go
func adaptContext(oldCtx YourOldContext) context.Context {
    ctx := context.Background()
    
    // Extract relevant data from old context
    if traceID := oldCtx.GetTraceID(); traceID != "" {
        ctx = context.WithValue(ctx, "trace_id", traceID)
    }
    
    return ctx
}
```

### Performance Differences

**Issue:** Different performance characteristics
**Solution:** Gradual migration with monitoring

```go
func monitoredMigration() {
    // Monitor both old and new systems during transition
    oldDuration := measureOldSystem()
    newDuration := measureNewSystem()
    
    if newDuration > oldDuration*1.5 {
        // Rollback if performance degrades significantly
        rollbackToOldSystem()
    }
}
```