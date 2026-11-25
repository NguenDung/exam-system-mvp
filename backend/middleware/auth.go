package middleware

import (
	"net/http"
	"strings"

	"exam-system-mvp/apiresponse"
	"exam-system-mvp/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims is an alias kept for backward compatibility with existing imports.
type Claims = utils.Claims

// AuthRequired validates JWT and injects user info into context.
func AuthRequired() gin.HandlerFunc {
	return AuthRequiredWithSecret(utils.GetJWTSecret())
}

func AuthRequiredWithSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apiresponse.Error(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			apiresponse.Error(c, http.StatusUnauthorized, "Invalid Authorization header format")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			apiresponse.Error(c, http.StatusUnauthorized, "Invalid or expired token", err)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole returns a middleware that allows only the specified roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		if _, ok := allowed[roleStr]; !ok {
			apiresponse.Error(c, http.StatusForbidden, "Access denied: insufficient role")
			c.Abort()
			return
		}
		c.Next()
	}
}
