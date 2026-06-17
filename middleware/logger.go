package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"ai-content-platform/database"
	"ai-content-platform/models"

	"github.com/gin-gonic/gin"
)

// LoggerToFile logs requests to the database
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Read the request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		}
		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user ID from context if available
		userIDValue, exists := c.Get("userID")
		var userID uint
		if !exists {
			userID = 0 // System operation
		} else {
			// Handle different possible types for userID
			switch v := userIDValue.(type) {
			case uint:
				userID = v
			case int:
				userID = uint(v)
			case int64:
				userID = uint(v)
			case float64: // JSON numbers are parsed as float64
				userID = uint(v)
			default:
				userID = 0
			}
		}

		// Log to database
		logEntry := models.SystemLog{
			UserID:      userID,
			Action:      c.Request.Method + " " + c.Request.URL.Path,
			ObjectType:  "",
			ObjectID:    0,
			Description: c.Request.Method + " " + c.Request.URL.Path + " - Status: " + fmt.Sprintf("%d", c.Writer.Status()),
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
		}

		// Save log entry to database
		if database.DB != nil {
			database.DB.Create(&logEntry)
		}

		// Also log to console
		fmt.Printf("[GIN] %s | %s | %s | %s | %s | %s | %s | %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			getStatusText(c.Writer.Status()),
			latency.String(),
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Proto,
			c.Request.UserAgent())
	}
}

// getStatusText converts status code to text representation
func getStatusText(statusCode int) string {
	switch statusCode {
	case 200:
		return "200 OK"
	case 201:
		return "201 Created"
	case 400:
		return "400 Bad Request"
	case 401:
		return "401 Unauthorized"
	case 403:
		return "403 Forbidden"
	case 404:
		return "404 Not Found"
	case 500:
		return "500 Internal Server Error"
	default:
		return fmt.Sprintf("%d", statusCode)
	}
}
