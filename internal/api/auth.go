package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secretKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			GatewayRequestsTotal.WithLabelValues("unauthorized").Inc()
			http.Error(w, "Unauthorized: Missing or invalid token", http.StatusUnauthorized)
			return
		}

		// 2. Parse the token string
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. Verify the token signature mathematically
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			log.Printf("[AUTH FAILED] Invalid token from IP: %s", r.RemoteAddr)
			GatewayRequestsTotal.WithLabelValues("unauthorized").Inc()
			http.Error(w, "Unauthorized: Invalid token signature", http.StatusUnauthorized)
			return
		}

		// 4. Token is valid! Pass the request to the next middleware (the rate limiter)
		next.ServeHTTP(w, r)
	}
}
