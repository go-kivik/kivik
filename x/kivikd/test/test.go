//go:build !js
// +build !js

package test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/spf13/viper"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
	"github.com/go-kivik/kivikd/v4"
	"github.com/go-kivik/kivikd/v4/auth"
	"github.com/go-kivik/kivikd/v4/auth/basic"
	"github.com/go-kivik/kivikd/v4/auth/cookie"
	"github.com/go-kivik/kivikd/v4/authdb/confadmin"
	"github.com/go-kivik/kivikd/v4/conf"
	"github.com/go-kivik/proxydb/v4"

	_ "github.com/go-kivik/kivik/v4/couchdb"    // CouchDB driver
	_ "github.com/go-kivik/kivik/v4/x/memorydb" // Memory driver
)

// RegisterKivikdSuites registers the Kivikd related integration test suites.
func RegisterKivikdSuites() {
	kiviktest.RegisterSuite(kiviktest.SuiteKivikServer, kt.SuiteConfig{
		"AllDBs.expected": []string{},
		"AllDBs/RW.skip":  true, // FIXME: Enable this when it's possible to delete DB from the server

		"CreateDB/RW.skip": true, // FIXME: Update when the server can destroy databases
		// "CreateDB/NoAuth.status":         http.StatusUnauthorized,
		// "CreateDB/Admin/Recreate.status": http.StatusPreconditionFailed,

		"DestroyDB.skip": true, // FIXME: Update when the server can destroy databases

		"AllDocs/Admin.databases":   []string{"foo"},
		"AllDocs/Admin/foo.status":  http.StatusNotFound,
		"AllDocs/RW.skip":           true, // FIXME: Update when the server can destroy databases
		"AllDocs/NoAuth.databases":  []string{"foo"},
		"AllDocs/NoAuth/foo.status": http.StatusNotFound,

		"DBExists.databases":            []string{"chicken"},
		"DBExists/Admin/chicken.exists": false,
		"DBExists/RW.skip":              true, // FIXME: Update when the server can destroy databases
		// "DBExists/RW/Admin.exists":      true,
		"DBExists/NoAuth.skip": true, // TODO

		"Log/Admin/Offset-1000.status":        http.StatusBadRequest,
		"Log/Admin/HTTP/TextBytes.status":     http.StatusBadRequest,
		"Log/Admin/HTTP/NegativeBytes.status": http.StatusBadRequest,
		"Log/NoAuth/Offset-1000.status":       http.StatusBadRequest,

		"Version.version":        `^0\.0\.0$`,
		"Version.vendor":         "Kivik",
		"Version.vendor_version": `^0\.0\.1$`,

		"Get.skip": true, // FIXME: Fix this when we can delete database

		"Put.skip": true, // FIXME: Fix this when we can write docs

		"Flush.databases":                     []string{"chicken"},
		"Flush/Admin/chicken/DoFlush.status":  http.StatusNotImplemented, // FIXME: Update when implemented
		"Flush/NoAuth/chicken/DoFlush.status": http.StatusNotImplemented, // FIXME: Update when implemented

		"Delete.skip": true, // FIXME: Fix this when we can delete docs.

		"Session/Get/Admin.info.authentication_handlers":  "default,cookie",
		"Session/Get/Admin.info.authentication_db":        "",
		"Session/Get/Admin.info.authenticated":            "cookie",
		"Session/Get/Admin.userCtx.roles":                 "_admin",
		"Session/Get/Admin.ok":                            "true",
		"Session/Get/NoAuth.info.authentication_handlers": "default,cookie",
		"Session/Get/NoAuth.info.authentication_db":       "",
		"Session/Get/NoAuth.info.authenticated":           "",
		"Session/Get/NoAuth.userCtx.roles":                "",
		"Session/Get/NoAuth.ok":                           "true",

		"Session/Post/EmptyJSON.status":                               http.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                           http.StatusBadRequest,
		"Session/Post/BogusTypeForm.status":                           http.StatusBadRequest,
		"Session/Post/EmptyForm.status":                               http.StatusBadRequest,
		"Session/Post/BadJSON.status":                                 http.StatusBadRequest,
		"Session/Post/BadForm.status":                                 http.StatusBadRequest,
		"Session/Post/MeaninglessJSON.status":                         http.StatusBadRequest,
		"Session/Post/MeaninglessForm.status":                         http.StatusBadRequest,
		"Session/Post/GoodJSON.status":                                http.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                           http.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                            http.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                            http.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirAbsolute.status":        http.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirRelativeNoSlash.status":       http.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.status": http.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.status":      http.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirEmpty.status":                 http.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirSchemaless.status":            http.StatusBadRequest,

		"Stats.skip":             true, // FIXME: Unimplemented
		"CreateDoc.skip":         true, // FIXME: Unimplemented
		"Compact.skip":           true, // FIXME: Unimplemented
		"ViewCleanup.skip":       true, // FIXME: Unimplemented
		"Security.skip":          true, // FIXME: Unimplemented
		"SetSecurity.skip":       true, // FIXME: Unimplemented
		"GetRev.skip":            true, // FIXME: When Get works
		"DBUpdates.skip":         true, // FIXME: Unimplemented
		"Changes.skip":           true, // FIXME: Unimplemented
		"Copy.skip":              true, // FIXME: Unimplemented, depends on Get/Put or Copy
		"BulkDocs.skip":          true, // FIXME: Unimplemented
		"GetAttachment.skip":     true, // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true, // FIXME: Unimplemented
		"PutAttachment.skip":     true, // FIXME: Unimplemented
		"DeleteAttachment.skip":  true, // FIXME: Unimplemented
		"Query.skip":             true, // FIXME: Unimplemented
		"Find.skip":              true, // FIXME: Unimplemented
		"Explain.skip":           true, // FIXME: Unimplemented
		"CreateIndex.skip":       true, // FIXME: Unimplemented
		"GetIndexes.skip":        true, // FIXME: Unimplemented
		"DeleteIndex.skip":       true, // FIXME: Unimplemented
		"GetReplications.skip":   true, // FIXME: Unimplemented
		"Replicate.skip":         true, // FIXME: Unimplemented
	})
}

type customDriver struct {
	driver.Client
}

func (cd customDriver) NewClient(string, driver.Options) (driver.Client, error) {
	return cd, nil
}

// ServerTest tests the kivikd server
func ServerTest(t *testing.T) {
	memClient, err := kivik.New("memory", "")
	if err != nil {
		t.Fatalf("Failed to connect to memory driver: %s", err)
	}
	kivik.Register("custom", customDriver{proxydb.NewClient(memClient)})
	backend, err := kivik.New("custom", "")
	if err != nil {
		t.Fatalf("Failed to connect to custom driver: %s", err)
	}
	c := &conf.Conf{Viper: viper.New()}
	// Set admin/abc123 credentials
	c.Set("admins.admin", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	service := kivikd.Service{}
	service.Config = c
	service.Client = backend
	service.UserStore = confadmin.New(c)
	service.AuthHandlers = []auth.Handler{
		&basic.HTTPBasicAuth{},
		&cookie.Auth{},
	}
	handler, err := service.Init()
	if err != nil {
		t.Fatalf("Failed to initialize server: %s\n", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	dsn, _ := url.Parse(server.URL)
	dsn.User = url.UserPassword("admin", "abc123")
	clients, err := kiviktest.ConnectClients(t, "couch", dsn.String(), nil)
	if err != nil {
		t.Fatalf("Failed to initialize client: %s", err)
	}
	clients.RW = true
	kiviktest.RunTestsInternal(clients, kiviktest.SuiteKivikServer)
}
