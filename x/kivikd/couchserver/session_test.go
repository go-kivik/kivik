package couchserver

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-kivik/kivikd/v4/auth"
	"github.com/go-kivik/kivikd/v4/authdb"
	"gitlab.com/flimzy/testy"
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
