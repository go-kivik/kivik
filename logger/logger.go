package logger

import (
	"net/http"
	"strings"
	"time"
)

// TimeFormat is the default time format used for CouchDB logs.
const TimeFormat = time.RFC1123

// LogLevel is a log level
type LogLevel int

// The log levels specified by CouchDB.
// See http://docs.couchdb.org/en/2.0.0/config/logging.html
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

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

// StringToLogLevel converts a string to the associated LogLevel. ok will be
// false if the log level is unknown.
func StringToLogLevel(str string) (level LogLevel, ok bool) {
	switch strings.ToLower(str) {
	case "debug":
		return LogLevelDebug, true
	case "info":
		return LogLevelInfo, true
	case "warn", "warning":
		return LogLevelWarn, true
	case "error":
		return LogLevelError, true
	default:
		return 0, false
	}
}

// LogWriter is an interface for a logging backend.
type LogWriter interface {
	// Init is used to (re)start the logger. When called, any log files should
	// be closed (if previously opened), and re-opened according to the passed
	// configuration. Configuration keys and values are backend-specific. Any
	// unrecognized configuration values should be ignored.
	Init(config map[string]string) error
	// Write log should write the passed message to the logging backend.
	// The message is guaranteed not to end with any trailing newline or spaces.
	WriteLog(level LogLevel, message string) error
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
