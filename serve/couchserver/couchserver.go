package couchserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pressly/chi"

	"github.com/flimzy/kivik"
)

const (
	typeJSON  = "application/json"
	typeText  = "text/plain"
	typeForm  = "application/x-www-form-urlencoded"
	typeMForm = "multipart/form-data"
)

// Handler is a CouchDB server handler.
type Handler struct {
	*kivik.Client
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
}

var _ http.Handler = &Handler{}

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
		ver = kivik.KivikVersion
	} else {
		ver = h.VendorVersion
	}
	return compatVer, vend, ver
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r := chi.NewRouter()
	r.Get("/", h.GetRoot())

	/*
	   ctxRoot.Handler(mGET, "/", handler(root))
	   ctxRoot.Handler(mGET, "/favicon.ico", handler(favicon))
	   ctxRoot.Handler(mGET, "/_all_dbs", handler(allDBs))
	   ctxRoot.Handler(mPUT, "/:db", handler(createDB))
	   ctxRoot.Handler(mHEAD, "/:db", handler(dbExists))
	   ctxRoot.Handler(mPOST, "/:db/_ensure_full_commit", handler(flush))
	*/
	r.ServeHTTP(w, req)
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

// GetRoot handles requests for: GET /
func (h *Handler) GetRoot() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		compatVer, vendName, vendVers := h.vendor()
		w.Header().Set("Content-Type", typeJSON)
		h.HandleError(w, json.NewEncoder(w).Encode(serverInfo{
			CouchDB: "Välkommen",
			Version: compatVer,
			Vendor: vendorInfo{
				Name:    vendName,
				Version: vendVers,
			},
		}))
	})
}
