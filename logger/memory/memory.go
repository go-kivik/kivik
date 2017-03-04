// Package memory provides a simple in-memory logger, intended for testing.
package memory

import (
	"container/ring"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/serve"
)

type log struct {
	level   serve.Level
	message string
}

// Logger is an in-memory logger instance. It fulfills both the serve.Logger
// and driver.Logger interfaces
type Logger struct {
	ring *ring.Ring
}

var _ serve.Logger = &Logger{}
var _ driver.Logger = &Logger{}

// New returns a new logger, with the specified capacity. Once more than cap
// lines have been logged, the oldest logs will be dropped.
func New(cap int) *Logger {
	if cap <= 0 {
		panic("cap must be > 0")
	}
	return &Logger{
		ring: ring.New(cap),
	}
}

// WriteLog logs the message at the designated level.
func (l *Logger) WriteLog(level serve.Level, message string) {
	l.ring.Value = log{
		level:   level,
		message: message,
	}
	l.ring = l.ring.Next()
}

// Log returns the requested log.
func (l *Logger) Log(buf []byte, offset int) (int, error) {
	return 0, nil
}
