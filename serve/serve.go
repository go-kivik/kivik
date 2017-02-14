package serve

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dimfeld/httptreemux"

	"github.com/flimzy/kivik"
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
	// DriverName is the name of the registerd driver.
	DriverName string
	// DataSourceName is the name passed to the driver's NewClient() method.
	// Typically the path to a remote server, or equivalent.
	DataSourceName string
	// CompatVersion is the compatibility version to report to clients. Defaults
	// to 1.6.1.
	CompatVersion string
	// VendorVersion is the vendor version string to report to clients. Defaults to the library
	// version.
	VendorVersion string
	// VendorName is the vendor name string to report to clients. Defaults to the library
	// vendor string.
	VendorName string
}

// New returns a new service definition.
func New(driverName, dataSource string) *Service {
	return &Service{DriverName: driverName, DataSourceName: dataSource}
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
	client, err := kivik.New(s.DriverName, s.DataSourceName)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, ClientContextKey, client)
	ctx = context.WithValue(ctx, ServiceContextKey, s)
	router := httptreemux.New()
	router.HeadCanUseGet = true
	router.DefaultContext = ctx
	ctxRoot := router.UsingContext()
	ctxRoot.Handler(mGET, "/", handler(root))
	ctxRoot.Handler(mGET, "/_all_dbs", handler(allDBs))
	ctxRoot.Handler(mPUT, "/:db", handler(createDB))
	ctxRoot.Handler(mHEAD, "/:db", handler(dbExists))
	// ctxRoot.Handler(mDELETE, "/:db", handler(destroyDB) )
	// ctxRoot.Handler(http.MethodGet, "/:db", handler(getDB))
	return router, nil
}

func getService(r *http.Request) *Service {
	service, _ := r.Context().Value(ServiceContextKey).(*Service)
	return service
}

func getClient(r *http.Request) *kivik.Client {
	client, _ := r.Context().Value(ClientContextKey).(*kivik.Client)
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

const jsonType = "application/json"

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
	w.Header().Set("Content-Type", jsonType)
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
	w.Header().Set("Content-Type", jsonType)
	client := getClient(r)
	dbs, err := client.AllDBs()
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(dbs)
}

func createDB(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", jsonType)
	params := getParams(r)
	client := getClient(r)
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
