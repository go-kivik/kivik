// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build !js
// +build !js

package kivikd

import (
	"net/http"
	"time"

	"github.com/go-kivik/kivik/v4/x/kivikd/logger"
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
				logger.FieldElapsedTime:  time.Since(start),
				logger.FieldResponseSize: sw.byteCount,
			}
			rlog.Log(r, sw.status, fields)
		})
	}
}
