package couchserver

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/go-kivik/kivik/v4"
)

const (
	typeJSON = "application/json"
)

type db interface {
	Stats(context.Context) (*kivik.DBStats, error)
	Flush(context.Context) error
}

type backend interface {
	AllDBs(context.Context, ...kivik.Option) ([]string, error)
	CreateDB(context.Context, string, ...kivik.Option) error
	DB(context.Context, string, ...kivik.Option) (db, error)
	DBExists(context.Context, string, ...kivik.Option) (bool, error)
}

type clientWrapper struct {
	*kivik.Client
}

var _ backend = &clientWrapper{}

func (c *clientWrapper) DB(_ context.Context, dbName string, options ...kivik.Option) (db, error) {
	db := c.Client.DB(dbName, options...)
	return db, db.Err()
}

// Handler is a CouchDB server handler.
type Handler struct {
	client backend
	// CompatVersion is the CouchDB compatibility version to report. If unset,
	// defaults to the CompatVersion constant/.
	CompatVersion string
	// Vendor is the vendor name to report. If unset, defaults to the
	// kivik.Vendor constant.
	Vendor string
	// VendorVersion is the vendor version to report. If unset, defaults to the
	// kivik.VendorVersion constant.
	VendorVersion string
	Logger        *log.Logger
	// Favicon is the path to a favicon.ico to serve.
	Favicon string
	// SessionKey is a temporary solution to avoid import cycles. Soon I will move the key to another package.
	SessionKey interface{}
}

func NewHandler(client *kivik.Client) *Handler {
	return &Handler{client: &clientWrapper{client}}
}

// CompatVersion is the default CouchDB compatibility provided by this package.
const CompatVersion = "0.0.0"

func (h *Handler) vendor() (compatVer, vend, ver string) {
	if h.CompatVersion == "" {
		compatVer = CompatVersion
	} else {
		compatVer = h.CompatVersion
	}
	if h.Vendor == "" {
		vend = "Kivik"
	} else {
		vend = h.Vendor
	}
	if h.VendorVersion == "" {
		ver = kivik.Version
	} else {
		ver = h.VendorVersion
	}
	return compatVer, vend, ver
}

// Main returns an http.Handler to handle all CouchDB endpoints.
func (h *Handler) Main() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.GetRoot())
	r.Get("/favicon.ico", h.GetFavicon())
	r.Get("/_all_dbs", h.GetAllDBs())
	r.Get("/{db}", h.GetDB())
	r.Put("/{db}", h.PutDB())
	r.Head("/{db}", h.HeadDB())
	r.Post("/{db}/_ensure_full_commit", h.Flush())
	r.Get("/_session", h.GetSession())
	return r
}

type serverInfo struct {
	CouchDB string     `json:"couchdb"`
	Version string     `json:"version"`
	Vendor  vendorInfo `json:"vendor"`
}

type vendorInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
