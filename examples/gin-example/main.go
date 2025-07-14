// examples/gin-example/main.go
package main

import (
	"context"
	"time"

	"github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Go-Insight client
	client := goinsight.New(goinsight.Config{
		APIKey:      "your-api-key",
		Endpoint:    "http://localhost:8080",
		ServiceName: "gin-example-service",
	})

	// Setup Gin with auto-instrumentation
	r := gin.Default()
	r.Use(client.GinMiddleware())

	// Your routes are automatically instrumented
	r.GET("/users", func(c *gin.Context) {
		// Manual logging with trace correlation
		client.LogInfo(c.Request.Context(), "Fetching users", map[string]interface{}{
			"user_count": 42,
		})

		// Simulate some work
		time.Sleep(50 * time.Millisecond)

		c.JSON(200, gin.H{"users": []string{"alice", "bob"}})
	})

	r.POST("/users", func(c *gin.Context) {
		// Start a custom span for database operation
		spanCtx, err := client.StartSpan(c.Request.Context(), "database_save_user")
		if err == nil {
			defer client.FinishSpan(spanCtx)

			// Simulate database save
			time.Sleep(100 * time.Millisecond)

			client.LogInfo(spanCtx, "User saved to database", map[string]interface{}{
				"user_id": "123",
			})
		}

		c.JSON(201, gin.H{"message": "User created"})
	})

	r.GET("/error", func(c *gin.Context) {
		// Example error logging
		client.LogError(c.Request.Context(), "Something went wrong", nil, map[string]interface{}{
			"error_code": "EXAMPLE_ERROR",
		})

		c.JSON(500, gin.H{"error": "Internal server error"})
	})

	// Manual instrumentation example
	r.GET("/manual", func(c *gin.Context) {
		// Wrap function with automatic instrumentation
		processData := client.Instrument("process_user_data", func(ctx context.Context) error {
			time.Sleep(75 * time.Millisecond)
			client.LogInfo(ctx, "Data processing complete")
			return nil
		})

		if err := processData(c.Request.Context()); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "processed"})
	})

	r.Run(":8081")
}
