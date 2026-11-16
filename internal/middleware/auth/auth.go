package auth

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const CookieName = "auth_token"
const UserIDKey = "user_id"

type Claims struct {
	jwt.StandardClaims
	UserID string `json:"user_id"`
}

func Authenticate(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(CookieName)
		if err == nil {
			userID, valid := verify(cookie.Value, secret)
			if valid && userID != "" {
				c.Set(UserIDKey, userID)
			}
		}

		c.Next()
	}
}

func verify(value, secret string) (string, bool) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", false
	}
	return claims.UserID, true
}
