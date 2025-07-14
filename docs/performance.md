# Performance Guide

Optimization strategies and performance characteristics of the Go-Insight SDK.

## Performance Characteristics

### Overhead Measurements

The Go-Insight SDK is designed for production use with minimal performance impact:

| Operation | Overhead | Notes |
|-----------|----------|-------|
| Middleware per request | < 0.5ms | Includes trace creation and metric collection |
| Log entry | < 0.1ms | Asynchronous operation |
| Metric send | < 0.1ms | Asynchronous operation |
| Span creation | < 0.05ms | In-memory operation |
| Span finishing | < 0.1ms | HTTP call, asynchronous |

### Benchmarks

```go
// Benchmark results on standard hardware
BenchmarkGinMiddleware-8        1000000    1156 ns/op    248 B/op    4 allocs/op
BenchmarkLogInfo-8              2000000     654 ns/op    128 B/op    2 allocs/op
BenchmarkSendMetric-8           1500000     892 ns/op    184 B/op    3 allocs/op
BenchmarkStartSpan-8            5000000     321 ns/op     64 B/op    1 allocs/op
```

## Optimization Strategies

### 1. Asynchronous Operations

All SDK operations are asynchronous by default to prevent blocking your application:

```go
// ✅ Good - operations are non-blocking
func handleRequest(c *gin.Context) {
    start := time.Now()
    
    // These don't block request processing
    client.LogInfo(c.Request.Context(), "Processing request")
    
    result := doBusinessLogic() // Your main logic
    
    client.LogInfo(c.Request.Context(), "Request completed")
    c.JSON(200, result)
    
    // Metrics are sent asynchronously after response
}
```

### 2. Connection Pooling

The SDK uses HTTP connection pooling for optimal performance:

```go
// SDK automatically configures connection pooling
client := goinsight.New(goinsight.Config{
    APIKey:      "your-api-key",
    Endpoint:    "http://localhost:8080",
    ServiceName: "my-service",
    Timeout:     5 * time.Second, // Configure timeout for your needs
})

// Internally, the SDK uses:
// - MaxIdleConns: 100
// - MaxIdleConnsPerHost: 10
// - IdleConnTimeout: 90 seconds
```

### 3. Memory Management

#### Context Usage

Use contexts efficiently to avoid memory leaks:

```go
// ✅ Good - proper context usage
func processWithSpan(ctx context.Context) error {
    spanCtx, err := client.StartSpan(ctx, "operation")
    if err == nil {
        defer client.FinishSpan(spanCtx) // Always finish spans
    } else {
        spanCtx = ctx // Fallback to original context
    }
    
    // Use spanCtx for child operations
    return doWork(spanCtx)
}

// ❌ Bad - context leak
func processWithLeak(ctx context.Context) error {
    spanCtx, err := client.StartSpan(ctx, "operation")
    if err != nil {
        return err
    }
    // Missing defer client.FinishSpan(spanCtx) - causes leak!
    
    return doWork(spanCtx)
}
```

#### Metadata Optimization

Keep metadata reasonably sized:

```go
// ✅ Good - concise metadata
client.LogInfo(ctx, "User action", map[string]interface{}{
    "user_id":   userID,
    "action":    "login",
    "timestamp": time.Now().Unix(),
})

// ❌ Bad - excessive metadata
client.LogInfo(ctx, "User action", map[string]interface{}{
    "user_object":    entireUserObject,     // Avoid large objects
    "request_body":   fullRequestPayload,   // Don't log entire payloads
    "system_state":   allSystemVariables,   // Too much data
})
```

### 4. Sampling for High-Volume Services

For services handling millions of requests, implement sampling:

```go
type SamplingConfig struct {
    SampleRate    float32 // 0.0 to 1.0
    AlwaysSample  []string // Always sample these operations
    NeverSample   []string // Never sample these operations
}

func (s *SamplingConfig) ShouldSample(operation string) bool {
    // Always sample critical operations
    for _, op := range s.AlwaysSample {
        if operation == op {
            return true
        }
    }
    
    // Never sample noise operations
    for _, op := range s.NeverSample {
        if operation == op {
            return false
        }
    }
    
    // Random sampling
    return rand.Float32() < s.SampleRate
}

// Usage in middleware
func samplingMiddleware(client *goinsight.Client, config SamplingConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        operation := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
        
        if config.ShouldSample(operation) {
            // Full instrumentation
            client.GinMiddleware()(c)
        } else {
            // Basic processing without observability
            c.Next()
        }
    }
}
```

## Performance Monitoring

### SDK Performance Metrics

Monitor the SDK's own performance:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    sdkOperationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "goinsight_sdk_operation_duration_seconds",
            Help: "Duration of Go-Insight SDK operations",
        },
        []string{"operation"},
    )
    
    sdkErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "goinsight_sdk_errors_total",
            Help: "Total number of Go-Insight SDK errors",
        },
        []string{"operation", "error_type"},
    )
)

// Instrumented client wrapper
type InstrumentedClient struct {
    client *goinsight.Client
}

func (ic *InstrumentedClient) LogInfo(ctx context.Context, message string, metadata ...map[string]interface{}) error {
    timer := prometheus.NewTimer(sdkOperationDuration.WithLabelValues("log"))
    defer timer.ObserveDuration()
    
    err := ic.client.LogInfo(ctx, message, metadata...)
    if err != nil {
        sdkErrors.WithLabelValues("log", "network").Inc()
    }
    
    return err
}
```

### Application Performance Impact

Monitor how the SDK affects your application:

```go
func performanceTestMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        // Track total request duration including SDK overhead
        requestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration.Seconds())
        
        // Log slow requests
        if duration > 500*time.Millisecond {
            log.Printf("Slow request: %s %s took %v", c.Request.Method, c.FullPath(), duration)
        }
    }
}
```

## Optimization Techniques

### 1. Batch Operations

For high-volume scenarios, implement batching:

```go
type BatchSender struct {
    client      *goinsight.Client
    logBatch    []LogEntry
    metricBatch []Metric
    mutex       sync.Mutex
    ticker      *time.Ticker
    batchSize   int
}

func NewBatchSender(client *goinsight.Client, batchSize int, flushInterval time.Duration) *BatchSender {
    bs := &BatchSender{
        client:      client,
        batchSize:   batchSize,
        logBatch:    make([]LogEntry, 0, batchSize),
        metricBatch: make([]Metric, 0, batchSize),
        ticker:      time.NewTicker(flushInterval),
    }
    
    go bs.flushPeriodically()
    return bs
}

func (bs *BatchSender) AddLog(ctx context.Context, level, message string, metadata map[string]interface{}) {
    bs.mutex.Lock()
    defer bs.mutex.Unlock()
    
    bs.logBatch = append(bs.logBatch, LogEntry{
        Level:    level,
        Message:  message,
        Metadata: metadata,
    })
    
    if len(bs.logBatch) >= bs.batchSize {
        bs.flushLogs()
    }
}

func (bs *BatchSender) flushLogs() {
    for _, entry := range bs.logBatch {
        go bs.client.Log(context.Background(), entry.Level, entry.Message, entry.Metadata)
    }
    bs.logBatch = bs.logBatch[:0]
}
```

### 2. Goroutine Pool

Limit goroutine creation for SDK operations:

```go
type WorkerPool struct {
    workers chan chan func()
    quit    chan bool
}

func NewWorkerPool(numWorkers int) *WorkerPool {
    pool := &WorkerPool{
        workers: make(chan chan func(), numWorkers),
        quit:    make(chan bool),
    }
    
    for i := 0; i < numWorkers; i++ {
        worker := NewWorker(pool.workers, pool.quit)
        worker.Start()
    }
    
    return pool
}

func (p *WorkerPool) Submit(job func()) {
    jobChannel := <-p.workers
    jobChannel <- job
}

// Use pool for SDK operations
var sdkWorkerPool = NewWorkerPool(10)

func optimizedLogInfo(ctx context.Context, message string, metadata map[string]interface{}) {
    sdkWorkerPool.Submit(func() {
        client.LogInfo(ctx, message, metadata)
    })
}
```

### 3. Circuit Breaker

Implement circuit breaker pattern for SDK resilience:

```go
type CircuitBreaker struct {
    maxFailures int
    resetTime   time.Duration
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
    mutex       sync.Mutex
}

func (cb *CircuitBreaker) Call(operation func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > cb.resetTime {
            cb.state = "half-open"
            cb.failures = 0
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }
    
    err := operation()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return err
    }
    
    if cb.state == "half-open" {
        cb.state = "closed"
    }
    cb.failures = 0
    
    return nil
}

// Usage with SDK
func resilientLogInfo(ctx context.Context, message string, metadata map[string]interface{}) {
    circuitBreaker.Call(func() error {
        return client.LogInfo(ctx, message, metadata)
    })
}
```

## Performance Testing

### Load Testing Setup

```go
func BenchmarkSDKOperations(b *testing.B) {
    client := goinsight.New(goinsight.Config{
        APIKey:      "test-key",
        Endpoint:    "http://test-server",
        ServiceName: "benchmark-test",
    })
    
    b.ResetTimer()
    
    b.Run("LogInfo", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            client.LogInfo(context.Background(), "test message", map[string]interface{}{
                "iteration": i,
            })
        }
    })
    
    b.Run("SendMetric", func(b *testing.B) {
        metric := goinsight.Metric{
            Path:       "/test",
            Method:     "GET",
            StatusCode: 200,
            Duration:   50.0,
            Source: goinsight.MetricSource{
                Language:  "go",
                Framework: "test",
                Version:   "1.0.0",
            },
        }
        
        for i := 0; i < b.N; i++ {
            client.SendMetric(metric)
        }
    })
    
    b.Run("SpanOperations", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            ctx, _, err := client.StartTrace(context.Background(), "benchmark_trace")
            if err == nil {
                spanCtx, err := client.StartSpan(ctx, "benchmark_span")
                if err == nil {
                    client.FinishSpan(spanCtx)
                }
                client.FinishTrace(ctx)
            }
        }
    })
}
```

### Concurrent Load Testing

```go
func TestConcurrentOperations(t *testing.T) {
    client := goinsight.New(goinsight.Config{
        APIKey:      "test-key",
        Endpoint:    "http://test-server",
        ServiceName: "concurrent-test",
    })
    
    const numGoroutines = 100
    const operationsPerGoroutine = 1000
    
    var wg sync.WaitGroup
    start := time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(goroutineID int) {
            defer wg.Done()
            
            for j := 0; j < operationsPerGoroutine; j++ {
                client.LogInfo(context.Background(), "concurrent test", map[string]interface{}{
                    "goroutine_id": goroutineID,
                    "operation_id": j,
                })
            }
        }(i)
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalOps := numGoroutines * operationsPerGoroutine
    opsPerSecond := float64(totalOps) / duration.Seconds()
    
    t.Logf("Completed %d operations in %v (%.2f ops/sec)", totalOps, duration, opsPerSecond)
    
    // Verify no goroutine leaks
    time.Sleep(100 * time.Millisecond)
    finalGoroutines := runtime.NumGoroutine()
    t.Logf("Final goroutine count: %d", finalGoroutines)
}
```

## Memory Optimization

### Context Pool

Reuse contexts to reduce allocations:

```go
var contextPool = sync.Pool{
    New: func() interface{} {
        return context.Background()
    },
}

func getContext() context.Context {
    return contextPool.Get().(context.Context)
}

func putContext(ctx context.Context) {
    // Clear any values before returning to pool
    contextPool.Put(context.Background())
}

// Usage
func optimizedOperation() {
    ctx := getContext()
    defer putContext(ctx)
    
    client.LogInfo(ctx, "optimized operation")
}
```

### Metadata Pool

Reuse metadata maps:

```go
var metadataPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]interface{}, 8) // Pre-allocate common size
    },
}

func getMetadata() map[string]interface{} {
    return metadataPool.Get().(map[string]interface{})
}

func putMetadata(m map[string]interface{}) {
    // Clear the map
    for k := range m {
        delete(m, k)
    }
    metadataPool.Put(m)
}

// Usage
func optimizedLogging(userID string, action string) {
    metadata := getMetadata()
    defer putMetadata(metadata)
    
    metadata["user_id"] = userID
    metadata["action"] = action
    metadata["timestamp"] = time.Now().Unix()
    
    client.LogInfo(context.Background(), "user action", metadata)
}
```

## Network Optimization

### Request Batching

Batch multiple operations into single HTTP requests:

```go
type BatchRequest struct {
    Logs    []LogEntry `json:"logs,omitempty"`
    Metrics []Metric   `json:"metrics,omitempty"`
}

func (c *Client) SendBatch(batch BatchRequest) error {
    return c.sendRequest("POST", "/batch", batch)
}

// Accumulate operations and send in batches
type BatchAccumulator struct {
    client    *goinsight.Client
    batch     BatchRequest
    mutex     sync.Mutex
    lastFlush time.Time
    maxSize   int
    maxAge    time.Duration
}

func (ba *BatchAccumulator) AddLog(entry LogEntry) {
    ba.mutex.Lock()
    defer ba.mutex.Unlock()
    
    ba.batch.Logs = append(ba.batch.Logs, entry)
    ba.checkFlush()
}

func (ba *BatchAccumulator) checkFlush() {
    totalItems := len(ba.batch.Logs) + len(ba.batch.Metrics)
    age := time.Since(ba.lastFlush)
    
    if totalItems >= ba.maxSize || age >= ba.maxAge {
        ba.flush()
    }
}

func (ba *BatchAccumulator) flush() {
    if len(ba.batch.Logs) == 0 && len(ba.batch.Metrics) == 0 {
        return
    }
    
    go ba.client.SendBatch(ba.batch)
    
    ba.batch = BatchRequest{}
    ba.lastFlush = time.Now()
}
```

### Connection Keep-Alive

Optimize HTTP client settings:

```go
func createOptimizedHTTPClient(timeout time.Duration) *http.Client {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
        
        // TCP optimizations
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        
        // TLS optimizations
        TLSHandshakeTimeout: 10 * time.Second,
    }
    
    return &http.Client{
        Timeout:   timeout,
        Transport: transport,
    }
}
```

## Profiling and Monitoring

### Built-in Profiling

Add profiling endpoints to monitor SDK performance:

```go
import _ "net/http/pprof"

func enableProfiling() {
    go func() {
        log.Println("Profiling server started on :6060")
        log.Println(http.ListenAndServe(":6060", nil))
    }()
}

// Usage:
// go tool pprof http://localhost:6060/debug/pprof/profile
// go tool pprof http://localhost:6060/debug/pprof/heap
```

### Custom Metrics

Track SDK-specific metrics:

```go
type SDKMetrics struct {
    operationCount    map[string]int64
    errorCount        map[string]int64
    averageDuration   map[string]time.Duration
    mutex            sync.RWMutex
}

func (sm *SDKMetrics) RecordOperation(operation string, duration time.Duration, err error) {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    
    sm.operationCount[operation]++
    
    if err != nil {
        sm.errorCount[operation]++
    }
    
    // Simple moving average
    current := sm.averageDuration[operation]
    count := sm.operationCount[operation]
    sm.averageDuration[operation] = time.Duration(
        (int64(current)*(count-1) + int64(duration)) / count,
    )
}

func (sm *SDKMetrics) GetStats() map[string]interface{} {
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()
    
    return map[string]interface{}{
        "operation_count":    sm.operationCount,
        "error_count":       sm.errorCount,
        "average_duration":  sm.averageDuration,
    }
}
```

## Production Recommendations

### 1. Configuration

```go
// Production-optimized configuration
client := goinsight.New(goinsight.Config{
    APIKey:      os.Getenv("GO_INSIGHT_API_KEY"),
    Endpoint:    os.Getenv("GO_INSIGHT_ENDPOINT"),
    ServiceName: os.Getenv("GO_INSIGHT_SERVICE_NAME"),
    Timeout:     3 * time.Second, // Shorter timeout for production
})
```

### 2. Graceful Shutdown

```go
func gracefulShutdown(client *goinsight.Client) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    <-c
    log.Println("Shutting down gracefully...")
    
    // Allow time for pending SDK operations
    time.Sleep(2 * time.Second)
    
    log.Println("Shutdown complete")
    os.Exit(0)
}
```

### 3. Health Checks

```go
func healthCheck(c *gin.Context) {
    // Don't include SDK health in application health
    // SDK failures shouldn't mark your app as unhealthy
    
    appHealth := checkApplicationHealth()
    
    c.JSON(200, gin.H{
        "status": appHealth.Status,
        "timestamp": time.Now(),
        "version": appVersion,
    })
}
```

### 4. Error Handling

```go
func productionErrorHandling(client *goinsight.Client) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Log panic but don't fail if SDK fails
                _ = client.LogError(c.Request.Context(), "Application panic", 
                    fmt.Errorf("panic: %v", err), map[string]interface{}{
                        "stack": string(debug.Stack()),
                    })
                
                c.JSON(500, gin.H{"error": "Internal server error"})
                c.Abort()
            }
        }()
        
        c.Next()
    })
}
```

## Performance Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Check for context leaks
   - Verify spans are being finished
   - Monitor goroutine count

2. **High Latency**
   - Check network connectivity to Go-Insight
   - Verify timeout settings
   - Monitor SDK error rates

3. **CPU Usage**
   - Profile the application
   - Check for excessive logging
   - Verify async operations

### Debugging Commands

```bash
# Check goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Check application metrics
curl http://localhost:8080/metrics
```