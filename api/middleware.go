
package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a custom type to create a key for the request context.
// Using a custom type prevents collisions with other context keys.
type contextKey string

// userContextKey is the key we will use to store the user's claims in the request context.
const userContextKey = contextKey("userClaims")

// jwtAuthMiddleware is the new middleware that validates a JWT from the Authorization header.
// It replaces the old static API key middleware.
func (s *Server) jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header from the request
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// The header should be in the format "Bearer <token>"
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
			return
		}

		tokenString := headerParts[1]
		claims := &Claims{} // This is the Claims struct we defined in auth_handlers.go

		// Parse the JWT string and validate the signature.
		// The key function (`func(token *jwt.Token)`) is called during parsing to provide the secret key for verification.
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				
				return nil, http.ErrAbortHandler
			}
			return []byte(s.config.JWTSecret), nil
		})

		// Check for parsing errors. This can happen if the token is malformed,
		// the signature is invalid, or the token has expired.
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				respondWithError(w, http.StatusUnauthorized, "Invalid token signature")
				return
			}
			// This handles other errors, including token expiration.
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Finally, check if the parsed token is valid.
		if !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// If the token is valid, we enrich the request's context with the user's claims.
		// Subsequent handlers can now access the UserID and Role.
		ctx := context.WithValue(r.Context(), userContextKey, claims)

		// The request is authorized. Pass it on to the next handler in the chain,
		// but with the new, enriched context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}