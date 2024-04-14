package middleware

import (
	"encoding/base64"
	"github.com/golang-jwt/jwt"
	"net/http"
)

func TokenAuth(secretKey string, forAdminOnly bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("token")
			decodedSecret, _ := base64.StdEncoding.DecodeString(secretKey)

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return decodedSecret, nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if forAdminOnly {
				role, _ := claims["role"]
				if role != "admin" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)

		})
	}
}
