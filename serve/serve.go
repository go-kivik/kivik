package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/authdb"
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/errors"
)

// Version is the version of this library.
const Version = "0.0.1"

// Vendor is the identifying vendor string for this library.
const Vendor = "Kivik"

// CompatVersion is the default compatibility version reported to clients.
const CompatVersion = "1.6.1"

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

	// config is the configuration backend. If none is specified, anew memconf
	// instance is instantiated the first time configuration is requested.
	config *config.Config

	// authHandlers is a map version of AuthHandlers for easier internal
	// use.
	authHandlers     map[string]auth.Handler
	authHandlerNames []string
}

// Config returns a connection to the configuration backend.
func (s *Service) Config() *config.Config {
	if s.config == nil {
		s.config = defaultConfig()
	}
	return s.config
}

// SetConfig sets the configuration backend.
func (s *Service) SetConfig(config *config.Config) {
	s.config = config
}

// Init initializes a configured server. This is automatically called when
// Start() is called, so this is meant to be used if you want to bind the server
// yourself.
func (s *Service) Init() (http.Handler, error) {
	s.authHandlersSetup()
	if s.Config().GetString(context.TODO(), "couch_httpd_auth", "secret") == "" {
		fmt.Fprintf(os.Stderr, "couch_httpd_auth.secret is not set. This is insecure!\n")
	}
	return s.setupRoutes()
}

// Start begins serving connections.
func (s *Service) Start() error {
	server, err := s.Init()
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d",
		s.Config().GetString(context.TODO(), "httpd", "bind_address"),
		s.Config().GetInt(context.TODO(), "httpd", "port"),
	)
	fmt.Fprintf(os.Stderr, "Listening on %s\n", addr)
	return http.ListenAndServe(addr, server)
}

func (s *Service) authHandlersSetup() {
	if s.AuthHandlers == nil || len(s.AuthHandlers) == 0 {
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
		return errors.Wrapf(err, "invalid port '%s'", port)
	}
	host := strings.TrimSuffix(addr, ":"+port)
	s.Config().Set(context.TODO(), "httpd", "bind_address", host)
	s.Config().Set(context.TODO(), "httpd", "port", port)
	return nil
}

const (
	mGET    = http.MethodGet
	mPUT    = http.MethodPut
	mHEAD   = http.MethodHead
	mPOST   = http.MethodPost
	mDELETE = http.MethodDelete
	mCOPY   = "COPY"
)

type vendorInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type serverInfo struct {
	CouchDB string     `json:"couchdb"`
	Version string     `json:"version"`
	Vendor  vendorInfo `json:"vendor"`
}

const (
	typeJSON  = "application/json"
	typeText  = "text/plain"
	typeForm  = "application/x-www-form-urlencoded"
	typeMForm = "multipart/form-data"
)

type handler func(w http.ResponseWriter, r *http.Request) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		reportError(w, err)
	}
}

func reportError(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", typeJSON)
	status := errors.StatusCode(err)
	if status == 0 {
		status = 500
	}
	w.WriteHeader(status)
	short := err.Error()
	reason := errors.Reason(err)
	if reason == "" {
		reason = short
	} else {
		short = strings.ToLower(http.StatusText(status))
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  short,
		"reason": reason,
	})
}

func root(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", typeJSON)
	svc := GetService(r)
	vendVers := svc.VendorVersion
	if vendVers == "" {
		vendVers = Version
	}
	vendName := svc.VendorName
	if vendName == "" {
		vendName = Vendor
	}
	return json.NewEncoder(w).Encode(serverInfo{
		CouchDB: "Välkommen",
		Version: CompatVersion,
		Vendor: vendorInfo{
			Name:    vendName,
			Version: vendVers,
		},
	})
}

func allDBs(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", typeJSON)
	client := getClient(r)
	dbs, err := client.AllDBs(r.Context())
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(dbs)
}

func createDB(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", typeJSON)
	params := getParams(r)
	client := getClient(r)
	if err := client.CreateDB(r.Context(), params["db"]); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
	})
}

func dbExists(w http.ResponseWriter, r *http.Request) error {
	params := getParams(r)
	client := getClient(r)
	exists, err := client.DBExists(r.Context(), params["db"])
	if err != nil {
		return err
	}
	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
	return nil
}

// serveJSON serves i as JSON to w.
func serveJSON(w http.ResponseWriter, i interface{}) error {
	w.Header().Set("Content-Type", typeJSON)
	return json.NewEncoder(w).Encode(i)
}
