package serve

import "github.com/flimzy/kivik/driver"

// LogLevel is a log level
type LogLevel int

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warning"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}

// The log levels specified by CouchDB.
// See http://docs.couchdb.org/en/2.0.0/config/logging.html
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogWriter is an interface for a logging backend.
type LogWriter interface {
	WriteLog(level LogLevel, message string) error
}

// LoggingClient bundles a driver.Client, driver.Logger, and a LogWriter, to
// both log requests, and serve logs.
type LoggingClient struct {
	driver.Client
	driver.Logger
	LogWriter
}

func (s *Service) log(level LogLevel, msg string) {
	if s.Logger == nil {
		return
	}
	s.Logger.WriteLog(level, msg)
}

// Debug logs a debug message to the registered logger.
func (s *Service) Debug(msg string) {
	s.log(LogLevelDebug, msg)
}

// Info logs an informational message to the registered logger.
func (s *Service) Info(msg string) {
	s.log(LogLevelInfo, msg)
}

// Warn logs a warning message to the registered logger.
func (s *Service) Warn(msg string) {
	s.log(LogLevelWarn, msg)
}

// Error logs an error message to the registered logger.
func (s *Service) Error(msg string) {
	s.log(LogLevelError, msg)
}
