package goinsight

import "time"

// Config holds the configuration for the Go-Insight client
type Config struct {
	APIKey      string
	Endpoint    string
	ServiceName string
	Timeout     time.Duration
}

// LogEntry represents a log entry to be sent to Go-Insight
type LogEntry struct {
	ServiceName string                 `json:"service_name"`
	LogLevel    string                 `json:"log_level"`
	Message     string                 `json:"message"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Metric represents a performance metric
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

// MetricSource represents the source of a metric
type MetricSource struct {
	Language  string `json:"language"`
	Framework string `json:"framework"`
	Version   string `json:"version"`
}

// Trace represents a distributed trace
type Trace struct {
	ID          string `json:"id,omitempty"`
	ServiceName string `json:"service_name"`
}

// Span represents a span within a trace
type Span struct {
	ID        string `json:"id,omitempty"`
	TraceID   string `json:"trace_id"`
	ParentID  string `json:"parent_id,omitempty"`
	Service   string `json:"service"`
	Operation string `json:"operation"`
}

// TraceContext holds trace information in context
type TraceContext struct {
	TraceID string
	SpanID  string
}
