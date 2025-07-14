package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"
)

// UserService demonstrates manual instrumentation patterns
type UserService struct {
	client *goinsight.Client
}

// NewUserService creates a new user service with Go-Insight integration
func NewUserService(client *goinsight.Client) *UserService {
	return &UserService{client: client}
}

// ProcessUser demonstrates manual span creation and trace correlation
func (s *UserService) ProcessUser(ctx context.Context, userID string) error {
	// Start a span for the entire user processing operation
	spanCtx, err := s.client.StartSpan(ctx, "process_user")
	if err != nil {
		return err
	}
	defer s.client.FinishSpan(spanCtx)

	s.client.LogInfo(spanCtx, "Starting user processing", map[string]interface{}{
		"user_id": userID,
	})

	// Step 1: Validate user
	if err := s.validateUser(spanCtx, userID); err != nil {
		return err
	}

	// Step 2: Fetch user data
	userData, err := s.fetchUserData(spanCtx, userID)
	if err != nil {
		return err
	}

	// Step 3: Process data
	if err := s.processData(spanCtx, userData); err != nil {
		return err
	}

	// Step 4: Save results
	if err := s.saveResults(spanCtx, userID, userData); err != nil {
		return err
	}

	s.client.LogInfo(spanCtx, "User processing completed successfully", map[string]interface{}{
		"user_id": userID,
		"steps":   4,
	})

	return nil
}

func (s *UserService) validateUser(ctx context.Context, userID string) error {
	// Create a child span for validation
	spanCtx, err := s.client.StartSpan(ctx, "validate_user")
	if err != nil {
		// Continue without span if creation fails
		spanCtx = ctx
	} else {
		defer s.client.FinishSpan(spanCtx)
	}

	s.client.LogDebug(spanCtx, "Validating user", map[string]interface{}{
		"user_id": userID,
	})

	// Simulate validation time
	time.Sleep(50 * time.Millisecond)

	// Simulate validation logic
	if userID == "invalid" {
		err := fmt.Errorf("user validation failed")
		s.client.LogError(spanCtx, "User validation failed", err, map[string]interface{}{
			"user_id":    userID,
			"validation": "failed",
		})
		return err
	}

	s.client.LogInfo(spanCtx, "User validation successful")
	return nil
}

func (s *UserService) fetchUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	spanCtx, err := s.client.StartSpan(ctx, "fetch_user_data")
	if err != nil {
		spanCtx = ctx
	} else {
		defer s.client.FinishSpan(spanCtx)
	}

	s.client.LogInfo(spanCtx, "Fetching user data from database", map[string]interface{}{
		"user_id": userID,
		"source":  "database",
	})

	// Simulate database query time
	queryTime := time.Duration(rand.Intn(100)+50) * time.Millisecond
	time.Sleep(queryTime)

	userData := map[string]interface{}{
		"id":       userID,
		"name":     "John Doe",
		"email":    "john@example.com",
		"status":   "active",
		"metadata": map[string]interface{}{"region": "us-west-2"},
	}

	s.client.LogInfo(spanCtx, "User data fetched successfully", map[string]interface{}{
		"user_id":    userID,
		"query_time": queryTime.Milliseconds(),
		"fields":     len(userData),
	})

	return userData, nil
}

func (s *UserService) processData(ctx context.Context, userData map[string]interface{}) error {
	spanCtx, err := s.client.StartSpan(ctx, "process_user_data")
	if err != nil {
		spanCtx = ctx
	} else {
		defer s.client.FinishSpan(spanCtx)
	}

	s.client.LogInfo(spanCtx, "Processing user data", map[string]interface{}{
		"data_size": len(userData),
	})

	// Simulate data processing
	for i := 0; i < 3; i++ {
		time.Sleep(30 * time.Millisecond)
		s.client.LogDebug(spanCtx, "Processing step completed", map[string]interface{}{
			"step":     i + 1,
			"progress": fmt.Sprintf("%.1f%%", float64(i+1)/3*100),
		})
	}

	s.client.LogInfo(spanCtx, "Data processing completed")
	return nil
}

func (s *UserService) saveResults(ctx context.Context, userID string, userData map[string]interface{}) error {
	spanCtx, err := s.client.StartSpan(ctx, "save_processed_results")
	if err != nil {
		spanCtx = ctx
	} else {
		defer s.client.FinishSpan(spanCtx)
	}

	s.client.LogInfo(spanCtx, "Saving processed results", map[string]interface{}{
		"user_id":   userID,
		"data_size": len(userData),
	})

	// Simulate save operation
	time.Sleep(75 * time.Millisecond)

	s.client.LogInfo(spanCtx, "Results saved successfully")
	return nil
}

// BatchProcessor demonstrates function decoration pattern
type BatchProcessor struct {
	client *goinsight.Client
}

func NewBatchProcessor(client *goinsight.Client) *BatchProcessor {
	return &BatchProcessor{client: client}
}

func (bp *BatchProcessor) ProcessBatch(ctx context.Context, items []string) error {
	// Use the Instrument decorator for automatic instrumentation
	processBatch := bp.client.Instrument("process_batch", func(ctx context.Context) error {
		bp.client.LogInfo(ctx, "Starting batch processing", map[string]interface{}{
			"batch_size": len(items),
		})

		for i, item := range items {
			// Process each item with its own instrumentation
			processItem := bp.client.Instrument("process_item", func(ctx context.Context) error {
				time.Sleep(time.Duration(rand.Intn(50)+25) * time.Millisecond)

				bp.client.LogDebug(ctx, "Processed item", map[string]interface{}{
					"item":     item,
					"position": i + 1,
				})

				return nil
			})

			if err := processItem(ctx); err != nil {
				return fmt.Errorf("failed to process item %s: %w", item, err)
			}
		}

		bp.client.LogInfo(ctx, "Batch processing completed", map[string]interface{}{
			"processed_count": len(items),
		})

		return nil
	})

	return processBatch(ctx)
}

func main() {
	// Initialize Go-Insight client
	client := goinsight.New(goinsight.Config{
		APIKey:      "your-api-key",
		Endpoint:    "http://localhost:8080",
		ServiceName: "manual-instrumentation-example",
	})

	ctx := context.Background()

	// Example 1: Manual trace and span management
	fmt.Println("=== Example 1: Manual Trace Management ===")

	// Start a new trace
	traceCtx, traceInfo, err := client.StartTrace(ctx, "user_workflow")
	if err != nil {
		fmt.Printf("Failed to start trace: %v\n", err)
		return
	}

	client.LogInfo(traceCtx, "Starting user workflow demonstration")

	// Create user service and process a user
	userService := NewUserService(client)
	if err := userService.ProcessUser(traceCtx, "user123"); err != nil {
		client.LogError(traceCtx, "User processing failed", err)
	}

	// Finish the trace
	client.FinishTrace(traceCtx)
	client.LogInfo(ctx, "Trace completed", map[string]interface{}{
		"trace_id": traceInfo.TraceID,
	})

	// Example 2: Function decoration pattern
	fmt.Println("\n=== Example 2: Function Decoration ===")

	batchProcessor := NewBatchProcessor(client)
	items := []string{"item1", "item2", "item3", "item4", "item5"}

	if err := batchProcessor.ProcessBatch(ctx, items); err != nil {
		client.LogError(ctx, "Batch processing failed", err)
	}

	// Example 3: Error handling and recovery
	fmt.Println("\n=== Example 3: Error Handling ===")

	errorHandlingExample := client.Instrument("error_handling_demo", func(ctx context.Context) error {
		client.LogInfo(ctx, "Demonstrating error handling")

		// Simulate an operation that might fail
		if rand.Float32() > 0.7 { // 30% chance of failure
			return fmt.Errorf("simulated operation failure")
		}

		client.LogInfo(ctx, "Operation completed successfully")
		return nil
	})

	if err := errorHandlingExample(ctx); err != nil {
		client.LogError(ctx, "Operation failed as expected", err)
	}

	// Example 4: Custom metrics
	fmt.Println("\n=== Example 4: Custom Metrics ===")

	// Send custom metrics
	customMetric := goinsight.Metric{
		Path:       "/batch/process",
		Method:     "POST",
		StatusCode: 200,
		Duration:   250.5,
		Source: goinsight.MetricSource{
			Language:  "go",
			Framework: "manual",
			Version:   "1.0.0",
		},
		Environment: "development",
		Metadata: map[string]interface{}{
			"batch_size": len(items),
			"success":    true,
		},
	}

	if err := client.SendMetric(customMetric); err != nil {
		client.LogError(ctx, "Failed to send custom metric", err)
	} else {
		client.LogInfo(ctx, "Custom metric sent successfully")
	}

	fmt.Println("\n=== Manual Instrumentation Examples Completed ===")
	fmt.Println("Check your Go-Insight dashboard to see the traces, logs, and metrics!")
}
