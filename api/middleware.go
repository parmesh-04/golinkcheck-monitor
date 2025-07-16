
package api

import (
	"net/http"
	"strings"
)

// authMiddleware is our bouncer. It checks for a valid API key.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the key from the request's Authorization header.
		//    The standard format is "Authorization: Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// 2. The header should be in the format "Bearer <key>". We split it to get the key.
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
			return
		}
		requestKey := parts[1]

		// 3. Compare the request key with the real secret key from our config.
		if requestKey != s.config.APISecretKey {
			respondWithError(w, http.StatusUnauthorized, "Invalid API Key")
			return
		}

		// 4. If the key is valid, call the next handler in the chain.
		//    This passes the request along to the actual route handler (e.g., handleCreateMonitor).
		next.ServeHTTP(w, r)
	})
}