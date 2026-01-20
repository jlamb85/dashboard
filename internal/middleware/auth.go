package middleware

import (
	"net/http"
	"strings"
)

// AuthMiddleware checks for the presence of an authorization token in the request headers.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !isValidToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// isValidToken validates the provided token.
func isValidToken(token string) bool {
	// Implement token validation logic here (e.g., check against a database or a predefined list)
	return strings.HasPrefix(token, "Bearer ")
}