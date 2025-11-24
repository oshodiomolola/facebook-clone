package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	// Ensure this import path matches your go.mod module name:
	"facebookapi/helpers"
)

// Authenticate is a middleware to validate JWT tokens from the Authorization header.
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			// Use AbortWithStatusJSON for cleaner error handling
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Check for the required "Bearer " prefix
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Expected 'Bearer <token>'"})
			return
		}

		// Extract the token string
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))

		// Validate the token using your helper function
		// NOTE: 'ValidateToken' MUST start with an uppercase 'V' in your helpers package file
		claims, err := helpers.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Token validation error: %v for token: %s", err, tokenString)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Store the claims in the Gin context for subsequent handlers (controllers)
		// Access it later in a handler using: claims := c.MustGet("claims").(*YourClaimsStruct)
		c.Set("claims", claims)

		// Continue down the chain to the actual route handler
		c.Next()
	}
}
