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

package cmd

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

func Test_ping_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("invalid URL on command line", cmdTest{
		args:   []string{"--debug", "ping", "http://localhost:1/foo/bar/%xxx"},
		status: errors.ErrUsage,
	})
	tests.Add("full url on command line", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{})

		return cmdTest{
			args: []string{"ping", s.URL},
		}
	})
	tests.Add("server only on command line", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		})

		return cmdTest{
			args: []string{"--config", "./testdata/localhost.yaml", "ping", s.URL},
		}
	})
	tests.Add("no server provided", cmdTest{
		args:   []string{"ping", "foo/bar"},
		status: errors.ErrUsage,
	})
	tests.Add("network error", cmdTest{
		args:   []string{"ping", "http://localhost:9999/"},
		status: errors.ErrUnavailable,
	})
	tests.Add("Couch 1.7, up", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusBadRequest,
			Header: http.Header{
				"Server": []string{"CouchDB/1.7.1"},
			},
		})

		return cmdTest{
			args: []string{"ping", s.URL},
		}
	})
	tests.Add("Couch 3.0, up", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
		})

		return cmdTest{
			args: []string{"ping", s.URL},
		}
	})
	tests.Add("Couch 3.0, down", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusNotFound,
		})

		return cmdTest{
			args:   []string{"ping", s.URL},
			status: errors.ErrNotFound,
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
