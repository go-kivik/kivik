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

	fmt.Printf("length = %d, offset = %d\n", length, offset)
	buf := make([]byte, length)
	n, err := logger.Log(buf, offset)
	if err != nil {
		return err
	}
	w.Write(buf[0:n])
	return nil
}
