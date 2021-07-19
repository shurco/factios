package logger

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	globalLogger   *zerolog.Logger
	getLoggerMutex sync.Mutex
)

// GetLogger returns a configured logger instance
func GetLogger(module string) zerolog.Logger {
	if globalLogger == nil {
		getLoggerMutex.Lock()
		defer getLoggerMutex.Unlock()
		logger := zerolog.New(os.Stderr)
		globalLogger = &logger
	}
	if module != "" {
		return globalLogger.With().Str("module", module).Timestamp().Logger()
	}
	return globalLogger.With().Timestamp().Logger()
}
