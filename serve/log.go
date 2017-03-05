package serve

import "net/http"

const (
	// DefaultLogBytes is the default number of log bytes to return.
	// See http://docs.couchdb.org/en/2.0.0/api/server/common.html#log
	DefaultLogBytes = 1000

	// DefaultLogOffset is the offset from the end of the log, in bytes.
	DefaultLogOffset = 0
)

func log(w http.ResponseWriter, r *http.Request) error {
	client := getClient(r)

	length, ok, err := intParam(r, "bytes")
	if err != nil {
		return err
	}
	if !ok {
		length = DefaultLogBytes
	}
	offset, ok, err := intParam(r, "offset")
	if err != nil {
		return err
	}
	if !ok {
		offset = DefaultLogOffset
	}

	buf := make([]byte, length)
	n, err := client.Log(buf, offset)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", typeText)
	w.Write(buf[0:n])
	return nil
}
