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

package couchserver

import (
	"context"
	"net/http/httptest"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/x/kivikd/auth"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
)

type testKey struct {
	string
}

func TestGetSession(t *testing.T) {
	key := &testKey{"key"}
	h := Handler{
		SessionKey: key,
	}
	req := httptest.NewRequest("GET", "/_session", nil)
	session := &auth.Session{
		AuthMethod: "magic",
		AuthDB:     "_users",
		User: &authdb.UserContext{
			Name: "bob",
		},
	}
	req = req.WithContext(context.WithValue(req.Context(), key, &session))
	w := httptest.NewRecorder()
	handler := h.GetSession()
	handler(w, req)
	expected := map[string]interface{}{
		"info": map[string]interface{}{
			"authenticated":           "magic",
			"authentication_db":       "_users",
			"authentication_handlers": nil,
		},
		"ok": true,
		"userCtx": map[string]interface{}{
			"name":  "bob",
			"roles": []string{},
		},
	}
	resp := w.Result()
	defer resp.Body.Close()
	if d := testy.DiffAsJSON(expected, resp.Body); d != nil {
		t.Error(d)
	}
}
