package serve

import (
	"fmt"
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

const (
	// DefaultLogBytes is the default number of log bytes to return.
	// See http://docs.couchdb.org/en/2.0.0/api/server/common.html#log
	DefaultLogBytes = 1000

	// DefaultLogOffset is the offset from the end of the log, in bytes.
	DefaultLogOffset = 0
)

// Level is a log level
type Level int

// The log levels specified by CouchDB.
// See http://docs.couchdb.org/en/2.0.0/config/logging.html
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger is an interface for a logging backend.
type Logger interface {
	WriteLog(level Level, message string)
}

func log(w http.ResponseWriter, r *http.Request) error {
	logger, ok := getClient(r).(driver.Logger)
	if !ok {
		return kivik.ErrNotImplemented
	}

	w.Header().Set("Content-Type", typeText)
	length, ok, err := intParam(r, "bytes")
	if err != nil {
		fmt.Printf("err1:%s", err)
		return err
	}
	if !ok {
		length = DefaultLogBytes
	}
	offset, ok, err := intParam(r, "offset")
	if err != nil {
		fmt.Printf("err2:%s", err)
		return err
	}
	if !ok {
		offset = DefaultLogOffset
	}

	buf := make([]byte, length)
	if _, err = logger.Log(buf, offset); err != nil {
		fmt.Printf("err3:%s", err)
		return err
	}
	w.Write(buf)
	return nil
}
