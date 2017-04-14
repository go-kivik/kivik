package serve

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/flimzy/kivik"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		ip := r.RemoteAddr
		ip = ip[0:strings.LastIndex(ip, ":")]
		status := sw.status
		if status == 0 {
			status = kivik.StatusOK
		}
		fmt.Fprintf(os.Stderr, "%s - - %s %s %d\n", ip, r.Method, r.URL.String(), status)
	})
}
