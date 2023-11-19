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

package logger

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// RequestLogger is a request logger.
type RequestLogger interface {
	Log(req *http.Request, status int, fields Fields)
}

// Pre-defined log fields
const (
	FieldUsername     = "username"
	FieldTimestamp    = "timestamp"
	FieldElapsedTime  = "elapsed"
	FieldResponseSize = "size"
)

// Fields is simple wrapper around logging fields.
type Fields map[string]interface{}

// Exists returns true if the requested key exists in the map.
func (f Fields) Exists(key string) bool {
	_, ok := f[key]
	return ok
}

// Get returns the value associated with a key.
func (f Fields) Get(key string) interface{} {
	return f[key]
}

// GetString returns a value as a string, or "-"
func (f Fields) GetString(key string) string {
	v, ok := f[key].(string)
	if ok {
		return v
	}
	return "-"
}

// GetDuration returns a value as a time.Duration
func (f Fields) GetDuration(key string) time.Duration {
	v, _ := f[key].(time.Duration)
	return v
}

// GetTime returns a value as a timestamp.
func (f Fields) GetTime(key string) time.Time {
	v, _ := f[key].(time.Time)
	return v
}

// GetInt returns a value as an int.
func (f Fields) GetInt(key string) int {
	v, _ := f[key].(int)
	return v
}

type logger struct {
	w io.Writer
}

var _ RequestLogger = &logger{}

// New returns a new RequestLogger that writes apache-style logs to an io.Writer.
func New(w io.Writer) RequestLogger {
	return &logger{w}
}

// DefaultLogger logs to stderr.
var DefaultLogger = New(os.Stderr)

func (l *logger) Log(req *http.Request, status int, fields Fields) {
	fmt.Fprintf(l.w, `%s %s [%s] (%s) "%s %s %s" %d %d "%s" "%s"%c`,
		req.RemoteAddr[0:strings.LastIndex(req.RemoteAddr, ":")],
		fields.GetString(FieldUsername),
		fields.GetTime(FieldTimestamp).Format("2006-01-02 15:04:05Z07:00"),
		fields.GetDuration(FieldElapsedTime).String(),
		req.Method,
		req.URL.String(),
		req.Proto,
		status,
		fields.GetInt(FieldResponseSize),
		req.Header.Get("Referer"),
		req.Header.Get("User-Agent"),
		'\n',
	)
}
