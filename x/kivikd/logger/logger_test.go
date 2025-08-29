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
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"
)

func TestFields(t *testing.T) {
	now := time.Now()
	f := Fields{
		"exists": "exists",
		"dur":    time.Second,
		"time":   now,
		"int":    123,
	}
	t.Run("Exists", func(t *testing.T) {
		if !f.Exists("exists") {
			t.Errorf("Should exist")
		}
	})
	t.Run("NotExists", func(t *testing.T) {
		if f.Exists("notExists") {
			t.Errorf("Should not exist")
		}
	})
	t.Run("Get", func(t *testing.T) {
		v, ok := f.Get("exists").(string)
		if !ok {
			t.Errorf("Should have returned a string")
		}
		if v != "exists" {
			t.Errorf("Should have returned expected value")
		}
	})
	t.Run("GetString", func(t *testing.T) {
		if f.GetString("exists") != "exists" {
			t.Errorf("Should have returned expected value")
		}
	})
	t.Run("GetStringNotExist", func(t *testing.T) {
		if f.GetString("notExists") != "-" {
			t.Errorf("Should have returned placeholder value")
		}
	})
	t.Run("GetDuration", func(t *testing.T) {
		if f.GetDuration("dur") != time.Second {
			t.Errorf("Should have returned expected value")
		}
	})
	t.Run("GetDurationNotExist", func(t *testing.T) {
		if f.GetDuration("notExists") != time.Duration(0) {
			t.Errorf("Should have returned zero value")
		}
	})
	t.Run("GetTime", func(t *testing.T) {
		if !f.GetTime("time").Equal(now) {
			t.Errorf("Should have returned expected value")
		}
	})
	t.Run("GetTimeNotExist", func(t *testing.T) {
		if !f.GetTime("notExist").IsZero() {
			t.Errorf("Should have returned zero time")
		}
	})
	t.Run("GetInt", func(t *testing.T) {
		if f.GetInt("int") != 123 {
			t.Errorf("Should have returned expected value")
		}
	})
	t.Run("GetIntNotExist", func(t *testing.T) {
		if f.GetInt("notExist") != 0 {
			t.Errorf("Should have returned 0")
		}
	})
}

func TestLogger(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	type logTest struct {
		Name     string
		Func     func(RequestLogger)
		Expected string
	}
	tests := []logTest{
		{
			Name: "EmptyishRequest",
			Func: func(l RequestLogger) {
				r, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/foo", nil)
				r.RemoteAddr = "127.0.0.1:123"
				l.Log(r, 200, nil)
			},
			Expected: `127.0.0.1 - [0001-01-01 00:00:00Z] (0s) "GET /foo HTTP/1.1" 200 0 "" ""` + "\n",
		},
		{
			Name: "Fields",
			Func: func(l RequestLogger) {
				r, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/foo", nil)
				r.RemoteAddr = "127.0.0.1:123"
				r.Header.Add("User-Agent", "Bog's special browser version 1.23")
				r.Header.Add("Referer", "http://somesite.com/")
				l.Log(r, 200, Fields{
					FieldUsername:     "bob",
					FieldTimestamp:    ts,
					FieldElapsedTime:  time.Duration(1234),
					FieldResponseSize: 56789,
				})
			},
			Expected: `127.0.0.1 bob [2006-01-02 15:04:05+07:00] (1.234µs) "GET /foo HTTP/1.1" 200 56789 "http://somesite.com/" "Bog's special browser version 1.23"` + "\n",
		},
	}
	for _, test := range tests {
		func(test logTest) {
			t.Run(test.Name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				l := New(buf)
				test.Func(l)
				if d := testy.DiffText(test.Expected, buf.String()); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}
