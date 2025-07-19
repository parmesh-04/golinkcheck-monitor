// api/helpers.go

package api

import (
	"encoding/json"
	"net/http"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// respondWithError is a helper to send a standardized JSON error message.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON is a helper to marshal data to JSON and write the response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func parseAndValidate(r *http.Request, v interface{}) error {
	// Decode the request body into the provided struct 'v'.
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}

	// Run the validation on the populated struct.
	if err := validate.Struct(v); err != nil {
		return err
	}

	return nil
}