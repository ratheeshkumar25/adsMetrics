package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/metrics"
	"golang.org/x/time/rate"
)

// RequestIDKey is the key for request ID in context
const RequestIDKey = "X-Request-ID"

// RequestLogger middleware for logging HTTP requests
func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Add request ID
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDKey, requestID)

		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		statusStr := strconv.Itoa(statusCode)

		// Log request
		log.Logger.Infow("HTTP Request",
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", statusCode,
			"duration", duration,
			"ip", c.ClientIP(),
			"user_agent", c.GetHeader("User-Agent"))

		// Record metrics
		metrics.RecordHTTPRequest(method, path, statusStr, duration.Seconds())
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is a Swagger route
		path := c.Request.URL.Path
		if len(path) >= 8 && path[:8] == "/swagger" {
			// Relaxed CSP for Swagger UI
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:")
		} else {
			// Strict CSP for API endpoints
			c.Header("Content-Security-Policy", "default-src 'self'")
		}

		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RateLimiter middleware for rate limiting
func RateLimiter() gin.HandlerFunc {
	// Create a rate limiter that allows 100 requests per second with burst of 200
	limiter := rate.NewLimiter(rate.Limit(100), 200)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ErrorHandler middleware for handling panics and errors
func ErrorHandler(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Logger.Errorw("Request panic",
					"error", err,
					"request_id", c.GetString(RequestIDKey))

				c.JSON(500, gin.H{
					"error":      "Internal server error",
					"request_id": c.GetString(RequestIDKey),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Timeout middleware for request timeout
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set timeout in context
		c.Set("timeout", timeout)
		c.Next()
	}
}
