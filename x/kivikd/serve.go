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

package kivikd

import (
	"context"
	"encoding/json"
	errs "errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/kivikd/auth"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/conf"
	"github.com/go-kivik/kivik/v4/x/kivikd/logger"
)

// Service defines a CouchDB-like service to serve. You will define one of these
// per server endpoint.
type Service struct {
	// Client is an instance of a driver.Client, which will be served.
	Client *kivik.Client
	// UserStore provides access to the user database. This is passed to auth
	// handlers, and is used to authenticate sessions. If unset, a nil UserStore
	// will be used which authenticates all uses. PERPETUAL ADMIN PARTY!
	UserStore authdb.UserStore
	// AuthHandler is a slice of authentication handlers. If no auth
	// handlers are configured, the server will operate as a PERPETUAL
	// ADMIN PARTY!
	AuthHandlers []auth.Handler
	// CompatVersion is the compatibility version to report to clients. Defaults
	// to 1.6.1.
	CompatVersion string
	// VendorVersion is the vendor version string to report to clients. Defaults to the library
	// version.
	VendorVersion string
	// VendorName is the vendor name string to report to clients. Defaults to the library
	// vendor string.
	VendorName string
	// Favicon is the path to a file to serve as favicon.ico. If unset, a default
	// image is used.
	Favicon string
	// RequestLogger receives logging information for each request.
	RequestLogger logger.RequestLogger

	// ConfigFile is the path to a config file to read during startup.
	ConfigFile string

	// Config is a complete config object. If this is set, config loading is
	// bypassed.
	Config *conf.Conf

	conf   *conf.Conf
	confMU sync.RWMutex

	// authHandlers is a map version of AuthHandlers for easier internal
	// use.
	authHandlers     map[string]auth.Handler
	authHandlerNames []string
}

// Init initializes a configured server. This is automatically called when
// Start() is called, so this is meant to be used if you want to bind the server
// yourself.
func (s *Service) Init() (http.Handler, error) {
	s.authHandlersSetup()
	if err := s.loadConf(); err != nil {
		return nil, err
	}
	if !s.Conf().IsSet("couch_httpd_auth.secret") {
		fmt.Fprintf(os.Stderr, "couch_httpd_auth.secret is not set. This is insecure!\n")
	}
	return s.setupRoutes()
}

func (s *Service) loadConf() error {
	s.confMU.Lock()
	defer s.confMU.Unlock()
	if s.Config != nil {
		s.conf = s.Config
		return nil
	}
	c, err := conf.Load(s.ConfigFile)
	if err != nil {
		return err
	}
	s.conf = c
	return nil
}

// Conf returns the initialized server configuration.
func (s *Service) Conf() *conf.Conf {
	s.confMU.RLock()
	defer s.confMU.RUnlock()
	if s.Config != nil {
		s.confMU.RUnlock()
		if err := s.loadConf(); err != nil {
			panic(err)
		}
		s.confMU.RLock()
	}
	return s.conf
}

// Start begins serving connections.
func (s *Service) Start() error {
	server, err := s.Init()
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d",
		s.Conf().GetString("httpd.bind_address"),
		s.Conf().GetInt("httpd.port"),
	)
	fmt.Fprintf(os.Stderr, "Listening on %s\n", addr)
	return http.ListenAndServe(addr, server)
}

func (s *Service) authHandlersSetup() {
	if len(s.AuthHandlers) == 0 {
		fmt.Fprintf(os.Stderr, "No AuthHandler specified! Welcome to the PERPETUAL ADMIN PARTY!\n")
	}
	s.authHandlers = make(map[string]auth.Handler)
	s.authHandlerNames = make([]string, 0, len(s.AuthHandlers))
	for _, handler := range s.AuthHandlers {
		name := handler.MethodName()
		if _, ok := s.authHandlers[name]; ok {
			panic(fmt.Sprintf("Multiple auth handlers for for `%s` registered", name))
		}
		s.authHandlers[name] = handler
		s.authHandlerNames = append(s.authHandlerNames, name)
	}
	if s.UserStore == nil {
		s.UserStore = &perpetualAdminParty{}
	}
}

type perpetualAdminParty struct{}

var _ authdb.UserStore = &perpetualAdminParty{}

func (p *perpetualAdminParty) Validate(ctx context.Context, username, _ string) (*authdb.UserContext, error) {
	return p.UserCtx(ctx, username)
}

func (p *perpetualAdminParty) UserCtx(_ context.Context, username string) (*authdb.UserContext, error) {
	return &authdb.UserContext{
		Name:  username,
		Roles: []string{"_admin"},
	}, nil
}

// Bind sets the HTTP daemon bind address and port.
func (s *Service) Bind(addr string) error {
	port := addr[strings.LastIndex(addr, ":")+1:]
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("invalid port '%s': %w", port, err)
	}
	host := strings.TrimSuffix(addr, ":"+port)
	s.Conf().Set("httpd.bind_address", host)
	s.Conf().Set("httpd.port", port)
	return nil
}

const (
	typeJSON = "application/json"
	// typeText  = "text/plain"
	typeForm = "application/x-www-form-urlencoded"
	// typeMForm = "multipart/form-data"
)

func reason(err error) string {
	kerr := new(internal.Error)
	if errs.As(err, &kerr) {
		return kerr.Message
	}
	return err.Error()
}

func reportError(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", typeJSON)
	status := kivik.HTTPStatus(err)
	w.WriteHeader(status)
	short := err.Error()
	reason := reason(err)
	if reason == "" {
		reason = short
	} else {
		short = strings.ToLower(http.StatusText(status))
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  short,
		"reason": reason,
	})
}
