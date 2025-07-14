# Best Practices

Guidelines for using Go-Insight SDK effectively in production environments.

## Configuration Best Practices

### Environment-Based Configuration

Use environment variables for different deployment environments:

```go
// config.go
type Config struct {
    GoInsight struct {
        APIKey      string `env:"GO_INSIGHT_API_KEY"`
        Endpoint    string `env:"GO_INSIGHT_ENDPOINT"`
        ServiceName string `env:"GO_INSIGHT_SERVICE_NAME"`
        Timeout     time.Duration `env:"GO_INSIGHT_TIMEOUT" default:"5s"`
    }
}

// main.go
func main() {
    cfg := loadConfig()
    
    client := goinsight.New(goinsight.Config{
        APIKey:      cfg.GoInsight.APIKey,
        Endpoint:    cfg.GoInsight.Endpoint,
        ServiceName: cfg.GoInsight.ServiceName,
        Timeout:     cfg.GoInsight.Timeout,
    })
}
```

### Service Naming Convention

Use consistent service naming across environments:

```bash
# Development
GO_INSIGHT_SERVICE_NAME="user-service-dev"

# Staging  
GO_INSIGHT_SERVICE_NAME="user-service-staging"

# Production
GO_INSIGHT_SERVICE_NAME="user-service-prod"
```

### API Key Management

- **Never hardcode API keys** in source code
- Use environment variables or secret management systems
- Rotate API keys regularly (recommended: every 90 days)
- Use different API keys for different environments

```go
// ❌ Don't do this
client := goinsight.New(goinsight.Config{
    APIKey: "abc123-hardcoded-key",
})

// ✅ Do this
client := goinsight.New(goinsight.Config{
    APIKey: os.Getenv("GO_INSIGHT_API_KEY"),
})
```

## Logging Best Practices

### Structured Logging

Use consistent metadata structure across your application:

```go
// Define standard fields
type LogFields struct {
    UserID    string `json:"user_id,omitempty"`
    RequestID string `json:"request_id,omitempty"`
    Operation string `json:"operation,omitempty"`
    Duration  int64  `json:"duration_ms,omitempty"`
}

func (l LogFields) ToMap() map[string]interface{} {
    data := make(map[string]interface{})
    if l.UserID != "" {
        data["user_id"] = l.UserID
    }
    if l.RequestID != "" {
        data["request_id"] = l.RequestID
    }
    // ... add other fields
    return data
}

// Usage
fields := LogFields{
    UserID:    "user123",
    RequestID: c.GetHeader("X-Request-ID"),
    Operation: "create_user",
}

client.LogInfo(ctx, "User creation started", fields.ToMap())
```

### Log Levels

Use appropriate log levels:

```go
// DEBUG: Detailed information for debugging
client.LogDebug(ctx, "Database query executed", map[string]interface{}{
    "query": "SELECT * FROM users WHERE id = ?",
    "params": []interface{}{userID},
})

// INFO: General information about application flow
client.LogInfo(ctx, "User authentication successful", map[string]interface{}{
    "user_id": userID,
    "method": "password",
})

// WARN: Warning conditions that should be addressed
client.LogWarn(ctx, "Rate limit approaching", map[string]interface{}{
    "current_rate": 95,
    "limit": 100,
})

// ERROR: Error conditions that require attention
client.LogError(ctx, "Database connection failed", err, map[string]interface{}{
    "database": "primary",
    "retry_count": 3,
})
```

### Sensitive Data

Never log sensitive information:

```go
// ❌ Don't log sensitive data
client.LogInfo(ctx, "User login", map[string]interface{}{
    "password": password,        // Never log passwords
    "credit_card": ccNumber,     // Never log payment info
    "ssn": socialSecurity,       // Never log PII
})

// ✅ Log safely
client.LogInfo(ctx, "User login", map[string]interface{}{
    "user_id": userID,
    "login_method": "password",
    "success": true,
})
```

## Tracing Best Practices

### Span Naming

Use descriptive and consistent span names:

```go
// ❌ Vague names
client.StartSpan(ctx, "work")
client.StartSpan(ctx, "stuff")

// ✅ Descriptive names
client.StartSpan(ctx, "fetch_user_from_database")
client.StartSpan(ctx, "validate_payment_details")
client.StartSpan(ctx, "send_notification_email")
```

### Span Hierarchy

Create logical span hierarchies:

```go
func ProcessOrder(ctx context.Context, orderID string) error {
    // Root span for the entire operation
    spanCtx, err := client.StartSpan(ctx, "process_order")
    if err == nil {
        defer client.FinishSpan(spanCtx)
    } else {
        spanCtx = ctx
    }
    
    // Child span for validation
    if err := validateOrder(spanCtx, orderID); err != nil {
        return err
    }
    
    // Child span for payment
    if err := processPayment(spanCtx, orderID); err != nil {
        return err
    }
    
    // Child span for fulfillment
    return fulfillOrder(spanCtx, orderID)
}

func validateOrder(ctx context.Context, orderID string) error {
    spanCtx, err := client.StartSpan(ctx, "validate_order")
    if err == nil {
        defer client.FinishSpan(spanCtx)
    } else {
        spanCtx = ctx
    }
    
    // Validation logic
    return nil
}
```

### Function Instrumentation

Use the `Instrument` method for automatic span management:

```go
// Manual span management (more control)
func processDataManual(ctx context.Context) error {
    spanCtx, err := client.StartSpan(ctx, "process_data")
    if err == nil {
        defer client.FinishSpan(spanCtx)
    } else {
        spanCtx = ctx
    }
    
    // Your logic here
    return nil
}

// Automatic instrumentation (simpler)
var processDataAuto = client.Instrument("process_data", func(ctx context.Context) error {
    // Your logic here
    return nil
})
```

## Error Handling Best Practices

### Graceful Degradation

Never let SDK errors break your application:

```go
func handleRequest(c *gin.Context) {
    // Your business logic should never fail due to observability
    user, err := fetchUser(userID)
    if err != nil {
        // Log the business error
        if logErr := client.LogError(c.Request.Context(), "Failed to fetch user", err, map[string]interface{}{
            "user_id": userID,
        }); logErr != nil {
            // Don't fail the request due to logging failure
            log.Printf("Failed to send error log: %v", logErr)
        }
        
        c.JSON(500, gin.H{"error": "User not found"})
        return
    }
    
    c.JSON(200, user)
}
```

### Error Context

Provide rich context in error logs:

```go
func processPayment(ctx context.Context, paymentInfo PaymentInfo) error {
    err := chargeCard(paymentInfo)
    if err != nil {
        client.LogError(ctx, "Payment processing failed", err, map[string]interface{}{
            "payment_method": paymentInfo.Method,
            "amount":        paymentInfo.Amount,
            "currency":      paymentInfo.Currency,
            "merchant_id":   paymentInfo.MerchantID,
            "attempt":       paymentInfo.AttemptNumber,
            "error_code":    extractErrorCode(err),
        })
        return err
    }
    
    return nil
}
```

## Performance Best Practices

### Async Operations

SDK operations are async by default, but ensure you're not blocking:

```go
// ✅ Good - operations are async
func handleRequest(c *gin.Context) {
    client.LogInfo(c.Request.Context(), "Processing request") // Async
    
    result := doBusinessLogic()
    
    client.LogInfo(c.Request.Context(), "Request completed") // Async
    c.JSON(200, result)
}

// ❌ Don't wait for SDK operations
func handleRequestBad(c *gin.Context) {
    if err := client.LogInfo(c.Request.Context(), "Processing request"); err != nil {
        // Don't fail the request due to logging
        c.JSON(500, gin.H{"error": "Logging failed"})
        return
    }
}
```

### Sampling for High-Volume Services

For very high-volume services, consider implementing sampling:

```go
func shouldSample() bool {
    return rand.Float32() < 0.1 // Sample 10% of requests
}

func handleHighVolumeRequest(c *gin.Context) {
    if shouldSample() {
        client.LogInfo(c.Request.Context(), "High volume request sampled")
    }
    
    // Always process the business logic
    result := processRequest()
    c.JSON(200, result)
}
```

### Metadata Size

Keep metadata reasonably sized:

```go
// ❌ Too much metadata
client.LogInfo(ctx, "Processing", map[string]interface{}{
    "huge_object": massiveDataStructure, // Avoid large objects
    "full_request_body": requestBody,    // Don't log entire payloads
})

// ✅ Reasonable metadata
client.LogInfo(ctx, "Processing", map[string]interface{}{
    "object_id":   obj.ID,
    "object_type": obj.Type,
    "payload_size": len(requestBody),
})
```

## Production Deployment

### Health Checks

Don't include SDK health in application health checks:

```go
func healthCheck(c *gin.Context) {
    // Check your application health, not SDK health
    if !database.IsHealthy() {
        c.JSON(503, gin.H{"status": "unhealthy", "reason": "database"})
        return
    }
    
    // SDK failures shouldn't affect health status
    c.JSON(200, gin.H{"status": "healthy"})
}
```

### Monitoring SDK Performance

Monitor SDK performance separately:

```go
var (
    sdkErrorCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "goinsight_sdk_errors_total",
            Help: "Total number of Go-Insight SDK errors",
        },
        []string{"operation"},
    )
)

func monitoredLog(ctx context.Context, level, message string, metadata map[string]interface{}) {
    if err := client.Log(ctx, level, message, metadata); err != nil {
        sdkErrorCount.WithLabelValues("log").Inc()
        log.Printf("SDK error: %v", err)
    }