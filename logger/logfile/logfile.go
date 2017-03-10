package logfile

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/logger"
)

// Logger is a file logger instance.
type Logger struct {
	mutex    sync.RWMutex
	filename string
	f        *os.File
}

var _ logger.LogWriter = &Logger{}
var _ driver.LogReader = &Logger{}

var now = time.Now

// Init initializes the logger. It looks for the following configuration
// parameters:
//
//  - file: The file to which logs are written. (required)
func (l *Logger) Init(conf map[string]string) error {
	if l.f != nil {
		l.f.Close()
		l.filename = ""
	}
	filename, ok := conf["file"]
	if !ok {
		return errors.New("log.file must be configured")
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}
	l.filename = filename
	l.f = f
	return nil
}

// WriteLog writes a log to the opened log file.
func (l *Logger) WriteLog(level logger.LogLevel, message string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	_, err := fmt.Fprintf(l.f, "[%s] [%s] [--] %s\n", now().Format(logger.TimeFormat), level, message)
	return err
}

type logReadCloser struct {
	io.Reader
	io.Closer
}

// Log reads the log file.
func (l *Logger) Log(length, offset int64) (io.ReadCloser, error) {
	if length < 0 {
		return nil, errors.Status(kivik.StatusBadRequest, "invalid length specified")
	}
	if offset < 0 {
		return nil, errors.Status(kivik.StatusBadRequest, "invalid offset specified")
	}
	if length == 0 {
		return ioutil.NopCloser(&bytes.Buffer{}), nil
	}
	l.f.Sync()
	f, err := os.Open(l.filename)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() > length {
		_, err := f.Seek(-(offset + length), os.SEEK_END)
		if err != nil {
			return nil, err
		}
	}
	return &logReadCloser{
		Reader: &io.LimitedReader{R: f, N: length},
		Closer: f,
	}, nil
}
