// checker/checker.go

package checker

import (
	"net/http"
	"time"

	"github.com/parmesh-04/golinkcheck-monitor/database"
)

// Check performs a single HTTP GET request to the given URL with a specific timeout.
// It returns a CheckResult containing the outcome.
func Check(url string, timeout time.Duration) database.CheckResult {
	// Create a custom HTTP client with the specified timeout.
	// This is crucial to prevent a check from hanging indefinitely on a slow server.
	client := &http.Client{
		Timeout: timeout,
	}

	// Start a timer to measure the request duration.
	startTime := time.Now()

	// Use our custom client to perform the GET request.
	resp, err := client.Get(url)

	// Calculate the time elapsed since we started.
	duration := time.Since(startTime)

	result := database.CheckResult{
		CheckedAt:      time.Now(),
		DurationMs: duration.Milliseconds(),
	}

	// If an error occurred (like a timeout), we record it and return.
	if err != nil {
		result.ErrorMessage = err.Error()
		result.StatusCode = 0 // No status code was received.
		return result
	}

	// We must close the response body to free up resources.
	// 'defer' ensures this runs right before the function returns.
	defer resp.Body.Close()

	// The request was successful, so we record the status code.
	result.StatusCode = resp.StatusCode

	return result
}