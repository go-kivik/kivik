package serve

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/dimfeld/httptreemux"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/proxy"
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
	Client driver.Client
	// CompatVersion is the compatibility version to report to clients. Defaults
	// to 1.6.1.
	CompatVersion string
	// VendorVersion is the vendor version string to report to clients. Defaults to the library
	// version.
	VendorVersion string
	// VendorName is the vendor name string to report to clients. Defaults to the library
	// vendor string.
	VendorName string
	// Logger is a logger where logs can be sent. It is the responsibility of
	// the developer to ensure that the logs are available to the backend driver,
	// if the Log() method is expected to work. By default, a null logger is used
	// which discards all log messages.
	Logger LogWriter
}

// NewKivikClient returns a new service to serve a standard *kivik.Client.
// This is the same as:
//
//    import "github.com/flimzy/kivik/driver/proxy"
//
//    New(proxy.NewClient(client))
func NewKivikClient(client *kivik.Client) *Service {
	return New(proxy.NewClient(client))
}

// New returns a new service definition.
func New(backend driver.Client) *Service {
	return &Service{Client: backend}
}

// Start begins serving connections on the requested bind address.
func (s *Service) Start(addr string) error {
	server, err := s.Server()
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, server)
}

// ContextKey is a type for context keys.
type ContextKey string

// ContextKeys are used to store values in the context passed to HTTP handlers.
const (
	ClientContextKey  ContextKey = "kivik client"
	ServiceContextKey ContextKey = "kivik service"
)

const (
	mGET    = http.MethodGet
	mPUT    = http.MethodPut
	mHEAD   = http.MethodHead
	mDELETE = http.MethodDelete
	mCOPY   = "COPY"
)

// Server returns an unstarted server instance.
func (s *Service) Server() (http.Handler, error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ClientContextKey, s.Client)
	ctx = context.WithValue(ctx, ServiceContextKey, s)
	router := httptreemux.New()
	router.HeadCanUseGet = true
	router.DefaultContext = ctx
	ctxRoot := router.UsingContext()
	ctxRoot.Handler(mGET, "/", handler(root))
	ctxRoot.Handler(mGET, "/_all_dbs", handler(allDBs))
	ctxRoot.Handler(mGET, "/_log", handler(log))
	ctxRoot.Handler(mPUT, "/:db", handler(createDB))
	ctxRoot.Handler(mHEAD, "/:db", handler(dbExists))
	// ctxRoot.Handler(mDELETE, "/:db", handler(destroyDB) )
	// ctxRoot.Handler(http.MethodGet, "/:db", handler(getDB))
	return router, nil
}

func getService(r *http.Request) *Service {
	service := r.Context().Value(ServiceContextKey).(*Service)
	return service
}

func getClient(r *http.Request) driver.Client {
	client := r.Context().Value(ClientContextKey).(driver.Client)
	return client
}

func getParams(r *http.Request) map[string]string {
	return httptreemux.ContextParams(r.Context())
}

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
	typeJSON = "application/json"
	typeText = "text/plain"
)

type handler func(w http.ResponseWriter, r *http.Request) error

type statusCoder interface {
	StatusCode() int
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}
	var status int
	if statuser, ok := err.(statusCoder); ok {
		status = statuser.StatusCode()
	}
	if status == 0 {
		status = http.StatusInternalServerError
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func root(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", typeJSON)
	svc := getService(r)
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
	dbs, err := client.AllDBs()
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(dbs)
}

func createDB(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", typeJSON)
	params := getParams(r)
	client := getClient(r)
	spew.Dump(client)
	if err := client.CreateDB(params["db"]); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
	})
}

func dbExists(w http.ResponseWriter, r *http.Request) error {
	params := getParams(r)
	client := getClient(r)
	exists, err := client.DBExists(params["db"])
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
