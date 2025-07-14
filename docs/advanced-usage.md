# Advanced Usage

Advanced patterns and techniques for using Go-Insight SDK in complex applications.

## Custom Span Management

### Creating Complex Span Hierarchies

```go
func ProcessComplexWorkflow(ctx context.Context, workflowID string) error {
    // Root span for entire workflow
    workflowCtx, err := client.StartSpan(ctx, "process_workflow")
    if err == nil {
        defer client.FinishSpan(workflowCtx)
    } else {
        workflowCtx = ctx
    }

    client.LogInfo(workflowCtx, "Starting complex workflow", map[string]interface{}{
        "workflow_id": workflowID,
    })

    // Phase 1: Data Collection
    dataCtx, err := client.StartSpan(workflowCtx, "collect_data")
    if err == nil {
        defer client.FinishSpan(dataCtx)
        
        // Multiple data sources
        if err := collectFromAPI(dataCtx, "source1"); err != nil {
            return err
        }
        if err := collectFromDatabase(dataCtx, "source2"); err != nil {
            return err
        }
    }

    // Phase 2: Data Processing
    processCtx, err := client.StartSpan(workflowCtx, "process_data")
    if err == nil {
        defer client.FinishSpan(processCtx)
        
        if err := transformData(processCtx); err != nil {
            return err
        }
        if err := validateData(processCtx); err != nil {
            return err
        }
    }

    // Phase 3: Results Storage
    storeCtx, err := client.StartSpan(workflowCtx, "store_results")
    if err == nil {
        defer client.FinishSpan(storeCtx)
        
        if err := saveToDatabase(storeCtx); err != nil {
            return err
        }
        if err := updateCache(storeCtx); err != nil {
            return err
        }
    }

    client.LogInfo(workflowCtx, "Complex workflow completed")
    return nil
}
```

### Parallel Span Execution

```go
func ProcessParallelTasks(ctx context.Context, tasks []Task) error {
    // Root span for parallel processing
    rootCtx, err := client.StartSpan(ctx, "parallel_processing")
    if err == nil {
        defer client.FinishSpan(rootCtx)
    } else {
        rootCtx = ctx
    }

    var wg sync.WaitGroup
    errChan := make(chan error, len(tasks))

    for i, task := range tasks {
        wg.Add(1)
        go func(taskIndex int, t Task) {
            defer wg.Done()
            
            // Each goroutine gets its own span
            taskCtx, err := client.StartSpan(rootCtx, fmt.Sprintf("process_task_%d", taskIndex))
            if err == nil {
                defer client.FinishSpan(taskCtx)
            } else {
                taskCtx = rootCtx
            }
            
            client.LogInfo(taskCtx, "Starting parallel task", map[string]interface{}{
                "task_id": t.ID,
                "task_index": taskIndex,
            })
            
            if err := processTask(taskCtx, t); err != nil {
                errChan <- err
                return
            }
            
            client.LogInfo(taskCtx, "Parallel task completed")
        }(i, task)
    }

    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        if err != nil {
            client.LogError(rootCtx, "Parallel task failed", err)
            return err
        }
    }

    return nil
}
```

## Cross-Service Tracing

### Propagating Trace Context

```go
// Service A - Starting service
func callDownstreamService(ctx context.Context, data interface{}) error {
    // Extract trace context
    traceCtx := goinsight.GetTraceFromContext(ctx)
    if traceCtx == nil {
        return fmt.Errorf("no trace context available")
    }

    // Create HTTP request to Service B
    req, err := http.NewRequest("POST", "http://service-b/process", nil)
    if err != nil {
        return err
    }

    // Add trace headers for propagation
    req.Header.Set("X-Trace-ID", traceCtx.TraceID)
    req.Header.Set("X-Span-ID", traceCtx.SpanID)
    req.Header.Set("X-Service-Name", "service-a")

    client.LogInfo(ctx, "Calling downstream service", map[string]interface{}{
        "trace_id": traceCtx.TraceID,
        "downstream_service": "service-b",
    })

    // Make the request
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

// Service B - Receiving service
func handleIncomingRequest(c *gin.Context) {
    // Extract trace context from headers
    traceID := c.GetHeader("X-Trace-ID")
    parentSpanID := c.GetHeader("X-Span-ID")
    sourceService := c.GetHeader("X-Service-Name")

    if traceID != "" {
        // Create child span in existing trace
        spanCtx, err := client.StartSpan(c.Request.Context(), "process_downstream_request")
        if err == nil {
            defer client.FinishSpan(spanCtx)
            
            client.LogInfo(spanCtx, "Processing request from upstream service", map[string]interface{}{
                "trace_id": traceID,
                "parent_span_id": parentSpanID,
                "source_service": sourceService,
            })
            
            // Process the request
            result := processRequest()
            
            c.JSON(200, result)
        }
    } else {
        // No trace context, start new trace
        client.LogWarn(c.Request.Context(), "No trace context from upstream service")
        // Handle normally...
    }
}
```

### Custom Context Propagation

```go
// Custom context key type for type safety
type contextKey string

const (
    userIDKey    contextKey = "user_id"
    requestIDKey contextKey = "request_id"
    tenantIDKey  contextKey = "tenant_id"
)

// Helper functions for context management
func withUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDKey, userID)
}

func getUserID(ctx context.Context) string {
    if userID, ok := ctx.Value(userIDKey).(string); ok {
        return userID
    }
    return ""
}

// Middleware to extract and propagate custom context
func contextPropagationMiddleware(client *goinsight.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        
        // Extract custom headers
        if userID := c.GetHeader("X-User-ID"); userID != "" {
            ctx = withUserID(ctx, userID)
        }
        
        if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
            ctx = context.WithValue(ctx, requestIDKey, requestID)
        }
        
        // Update request context
        c.Request = c.Request.WithContext(ctx)
        
        // Log with enriched context
        client.LogInfo(ctx, "Request started", map[string]interface{}{
            "user_id":    getUserID(ctx),
            "request_id": ctx.Value(requestIDKey),
        })
        
        c.Next()
    }
}
```

## Advanced Metrics

### Custom Business Metrics

```go
// Business metrics structure
type BusinessMetrics struct {
    client *goinsight.Client
}

func NewBusinessMetrics(client *goinsight.Client) *BusinessMetrics {
    return &BusinessMetrics{client: client}
}

func (bm *BusinessMetrics) TrackUserAction(ctx context.Context, action string, userID string, duration time.Duration) {
    metric := goinsight.Metric{
        Path:       fmt.Sprintf("/business/%s", action),
        Method:     "BUSINESS",
        StatusCode: 200,
        Duration:   float64(duration.Milliseconds()),
        Source: goinsight.MetricSource{
            Language:  "go",
            Framework: "business",
            Version:   "1.0.0",
        },
        Metadata: map[string]interface{}{
            "user_id":    userID,
            "action":     action,
            "event_type": "user_action",
        },
    }
    
    bm.client.SendMetric(metric)
}

func (bm *BusinessMetrics) TrackSalesFunnel(ctx context.Context, stage string, value float64) {
    metric := goinsight.Metric{
        Path:       fmt.Sprintf("/funnel/%s", stage),
        Method:     "FUNNEL",
        StatusCode: 200,
        Duration:   0, // Not time-based
        Source: goinsight.MetricSource{
            Language:  "go",
            Framework: "business",
            Version:   "1.0.0",
        },
        Metadata: map[string]interface{}{
            "stage":      stage,
            "value":      value,
            "event_type": "sales_funnel",
        },
    }
    
    bm.client.SendMetric(metric)
}

// Usage in handlers
func handlePurchase(c *gin.Context) {
    start := time.Now()
    userID := getUserID(c.Request.Context())
    
    // Business logic
    purchase, err := processPurchase(userID)
    if err != nil {
        businessMetrics.TrackUserAction(c.Request.Context(), "purchase_failed", userID, time.Since(start))
        c.JSON(500, gin.H{"error": "Purchase failed"})
        return
    }
    
    // Track successful purchase
    businessMetrics.TrackUserAction(c.Request.Context(), "purchase_success", userID, time.Since(start))
    businessMetrics.TrackSalesFunnel(c.Request.Context(), "conversion", purchase.Amount)
    
    c.JSON(200, purchase)
}
```

### Performance Monitoring

```go
// Performance monitoring wrapper
type PerformanceMonitor struct {
    client *goinsight.Client
}

func NewPerformanceMonitor(client *goinsight.Client) *PerformanceMonitor {
    return &PerformanceMonitor{client: client}
}

func (pm *PerformanceMonitor) MonitorDatabaseQuery(query string) func() {
    start := time.Now()
    return func() {
        duration := time.Since(start)
        
        metric := goinsight.Metric{
            Path:       "/database/query",
            Method:     "DB",
            StatusCode: 200,
            Duration:   float64(duration.Milliseconds()),
            Source: goinsight.MetricSource{
                Language:  "go",
                Framework: "database",
                Version:   "1.0.0",
            },
            Metadata: map[string]interface{}{
                "query_type": extractQueryType(query),
                "table":      extractTableName(query),
            },
        }
        
        pm.client.SendMetric(metric)
        
        // Log slow queries
        if duration > 100*time.Millisecond {
            pm.client.LogWarn(context.Background(), "Slow database query detected", map[string]interface{}{
                "query":       query,
                "duration_ms": duration.Milliseconds(),
            })
        }
    }
}

// Usage
func fetchUser(userID string) (*User, error) {
    defer performanceMonitor.MonitorDatabaseQuery("SELECT * FROM users WHERE id = ?")()
    
    // Database query logic
    return database.QueryUser(userID)
}
```

## Event-Driven Architecture

### Event Publishing

```go
type EventPublisher struct {
    client *goinsight.Client
}

func (ep *EventPublisher) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
    // Start span for event publishing
    spanCtx, err := ep.client.StartSpan(ctx, "publish_event")
    if err == nil {
        defer ep.client.FinishSpan(spanCtx)
    } else {
        spanCtx = ctx
    }
    
    ep.client.LogInfo(spanCtx, "Publishing event", map[string]interface{}{
        "event_type": eventType,
        "payload_size": getPayloadSize(payload),
    })
    
    // Publish to message queue (implementation specific)
    if err := messageQueue.Publish(eventType, payload); err != nil {
        ep.client.LogError(spanCtx, "Failed to publish event", err, map[string]interface{}{
            "event_type": eventType,
        })
        return err
    }
    
    // Track as metric
    metric := goinsight.Metric{
        Path:       fmt.Sprintf("/events/%s", eventType),
        Method:     "PUBLISH",
        StatusCode: 200,
        Duration:   0,
        Source: goinsight.MetricSource{
            Language:  "go",
            Framework: "events",
            Version:   "1.0.0",
        },
        Metadata: map[string]interface{}{
            "event_type": eventType,
        },
    }
    
    ep.client.SendMetric(metric)
    return nil
}
```

### Event Consumption

```go
func handleUserCreatedEvent(ctx context.Context, event UserCreatedEvent) error {
    // Start trace for event processing
    traceCtx, _, err := client.StartTrace(ctx, "process_user_created_event")
    if err != nil {
        traceCtx = ctx
    }
    
    client.LogInfo(traceCtx, "Processing user created event", map[string]interface{}{
        "user_id": event.UserID,
        "event_id": event.EventID,
    })
    
    // Process in phases with separate spans
    if err := sendWelcomeEmail(traceCtx, event.UserID); err != nil {
        return err
    }
    
    if err := setupUserProfile(traceCtx, event.UserID); err != nil {
        return err
    }
    
    if err := trackUserRegistration(traceCtx, event); err != nil {
        return err
    }
    
    client.FinishTrace(traceCtx)
    return nil
}
```

## Testing Advanced Patterns

### Testing with Custom Context

```go
func TestComplexWorkflowWithTracing(t *testing.T) {
    // Create test client
    client := goinsight.New(goinsight.Config{
        APIKey:      "test-key",
        Endpoint:    "http://test-server",
        ServiceName: "test-service",
    })
    
    // Create context with trace
    ctx, traceCtx, err := client.StartTrace(context.Background(), "test_workflow")
    require.NoError(t, err)
    defer client.FinishTrace(ctx)
    
    // Test the workflow
    err = ProcessComplexWorkflow(ctx, "test-workflow-123")
    assert.NoError(t, err)
    
    // Verify trace context exists
    assert.NotNil(t, traceCtx)
    assert.NotEmpty(t, traceCtx.TraceID)
}
```

### Mocking for Integration Tests

```go
type MockGoInsightClient struct {
    logs    []LogEntry
    metrics []Metric
    traces  []string
    spans   []string
}

func (m *MockGoInsightClient) LogInfo(ctx context.Context, message string, metadata ...map[string]interface{}) error {
    m.logs = append(m.logs, LogEntry{
        Level:   "INFO",
        Message: message,
        Context: ctx,
    })
    return nil
}

func (m *MockGoInsightClient) StartSpan(ctx context.Context, operation string) (context.Context, error) {
    spanID := fmt.Sprintf("span-%d", len(m.spans))
    m.spans = append(m.spans, spanID)
    
    // Create new context with span
    return context.WithValue(ctx, "span_id", spanID), nil
}

// Verification methods
func (m *MockGoInsightClient) GetLogCount() int {
    return len(m.logs)
}

func (m *MockGoInsightClient) GetSpanCount() int {
    return len(m.spans)
}
```

## Performance Optimization

### Batching Operations

```go
type BatchLogger struct {
    client     *goinsight.Client
    batch      []LogEntry
    batchSize  int
    mutex      sync.Mutex
    ticker     *time.Ticker
}

func NewBatchLogger(client *goinsight.Client, batchSize int, flushInterval time.Duration) *BatchLogger {
    bl := &BatchLogger{
        client:    client,
        batchSize: batchSize,
        batch:     make([]LogEntry, 0, batchSize),
        ticker:    time.NewTicker(flushInterval),
    }
    
    go bl.flushPeriodically()
    return bl
}

func (bl *BatchLogger) LogAsync(ctx context.Context, level, message string, metadata map[string]interface{}) {
    bl.mutex.Lock()
    defer bl.mutex.Unlock()
    
    bl.batch = append(bl.batch, LogEntry{
        Level:    level,
        Message:  message,
        Metadata: metadata,
        Context:  ctx,
    })
    
    if len(bl.batch) >= bl.batchSize {
        bl.flushBatch()
    }
}

func (bl *BatchLogger) flushBatch() {
    if len(bl.batch) == 0 {
        return
    }
    
    // Send batch to Go-Insight
    for _, entry := range bl.batch {
        go bl.client.Log(entry.Context, entry.Level, entry.Message, entry.Metadata)
    }
    
    bl.batch = bl.batch[:0] // Reset slice
}
```