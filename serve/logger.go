package serve

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flimzy/kivik/driver"
)

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
	// Write log should write the passed message to the logging backend.
	// The message is guaranteed not to end with any trailing newline or spaces.
	WriteLog(level LogLevel, message string) error
}

// LoggingClient bundles a driver.Client, driver.Logger, and a LogWriter, to
// both log requests, and serve logs.
type LoggingClient struct {
	driver.Client
	driver.Logger
	LogWriter
}

func (s *Service) log(level LogLevel, format string, args ...interface{}) {
	if s.LogWriter == nil {
		return
	}
	msg := strings.TrimSpace(fmt.Sprintf(format, args...))
	s.LogWriter.WriteLog(level, msg)
}

// Debug logs a debug message to the registered logger.
func (s *Service) Debug(format string, args ...interface{}) {
	s.log(LogLevelDebug, format, args...)
}

// Info logs an informational message to the registered logger.
func (s *Service) Info(format string, args ...interface{}) {
	s.log(LogLevelInfo, format, args...)
}

// Warn logs a warning message to the registered logger.
func (s *Service) Warn(format string, args ...interface{}) {
	s.log(LogLevelWarn, format, args...)
}

// Error logs an error message to the registered logger.
func (s *Service) Error(format string, args ...interface{}) {
	s.log(LogLevelError, format, args...)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func requestLogger(s *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		ip := r.RemoteAddr
		ip = ip[0:strings.LastIndex(ip, ":")]
		s.Info("%s - - %s %s %d", ip, r.Method, r.URL.String(), sw.status)
	})
}
