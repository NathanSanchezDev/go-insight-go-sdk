package main

import (
	"context"
	"net/http"
	"time"

	"github.com/NathanSanchezDev/go-insight-go-sdk/goinsight"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize Go-Insight client
	client := goinsight.New(goinsight.Config{
		APIKey:      "your-api-key",
		Endpoint:    "http://localhost:8080",
		ServiceName: "echo-example-service",
	})

	// Create Echo instance
	e := echo.New()

	// Built-in middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Add Go-Insight auto-instrumentation
	e.Use(client.EchoMiddleware())

	// Routes with automatic instrumentation
	e.GET("/users", func(c echo.Context) error {
		// Manual logging with trace correlation
		client.LogInfo(c.Request().Context(), "Fetching users list", map[string]interface{}{
			"page":     1,
			"per_page": 10,
		})

		// Simulate some processing time
		time.Sleep(75 * time.Millisecond)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "Alice", "email": "alice@example.com"},
				{"id": 2, "name": "Bob", "email": "bob@example.com"},
			},
			"total": 2,
		})
	})

	e.POST("/users", func(c echo.Context) error {
		// Start a custom span for validation
		spanCtx, err := client.StartSpan(c.Request().Context(), "validate_user_input")
		if err == nil {
			defer client.FinishSpan(spanCtx)

			client.LogInfo(spanCtx, "Validating user input")
			time.Sleep(25 * time.Millisecond) // Simulate validation
		}

		// Start another span for database operation
		dbSpanCtx, err := client.StartSpan(c.Request().Context(), "create_user_in_db")
		if err == nil {
			defer client.FinishSpan(dbSpanCtx)

			client.LogInfo(dbSpanCtx, "Creating user in database", map[string]interface{}{
				"table": "users",
			})
			time.Sleep(150 * time.Millisecond) // Simulate database save
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "User created successfully",
			"user_id": 123,
		})
	})

	e.GET("/users/:id", func(c echo.Context) error {
		userID := c.Param("id")

		client.LogInfo(c.Request().Context(), "Fetching specific user", map[string]interface{}{
			"user_id": userID,
		})

		// Simulate user lookup
		time.Sleep(30 * time.Millisecond)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
		})
	})

	e.GET("/error", func(c echo.Context) error {
		// Example error logging
		client.LogError(c.Request().Context(), "Simulated error occurred", nil, map[string]interface{}{
			"error_type": "SIMULATION",
			"severity":   "high",
		})

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Internal server error",
			"message": "Something went wrong",
		})
	})

	// Example with manual instrumentation
	e.GET("/process", func(c echo.Context) error {
		// Wrap function with automatic instrumentation
		processTask := client.Instrument("background_processing", func(ctx context.Context) error {
			client.LogInfo(ctx, "Starting background task")

			// Simulate some heavy processing
			for i := 0; i < 3; i++ {
				time.Sleep(50 * time.Millisecond)
				client.LogDebug(ctx, "Processing step completed", map[string]interface{}{
					"step": i + 1,
				})
			}

			client.LogInfo(ctx, "Background task completed")
			return nil
		})

		if err := processTask(c.Request().Context()); err != nil {
			client.LogError(c.Request().Context(), "Background task failed", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "completed",
			"message": "Background processing finished",
		})
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"service":   "echo-example-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Start server
	client.LogInfo(context.Background(), "Starting Echo server", map[string]interface{}{
		"port": "8082",
	})

	e.Logger.Fatal(e.Start(":8082"))
}
