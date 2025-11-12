package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// needed to ensure we have the id for tracking every request for its lifetime
func CorrelationID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		correlationId := ctx.GetHeader("X-Correlation-ID")
		if correlationId == "" {
			correlationId = uuid.New().String()
		}
		ctx.Set("X-Correlation-ID", correlationId)
		ctx.Header("X-Correlation-ID", correlationId)
		ctx.Next()
	}
}
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authKey := c.GetHeader("Authorization")
		if authKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}
		parts := strings.Split(authKey, "")
		if len(parts) != 2 || parts[0] == "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid Api Key",
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("my-secret-key"), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid Token",
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
		}
		c.Next()

	}
}
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
