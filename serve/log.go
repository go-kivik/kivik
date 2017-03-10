package serve

import (
	"net/http"

	"github.com/flimzy/kivik/errors"
)

const (
	// DefaultLogBytes is the default number of log bytes to return.
	// See http://docs.couchdb.org/en/2.0.0/api/server/common.html#log
	DefaultLogBytes = 1000

	// DefaultLogOffset is the offset from the end of the log, in bytes.
	DefaultLogOffset = 0
)

func log(w http.ResponseWriter, r *http.Request) error {
	client := getClient(r)

	length, ok, err := intQueryParam(r, "bytes")
	if err != nil {
		return err
	}
	if !ok {
		length = DefaultLogBytes
	}
	if length < 0 {
		return errors.Status(http.StatusBadRequest, "bytes must be a positive integer")
	}
	offset, ok, err := intQueryParam(r, "offset")
	if err != nil {
		return err
	}
	if !ok {
		offset = DefaultLogOffset
	}
	if offset < 0 {
		return errors.Status(http.StatusBadRequest, "offset must be a positive integer")
	}

	logR, err := client.LogContext(context.Background(), length, offset)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", typeText)
	_, err = io.Copy(w, logR)
	return err
}
