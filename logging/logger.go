

package logging

import (
	"log/slog"
	"os"
)

// InitLogger sets up our structured JSON logger.
func InitLogger() {
	// Create a new JSON handler that writes to standard output.

	handler := slog.NewJSONHandler(os.Stdout, nil)

	// Set this handler as the default logger for the whole application.
	slog.SetDefault(slog.New(handler))
}