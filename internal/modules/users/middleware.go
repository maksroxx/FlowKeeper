package users

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/maksroxx/flowkeeper/internal/config"
)

func AuthMiddleware(cfg config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid auth header format"})
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userID := claims["sub"]
		role := claims["role"]
		permsRaw := claims["permissions"]

		c.Set("userID", userID)
		c.Set("userRole", role)

		var perms []string
		if slice, ok := permsRaw.([]interface{}); ok {
			for _, item := range slice {
				if s, ok := item.(string); ok {
					perms = append(perms, s)
				}
			}
		}
		c.Set("permissions", perms)

		c.Next()
	}
}

func RequirePermission(requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")

		if role == "admin" {
			c.Next()
			return
		}

		permsInterface, exists := c.Get("permissions")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: no permissions found"})
			return
		}

		perms, ok := permsInterface.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal permission error"})
			return
		}

		for _, p := range perms {
			if p == requiredPerm {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Missing permission: %s", requiredPerm)})
	}
}
