
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"golang.org/x/crypto/bcrypt"
)

// Claims defines the structure of the data we'll store in the JWT.
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// hashPassword is a helper function to securely hash a password using bcrypt.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) 
	return string(bytes), err
}

// checkPasswordHash compares a plain-text password with its hashed version.
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// handleRegister handles the user registration process.
// Endpoint: POST /auth/register
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	hashedPassword, err := hashPassword(creds.Password)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	user := database.User{
		Username:     creds.Username,
		Email:        creds.Email,
		PasswordHash: hashedPassword,
		// Role defaults to 'user' as defined in the model
	}

	// Create the user in the database
	if result := s.db.Create(&user); result.Error != nil {
		respondWithJSON(w, http.StatusBadRequest, "Username or email may already be in use")
		return
	}

	
	w.WriteHeader(http.StatusCreated)
}

// handleLogin handles the user login process.
// It authenticates the user and returns a signed JWT.
// Endpoint: POST /auth/login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var user database.User
	// Find the user by email
	if result := s.db.Where("email = ?", creds.Email).First(&user); result.Error != nil {
		// Use a generic error message for security to prevent email enumeration
		respondWithJSON(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check if the provided password matches the stored hash
	if !checkPasswordHash(creds.Password, user.PasswordHash) {
		respondWithJSON(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// If credentials are valid, create a new JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Sign the token with our secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to create authentication token")
		return
	}

	// Return the token to the client
	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": tokenString,
	})
}