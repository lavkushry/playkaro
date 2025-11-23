package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		var isAdmin bool
		err := db.DB.QueryRow("SELECT is_admin FROM users WHERE id=$1", userID).Scan(&isAdmin)
		if err != nil || !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
