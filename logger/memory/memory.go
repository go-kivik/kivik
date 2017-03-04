// Package memory provides a simple in-memory logger, intended for testing.
package memory

import (
	"container/ring"
	"fmt"
	"time"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/serve"
)

// DateFormat is the date format used by CouchDB logs.
const DateFormat = time.RFC1123

type log struct {
	time    time.Time
	level   serve.LogLevel
	message string
}

var now = time.Now

func (l log) String() string {
	return fmt.Sprintf("[%s] [%s] [--] %s\n", l.time.Format(DateFormat), l.level, l.message)
}

// Logger is an in-memory logger instance. It fulfills both the serve.Logger
// and driver.Logger interfaces
type Logger struct {
	ring *ring.Ring
}

var _ serve.LogWriter = &Logger{}
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
func (l *Logger) WriteLog(level serve.LogLevel, message string) error {
	l.ring.Value = log{
		time:    now(),
		level:   level,
		message: message,
	}
	l.ring = l.ring.Next()
	return nil
}

// Log returns the requested log.
func (l *Logger) Log(buf []byte, offset int) (int, error) {
	cur := l.ring.Prev()
	var i int
	max := len(buf)
	for i = max; i > 0; {
		if cur.Value == nil {
			// We reached the end of the log
			break
		}
		msg := cur.Value.(log).String()
		if i-len(msg) <= 0 {
			copy(buf[:i], msg[len(msg)-i:])
			i = 0
			break
		}
		copy(buf[i-len(msg):i], msg)
		i = i - len(msg)
		if cur == l.ring {
			// We did a full circle
			break
		}
		cur = cur.Prev()
	}
	if i == 0 {
		return len(buf), nil
	}
	// This means there were fewer logs than requested, so we need to
	// shift the buffer to the beginning.
	len := max - i
	copy(buf, buf[i:])
	return len, nil
}
