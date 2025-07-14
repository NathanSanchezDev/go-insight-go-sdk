package goinsight

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

// GinMiddleware returns a Gin middleware for automatic instrumentation
func (c *Client) GinMiddleware() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		start := time.Now()

		// Start trace for this request
		ctx, traceCtx, err := c.StartTrace(ginCtx.Request.Context(), fmt.Sprintf("%s %s", ginCtx.Request.Method, ginCtx.FullPath()))
		if err == nil {
			ginCtx.Request = ginCtx.Request.WithContext(ctx)
			ginCtx.Set("go-insight-trace", traceCtx)
		}

		// Process request
		ginCtx.Next()

		// Calculate duration
		duration := time.Since(start)

		// Send metric asynchronously
		go func() {
			metric := Metric{
				ServiceName: c.serviceName,
				Path:        ginCtx.FullPath(),
				Method:      ginCtx.Request.Method,
				StatusCode:  ginCtx.Writer.Status(),
				Duration:    float64(duration.Nanoseconds()) / 1e6, // Convert to milliseconds
				Source: MetricSource{
					Language:  "go",
					Framework: "gin",
					Version:   gin.Version,
				},
				RequestID: ginCtx.GetHeader("X-Request-ID"),
			}
			c.SendMetric(metric)
		}()

		// Log request completion asynchronously
		go func() {
			level := "INFO"
			if ginCtx.Writer.Status() >= 400 {
				level = "ERROR"
			} else if ginCtx.Writer.Status() >= 300 {
				level = "WARN"
			}

			metadata := map[string]interface{}{
				"method":      ginCtx.Request.Method,
				"path":        ginCtx.FullPath(),
				"status_code": ginCtx.Writer.Status(),
				"duration_ms": duration.Milliseconds(),
				"user_agent":  ginCtx.GetHeader("User-Agent"),
			}

			c.Log(ginCtx.Request.Context(), level, fmt.Sprintf("Request completed: %s %s", ginCtx.Request.Method, ginCtx.FullPath()), metadata)
		}()

		// Finish trace asynchronously
		if traceCtx != nil {
			go func() {
				c.FinishSpan(ginCtx.Request.Context())
				c.FinishTrace(ginCtx.Request.Context())
			}()
		}
	}
}

// EchoMiddleware returns an Echo middleware for automatic instrumentation
func (c *Client) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			start := time.Now()

			// Start trace for this request
			ctx, traceCtx, err := c.StartTrace(echoCtx.Request().Context(), fmt.Sprintf("%s %s", echoCtx.Request().Method, echoCtx.Path()))
			if err == nil {
				echoCtx.SetRequest(echoCtx.Request().WithContext(ctx))
				echoCtx.Set("go-insight-trace", traceCtx)
			}

			// Process request
			err = next(echoCtx)

			// Calculate duration
			duration := time.Since(start)

			// Get status code
			statusCode := echoCtx.Response().Status
			if statusCode == 0 {
				statusCode = 200
			}

			// Send metric asynchronously
			go func() {
				metric := Metric{
					ServiceName: c.serviceName,
					Path:        echoCtx.Path(),
					Method:      echoCtx.Request().Method,
					StatusCode:  statusCode,
					Duration:    float64(duration.Nanoseconds()) / 1e6, // Convert to milliseconds
					Source: MetricSource{
						Language:  "go",
						Framework: "echo",
						Version:   echo.Version,
					},
					RequestID: echoCtx.Request().Header.Get("X-Request-ID"),
				}
				c.SendMetric(metric)
			}()

			// Log request completion asynchronously
			go func() {
				level := "INFO"
				if statusCode >= 400 {
					level = "ERROR"
				} else if statusCode >= 300 {
					level = "WARN"
				}

				metadata := map[string]interface{}{
					"method":      echoCtx.Request().Method,
					"path":        echoCtx.Path(),
					"status_code": statusCode,
					"duration_ms": duration.Milliseconds(),
					"user_agent":  echoCtx.Request().Header.Get("User-Agent"),
				}

				c.Log(echoCtx.Request().Context(), level, fmt.Sprintf("Request completed: %s %s", echoCtx.Request().Method, echoCtx.Path()), metadata)
			}()

			// Finish trace asynchronously
			if traceCtx != nil {
				go func() {
					c.FinishSpan(echoCtx.Request().Context())
					c.FinishTrace(echoCtx.Request().Context())
				}()
			}

			return err
		}
	}
}
