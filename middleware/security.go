package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-content-platform/database"
	"ai-content-platform/models"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Strict Transport Security
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';")
		
		c.Next()
	}
}

// RateLimit implements basic rate limiting
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	type RequestInfo struct {
		Count    int
		FirstReq time.Time
	}
	
	requests := make(map[string]*RequestInfo)
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		now := time.Now()
		
		if reqInfo, exists := requests[ip]; exists {
			if now.Sub(reqInfo.FirstReq) < window {
				if reqInfo.Count >= maxRequests {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "Rate limit exceeded",
					})
					c.Abort()
					return
				}
				requests[ip] = &RequestInfo{
					Count:    reqInfo.Count + 1,
					FirstReq: reqInfo.FirstReq,
				}
			} else {
				requests[ip] = &RequestInfo{
					Count:    1,
					FirstReq: now,
				}
			}
		} else {
			requests[ip] = &RequestInfo{
				Count:    1,
				FirstReq: now,
			}
		}
		
		c.Next()
	}
}

// InputValidation validates common input patterns to prevent injection attacks
func InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for potential SQL injection patterns in query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSQLInjection(value) {
					logSecurityEvent(c, "SQL Injection Attempt", fmt.Sprintf("Parameter: %s, Value: %s", key, value))
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid input detected",
					})
					c.Abort()
					return
				}
				
				if containsXSS(value) {
					logSecurityEvent(c, "XSS Attempt", fmt.Sprintf("Parameter: %s, Value: %s", key, value))
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid input detected",
					})
					c.Abort()
					return
				}
			}
		}
		
		// For POST/PUT requests, check form data and JSON body
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			// Check form data
			for key, values := range c.Request.PostForm {
				for _, value := range values {
					if containsSQLInjection(value) {
						logSecurityEvent(c, "SQL Injection Attempt", fmt.Sprintf("Form Parameter: %s, Value: %s", key, value))
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Invalid input detected",
						})
						c.Abort()
						return
					}
					
					if containsXSS(value) {
						logSecurityEvent(c, "XSS Attempt", fmt.Sprintf("Form Parameter: %s, Value: %s", key, value))
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Invalid input detected",
						})
						c.Abort()
						return
					}
				}
			}
		}
		
		c.Next()
	}
}

// containsSQLInjection checks if a string contains potential SQL injection patterns
func containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		"' OR 1=1", "--", "/*", "*/", "@@", "CHAR(", "N'", "EXEC", "SELECT", "INSERT", "UPDATE", "DELETE",
		"DROP", "CREATE", "ALTER", "UNION", "TABLE", "DATABASE", "FROM", "WHERE", "HAVING", "ORDER BY",
	}
	
	inputUpper := strings.ToUpper(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(inputUpper, pattern) {
			return true
		}
	}
	
	return false
}

// containsXSS checks if a string contains potential XSS patterns
func containsXSS(input string) bool {
	xssPatterns := []string{
		"<script", "javascript:", "onerror=", "onload=", "onclick=", "onmouseover=", "onfocus=",
		"document.cookie", "window.location", "<iframe", "<object", "<embed", "eval(", "expression(",
	}
	
	inputLower := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(inputLower, pattern) {
			return true
		}
	}
	
	return false
}

// logSecurityEvent logs security-related events to the database
func logSecurityEvent(c *gin.Context, eventType, description string) {
	logEntry := models.SystemLog{
		UserID:      0, // Anonymous for security events
		Action:      eventType,
		ObjectType:  "security_event",
		ObjectID:    0,
		Description: description,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}
	
	// Save log entry to database
	database.DB.Create(&logEntry)
}

// SanitizeInput sanitizes user inputs to prevent injection attacks
func SanitizeInput(input string) string {
	// Remove potential harmful characters/patterns
	// This is a basic implementation - in production, use a proper sanitization library
	sanitized := strings.ReplaceAll(input, "'", "")
	sanitized = strings.ReplaceAll(sanitized, "\"", "")
	sanitized = strings.ReplaceAll(sanitized, ";", "")
	sanitized = strings.ReplaceAll(sanitized, "--", "")
	sanitized = strings.ReplaceAll(sanitized, "/*", "")
	sanitized = strings.ReplaceAll(sanitized, "*/", "")
	
	return sanitized
}

// SanitizeMiddleware sanitizes input parameters
func SanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = SanitizeInput(value)
			}
		}
		
		c.Next()
	}
}