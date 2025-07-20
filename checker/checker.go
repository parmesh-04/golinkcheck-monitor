package checker

import (
	"net/http"
	"time"
	"github.com/parmesh-04/golinkcheck-monitor/database"
)

// Check performs a health check on the given URL and returns the result
func Check(url string, timeout time.Duration) database.CheckResult {
	client := &http.Client{
		Timeout: timeout,
	}

	startTime := time.Now()

	resp, err := client.Get(url)

	duration := time.Since(startTime)

	result := database.CheckResult{
		CheckedAt:  time.Now(),
		DurationMs: duration.Milliseconds(),
	}

	if err != nil {
		result.ErrorMessage = err.Error()
		result.StatusCode = 0
		return result
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	return result
}