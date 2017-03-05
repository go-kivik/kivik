package logfile

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/logger"
)

// DateFormat is the date format used by CouchDB logs.
const DateFormat = time.RFC1123

// Logger is a file logger instance.
type Logger struct {
	mutex    sync.RWMutex
	filename string
	level    logger.LogLevel
	f        *os.File
}

var _ logger.LogWriter = &Logger{}
var _ driver.Logger = &Logger{}

var now = time.Now

// Init initializes the logger. It looks for the following configuration
// parameters:
//
//  - file: The file to which logs are written. (required)
//  - level: The minimum log level to log to the file. (default: info)
func (l *Logger) Init(conf map[string]string) error {
	if l.f != nil {
		l.f.Close()
		l.filename = ""
		l.level = 0
	}
	filename, ok := conf["file"]
	if !ok {
		return errors.New("log.file must be configured")
	}
	if err := l.setLevel(conf); err != nil {
		return err
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}
	l.filename = filename
	l.f = f
	return nil
}

func (l *Logger) setLevel(conf map[string]string) error {
	level, ok := conf["level"]
	if !ok {
		// Default to Info
		l.level = logger.LogLevelInfo
		return nil
	}
	l.level, ok = logger.StringToLogLevel(level)
	if !ok {
		return errors.Errorf("unknown loglevel '%s'", level)
	}
	return nil
}

// WriteLog writes a log to the opened log file.
func (l *Logger) WriteLog(level logger.LogLevel, message string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	_, err := fmt.Fprintf(l.f, "[%s] [%s] [--] %s\n", now().Format(DateFormat), level, message)
	return err
}

// Log reads the log file.
func (l *Logger) Log(buf []byte, offset int) (int, error) {
	l.mutex.Lock()
	l.f.Sync()
	l.mutex.Unlock()
	f, err := os.Open(l.filename)
	defer f.Close()
	if err != nil {
		return 0, err
	}
	st, err := f.Stat()
	if err != nil {
		return 0, err
	}
	if st.Size() > int64(len(buf)) {
		_, err = f.Seek(-int64(offset+len(buf)), os.SEEK_END)
		if err != nil {
			return 0, err
		}
	}
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return n, err
	}
	return n, nil
}
