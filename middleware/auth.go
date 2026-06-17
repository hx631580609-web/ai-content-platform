package middleware

import (
	"net/http"
	"strings"

	"ai-content-platform/models"
	"ai-content-platform/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware authenticates the user using JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// If the prefix wasn't found, the format is wrong
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Store user ID and role in the context for later use
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminMiddleware checks if the user has admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserFromContext retrieves the user ID from the context
func GetUserFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	if id, ok := userID.(uint); ok {
		return id, true
	}

	return 0, false
}

// GetRoleFromContext retrieves the user role from the context
func GetRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}

	if r, ok := role.(string); ok {
		return r, true
	}

	return "", false
}

// HasPermission checks if the user has permission to access a resource
func HasPermission(c *gin.Context, resourceOwnerID uint) bool {
	userID, exists := GetUserFromContext(c)
	if !exists {
		return false
	}

	// Admins can access all resources
	role, _ := GetRoleFromContext(c)
	if role == "admin" {
		return true
	}

	// Regular users can only access their own resources
	return userID == resourceOwnerID
}

// AuthorizeResource checks if the user has permission to access a specific resource
func AuthorizeResource(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ownerID uint

		// Extract owner ID based on the model type
		switch v := model.(type) {
		case *models.User:
			ownerID = v.ID
		case *models.Content:
			ownerID = v.UserID
		case *models.BlogPost:
			ownerID = v.AuthorID
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unknown model type for authorization",
			})
			c.Abort()
			return
		}

		if !HasPermission(c, ownerID) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
