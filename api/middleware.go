package api

import (
	"net/http"
	"strings"
)

// authMiddleware checks for a valid API key in the Authorization header
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// checks if the header is in the format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
			return
		}
		requestKey := parts[1]

		if requestKey != s.config.APISecretKey {
			respondWithError(w, http.StatusUnauthorized, "Invalid API Key")
			return
		}

		next.ServeHTTP(w, r)
	})
}