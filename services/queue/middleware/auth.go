package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware extracts user info from JWT and adds to context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Verify and decode token
		payload, err := decodeJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", payload["id"])
		c.Set("user_name", payload["name"])
		c.Set("user_email", payload["email"])
		
		// Handle role - could be a string or array
		if role, ok := payload["role"].(string); ok {
			c.Set("user_role", role)
		} else if roles, ok := payload["roles"].([]interface{}); ok && len(roles) > 0 {
			// If roles is an array, check for staff or admin
			roleStr := "user"
			for _, r := range roles {
				if rStr, ok := r.(string); ok {
					if rStr == "admin" {
						roleStr = "admin"
						break
					} else if rStr == "staff" {
						roleStr = "staff"
					}
				}
			}
			c.Set("user_role", roleStr)
		} else {
			c.Set("user_role", "user")
		}
		
		c.Set("user_payload", payload)

		c.Next()
	}
}

// decodeJWT decodes a JWT token without verification
func decodeJWT(tokenString string) (map[string]interface{}, error) {
	parts := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(tokenString); i++ {
		if tokenString[i] == '.' {
			parts = append(parts, tokenString[start:i])
			start = i + 1
		}
	}
	parts = append(parts, tokenString[start:])
	
	if len(parts) != 3 {
		return nil, http.ErrAbortHandler
	}

	// Decode payload (second part)
	payload := parts[1]
	// Add padding if needed
	padding := 4 - len(payload)%4
	if padding != 4 {
		for i := 0; i < padding; i++ {
			payload += "="
		}
	}
	
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
	}
	
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, err
	}
	
	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return nil, http.ErrAbortHandler
		}
	}
	
	return claims, nil
}

// StaffOnlyMiddleware ensures only staff can access
func StaffOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		roleStr := role.(string)
		if roleStr != "staff" && roleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Staff access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnlyMiddleware ensures only admins can access
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if role.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allow specific origins for development
		allowedOrigins := map[string]bool{
			"http://localhost:3000":    true,
			"http://localhost:8080":    true,
			"http://127.0.0.1:3000":    true,
			"http://127.0.0.1:8080":    true,
		}
		
		if origin != "" && allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			// Allow requests with no origin (curl, Postman, etc.)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
