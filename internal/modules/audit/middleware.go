package audit

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func AutomaticAuditMiddleware(auditService Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {

			status := c.Writer.Status()
			if status >= 200 && status < 300 {

				userIDVal, exists := c.Get("userID")
				var userID uint
				if exists {
					switch v := userIDVal.(type) {
					case uint:
						userID = v
					case int:
						userID = uint(v)
					case float64:
						userID = uint(v)
					}
				}

				path := c.Request.URL.Path
				clientIP := c.ClientIP()

				details := fmt.Sprintf("Auto Log: %s %s (Status: %d)", method, path, status)

				if userID > 0 {
					auditService.Log(userID, method, path, 0, details, clientIP)
				}
			}
		}
	}
}
