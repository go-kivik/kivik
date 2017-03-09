// Package memlogger provides a simple in-memory logger, intended for testing.
package memlogger

import (
	"container/ring"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/logger"
	"github.com/pkg/errors"
)

var now = time.Now

// Logger is an in-memory logger instance. It fulfills both the logger.Logger
// and driver.Logger interfaces
type Logger struct {
	ring *ring.Ring
}

var _ logger.LogWriter = &Logger{}
var _ driver.LogReader = &Logger{}

// Init initializes the memory logger. It considers the following configuration
// parameters:
//
// - capacity: The number of log entries to keep in memory. Defaults to 100.
//  - level: The minimum log level to log to the file. (default: info)
func (l *Logger) Init(conf map[string]string) error {
	l.ring = nil
	cap, err := getCapacity(conf)
	if err != nil {
		return err
	}
	l.ring = ring.New(cap)
	return nil
}

func getCapacity(conf map[string]string) (int, error) {
	cap, ok := conf["capacity"]
	if !ok {
		return 100, nil
	}
	c, err := strconv.Atoi(cap)
	if err != nil {
		return 0, errors.Wrapf(err, "invalid capacity '%s'", cap)
	}
	return c, nil
}

// WriteLog logs the message at the designated level.
func (l *Logger) WriteLog(level logger.LogLevel, message string) error {
	msg := fmt.Sprintf("[%s] [%s] [--] %s\n", now().Format(logger.TimeFormat), level, message)
	l.ring.Value = &msg
	l.ring = l.ring.Next()
	return nil
}

type memLogReader struct {
	list []*string
}

func (r *memLogReader) Read(p []byte) (n int, err error) {
	if len(r.list) == 0 {
		return 0, io.EOF
	}
	msg := *r.list[len(r.list)-1]
	n = copy(p, msg)
	if n < len(msg) {
		// We didn't send the whole message, so lets save the remainder
		remain := msg[n+1:]
		r.list[len(r.list)-1] = &remain
		return
	}
	r.list = r.list[:len(r.list)-1]
	return
}

// Log returns the requested log.
func (l *Logger) Log(length, offset int64) (io.ReadCloser, error) {
	cur := l.ring.Prev()
	list := make([]*string, 0, length/100)
	remain := length
	for {
		if cur.Value == nil {
			// We reached the end of the log
			break
		}
		msg := cur.Value.(*string)
		msgLen := int64(len(*msg))
		if offset > 0 {
			offset -= msgLen
			if offset < 0 {
				remain += offset
			}
		} else {
			remain -= msgLen
		}
		if offset <= 0 {
			list = append(list, msg)
		}
		if remain < 0 {
			msg := *list[len(list)-1]
			msg = msg[-remain:]
			list[len(list)-1] = &msg
		}
		if remain <= 0 {
			break
		}
		if cur == l.ring {
			// We did a full circle
			break
		}
		cur = cur.Prev()
	}
	return ioutil.NopCloser(&io.LimitedReader{
		R: &memLogReader{list: list},
		N: length,
	}), nil
}
