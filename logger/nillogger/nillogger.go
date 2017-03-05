// Package nillogger provides a logger that ignores all log messages.
package nillogger

import logger "github.com/flimzy/kivik/logger"

// Logger is a nil logger. It discards all messages.
type Logger struct{}

var _ logger.LogWriter = &Logger{}

// Init does nothing
func (l *Logger) Init(_ map[string]string) error { return nil }

// WriteLog does nothing
func (l *Logger) WriteLog(_ logger.LogLevel, _ string) error { return nil }
