package goinsight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents the Go-Insight SDK client
type Client struct {
	apiKey      string
	endpoint    string
	client      *http.Client
	serviceName string
}

// New creates a new Go-Insight client
func New(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	return &Client{
		apiKey:      config.APIKey,
		endpoint:    config.Endpoint,
		serviceName: config.ServiceName,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Log sends a log entry to Go-Insight
func (c *Client) Log(ctx context.Context, level, message string, metadata map[string]interface{}) error {
	traceCtx := GetTraceFromContext(ctx)

	entry := LogEntry{
		ServiceName: c.serviceName,
		LogLevel:    level,
		Message:     message,
		Metadata:    metadata,
	}

	if traceCtx != nil {
		entry.TraceID = traceCtx.TraceID
		entry.SpanID = traceCtx.SpanID
	}

	return c.sendLog(entry)
}

// LogInfo sends an info log with optional metadata
func (c *Client) LogInfo(ctx context.Context, message string, metadata ...map[string]interface{}) error {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	return c.Log(ctx, "INFO", message, meta)
}

// LogError sends an error log with optional error and metadata
func (c *Client) LogError(ctx context.Context, message string, errAndMetadata ...interface{}) error {
	var err error
	var metadata map[string]interface{}

	for _, arg := range errAndMetadata {
		switch v := arg.(type) {
		case error:
			err = v
		case map[string]interface{}:
			metadata = v
		}
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if err != nil {
		metadata["error"] = err.Error()
	}

	return c.Log(ctx, "ERROR", message, metadata)
}

// LogWarn sends a warning log with optional metadata
func (c *Client) LogWarn(ctx context.Context, message string, metadata ...map[string]interface{}) error {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	return c.Log(ctx, "WARN", message, meta)
}

// LogDebug sends a debug log with optional metadata
func (c *Client) LogDebug(ctx context.Context, message string, metadata ...map[string]interface{}) error {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	return c.Log(ctx, "DEBUG", message, meta)
}

// SendMetric sends a performance metric to Go-Insight
func (c *Client) SendMetric(metric Metric) error {
	if metric.ServiceName == "" {
		metric.ServiceName = c.serviceName
	}

	return c.sendMetric(metric)
}

// HTTP client methods
func (c *Client) sendLog(entry LogEntry) error {
	return c.sendRequest("POST", "/logs", entry)
}

func (c *Client) sendMetric(metric Metric) error {
	return c.sendRequest("POST", "/metrics", metric)
}

func (c *Client) sendTrace(trace Trace) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.sendRequestWithResponse("POST", "/traces", trace, &resp)
	return resp, err
}

func (c *Client) sendSpan(span Span) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.sendRequestWithResponse("POST", "/spans", span, &resp)
	return resp, err
}

func (c *Client) endSpan(spanID string) error {
	return c.sendRequest("POST", fmt.Sprintf("/spans/%s/end", spanID), nil)
}

func (c *Client) endTrace(traceID string) error {
	return c.sendRequest("POST", fmt.Sprintf("/traces/%s/end", traceID), nil)
}

func (c *Client) sendRequest(method, path string, data interface{}) error {
	return c.sendRequestWithResponse(method, path, data, nil)
}

func (c *Client) sendRequestWithResponse(method, path string, data interface{}, response interface{}) error {
	var body []byte
	var err error

	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequest(method, c.endpoint+path, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Instrument wraps a function with automatic instrumentation
func (c *Client) Instrument(operation string, fn func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// Start span
		spanCtx, err := c.StartSpan(ctx, operation)
		if err != nil {
			// If we can't start span, still execute function
			spanCtx = ctx
		}

		// Defer span finishing
		defer func() {
			if err := c.FinishSpan(spanCtx); err != nil {
				// Log error but don't fail the operation
				c.LogError(spanCtx, "Failed to finish span", err, map[string]interface{}{
					"operation": operation,
				})
			}
		}()

		// Execute function
		start := time.Now()
		fnErr := fn(spanCtx)
		duration := time.Since(start)

		// Log operation result
		if fnErr != nil {
			c.LogError(spanCtx, fmt.Sprintf("Operation failed: %s", operation), fnErr, map[string]interface{}{
				"operation":   operation,
				"duration_ms": duration.Milliseconds(),
			})
		} else {
			c.LogInfo(spanCtx, fmt.Sprintf("Operation completed: %s", operation), map[string]interface{}{
				"operation":   operation,
				"duration_ms": duration.Milliseconds(),
			})
		}

		return fnErr
	}
}
