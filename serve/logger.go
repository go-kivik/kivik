package serve

import (
	"net/http"
	"time"

	"github.com/flimzy/kivik/serve/logger"
)

type statusWriter struct {
	http.ResponseWriter
	status    int
	byteCount int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.byteCount += n
	return n, err
}

func loggerMiddleware(rlog logger.RequestLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)
			session := MustGetSession(r.Context())
			var username string
			if session.User != nil {
				username = session.User.Name
			}
			fields := logger.Fields{
				logger.FieldUsername:     username,
				logger.FieldTimestamp:    start,
				logger.FieldElapsedTime:  time.Now().Sub(start),
				logger.FieldResponseSize: sw.byteCount,
			}
			rlog.Log(r, sw.status, fields)
		})
	}
}
