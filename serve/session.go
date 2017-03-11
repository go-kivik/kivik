package serve

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

// DefaultInsecureSecret is the hash secret used if couch_httpd_auth.secret
// is unconfigured. Please configure couch_httpd_auth.secret, or they're all
// gonna laugh at you!
const DefaultInsecureSecret = "They're all gonna laugh at you!"

// DefaultSessionTimeout is the default session timeout, in seconds, used if
// couch_httpd_auth.timeout is inuset.
const DefaultSessionTimeout = 600

func getSession(w http.ResponseWriter, r *http.Request) error {
	session := MustGetSession(r.Context())
	w.Header().Add("Content-Type", typeJSON)
	return json.NewEncoder(w).Encode(session)
}

func (s *Service) getAuthSecret(ctx context.Context) (string, error) {
	secret, err := s.Config().GetContext(ctx, "couch_httpd_auth", "secret")
	if errors.StatusCode(err) == kivik.StatusNotFound {
		return DefaultInsecureSecret, nil
	}
	if err != nil {
		return "", err
	}
	return secret, nil
}

func (s *Service) getSessionTimeout(ctx context.Context) (int, error) {
	timeout, err := s.Config().GetContext(ctx, "couch_httpd_auth", "timeout")
	if errors.StatusCode(err) == kivik.StatusNotFound {
		return DefaultSessionTimeout, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(timeout)
}

func postSession(w http.ResponseWriter, r *http.Request) error {
	authData := struct {
		Name     *string `form:"name" json:"name"`
		Password *string `form:"password" json:"password"`
	}{}
	if err := BindParams(r, &authData); err != nil {
		return errors.Status(kivik.StatusBadRequest, "unable to parse request data")
	}
	fmt.Printf("Got %v\n", authData)
	if authData.Name == nil || authData.Password == nil {
		return errors.Status(kivik.StatusUnauthorized, "Name or password is incorrect.")
	}
	s := getService(r)
	user, err := s.UserStore.Validate(r.Context(), *authData.Name, *authData.Password)
	if err != nil {
		return err
	}
	secret, err := s.getAuthSecret(r.Context())
	if err != nil {
		return err
	}
	timeout, err := s.getSessionTimeout(r.Context())
	if err != nil {
		return err
	}

	// Success, so create a cookie
	token := createAuthToken(*authData.Name, secret, user.Salt, time.Now().Unix())
	w.Header().Set("Cache-Control", "must-revalidate")
	http.SetCookie(w, &http.Cookie{
		Name:     "AuthSession",
		Value:    token,
		Path:     "/",
		MaxAge:   timeout,
		HttpOnly: true,
	})
	spew.Dump(token)
	return nil
}

func createAuthToken(name, secret, salt string, time int64) string {
	sessionData := fmt.Sprintf("%s:%X", name, time)
	h := hmac.New(sha1.New, []byte(secret+salt))
	h.Write([]byte(sessionData))
	hashData := string(h.Sum(nil))
	b64Data := base64.StdEncoding.EncodeToString([]byte(sessionData + ":" + hashData))
	return b64Data
}

func deleteSession(w http.ResponseWriter, r *http.Request) error {
	return nil
}
