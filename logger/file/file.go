package file

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/serve"
)

// DateFormat is the date format used by CouchDB logs.
const DateFormat = time.RFC1123

// Logger is a file logger instance.
type Logger struct {
	mutex    sync.RWMutex
	filename string
	f        *os.File
}

var _ serve.LogWriter = &Logger{}
var _ driver.Logger = &Logger{}

var now = time.Now

// New opens a new file logger.
func New(filename string) (*Logger, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return &Logger{
		filename: filename,
		f:        f,
	}, nil
}

// WriteLog writes a log to the opened log file.
func (l *Logger) WriteLog(level serve.LogLevel, message string) error {
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
	if err != nil {
		return 0, err
	}
	st, err := f.Stat()
	if err != nil {
		return 0, err
	}
	if st.Size() > int64(len(buf)) {
		var x int64
		x, err = f.Seek(-int64(offset+len(buf)), os.SEEK_END)
		if err != nil {
			return 0, err
		}
		fmt.Printf("x = %d\n", x)
	}
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return n, err
	}
	return n, nil
}
