package auth

import (
	"github.com/gin-gonic/gin"
	"strings"
	"url-shortener/internal/utils"
)

func Middleware(jwtService *utils.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token format"})
			return
		}

		userId, err := jwtService.Validate(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}


		c.Set("userId", userId)

		c.Next()
	}
}

