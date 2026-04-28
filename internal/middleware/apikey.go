package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const RequestSKHeader = "X-Internal-API-Key"

type APIKey struct {
	sk map[string]struct{}
}

func NewAPIKey(sk map[string]struct{}) APIKey {
	return APIKey{
		sk: sk,
	}
}

func CurrentSK(r *http.Request) string {
	sk := strings.TrimSpace(r.Header.Get(RequestSKHeader))
	if sk != "" {
		return sk
	}

	sk = strings.TrimSpace(r.URL.Query().Get("auth"))
	if sk != "" {
		return sk
	}

	if token, ok := strings.CutPrefix(strings.TrimSpace(r.Header.Get("Authorization")), "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return ""
}

func (a *APIKey) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sk := CurrentSK(c.Request)

		if _, ok := a.sk[sk]; sk == "" || !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "sk not allowed"})
			c.Abort()
			return
		}
		c.Request.Header.Set(RequestSKHeader, sk)

		c.Next()
	}
}
