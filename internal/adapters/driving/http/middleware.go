package http

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// apiKeyAuth rejects any request whose X-Api-Key header does not match the
// configured shared key. The comparison is constant-time.
func apiKeyAuth(want string) gin.HandlerFunc {
	wantBytes := []byte(want)
	return func(c *gin.Context) {
		got := []byte(c.GetHeader("X-Api-Key"))
		if subtle.ConstantTimeCompare(got, wantBytes) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
