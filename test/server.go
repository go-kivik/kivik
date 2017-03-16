package test

import (
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikServer, kt.SuiteConfig{
		"AllDBs.expected": []string{},
		"AllDBs/RW.skip":  true, // FIXME: Enable this when it's possible to delete DB from the server

		"Config/Admin/GetAll.expected_sections":      []string{"admins", "log"},
		"Config/Admin/GetSection.sections":           []string{"log", "chicken"},
		"Config/Admin/GetSection/log.keys":           []string{"capacity"},
		"Config/Admin/GetSection/chicken.keys":       []string{},
		"Config/Admin/GetItem.items":                 []string{"log.capacity", "log.foobar", "logx.level"},
		"Config/Admin/GetItem/log.foobar.status":     http.StatusNotFound,
		"Config/Admin/GetItem/logx.level.status":     http.StatusNotFound,
		"Config/Admin/GetItem/log.capacity.expected": "10",
		"Config/NoAuth.skip":                         true, // FIXME: Update this when the server supports auth
		"Config/RW/NoAuth.skip":                      true, // FIXME: Update this when the server supports auth
		"Config/RW.skip":                             true, // FIXME: Update this when the server can write config

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

		"Membership.status": http.StatusMethodNotAllowed, // FIXME: Make the server respond with 404, or unimplemented

		"UUIDs/Admin.counts":          []int{-1, 0, 1, 10},
		"UUIDs.status":                http.StatusMethodNotAllowed, // FIXME: Implement UUIDs in the server
		"UUIDs/Admin/-1Count.status":  http.StatusBadRequest,
		"UUIDs/NoAuth.counts":         []int{-1, 0, 1, 10},
		"UUIDs/NoAuth/-1Count.status": http.StatusBadRequest,

		"Log/Admin/Offset-1000.status":        http.StatusBadRequest,
		"Log/Admin/HTTP/TextBytes.status":     http.StatusBadRequest,
		"Log/Admin/HTTP/NegativeBytes.status": http.StatusBadRequest,
		"Log/NoAuth/Offset-1000.status":       http.StatusBadRequest,

		"ServerInfo.version":        `^1\.6\.1$`,
		"ServerInfo.vendor":         "Kivik",
		"ServerInfo.vendor_version": `^0\.0\.1$`,

		"Get.skip": true, // FIXME: Fix this when we can delete database

		"Put.skip": true, // FIXME: Fix this when we can write docs

		"Flush.databases":                     []string{"chicken"},
		"Flush/Admin/chicken/DoFlush.status":  kivik.StatusNotImplemented, // FIXME: Update when implemented
		"Flush/NoAuth/chicken/DoFlush.status": kivik.StatusNotImplemented, // FIXME: Update when implemented

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

		"Session/Post/EmptyJSON.status":                               kivik.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                           kivik.StatusBadRequest,
		"Session/Post/BogusTypeForm.status":                           kivik.StatusBadRequest,
		"Session/Post/EmptyForm.status":                               kivik.StatusBadRequest,
		"Session/Post/BadJSON.status":                                 kivik.StatusBadRequest,
		"Session/Post/BadForm.status":                                 kivik.StatusBadRequest,
		"Session/Post/MeaninglessJSON.status":                         kivik.StatusBadRequest,
		"Session/Post/MeaninglessForm.status":                         kivik.StatusBadRequest,
		"Session/Post/GoodJSON.status":                                kivik.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                           kivik.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                            kivik.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                            kivik.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirAbsolute.status":        kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirRelativeNoSlash.status":       kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.status": kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.status":      kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirEmpty.status":                 kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirSchemaless.status":            kivik.StatusBadRequest,

		"DBInfo.skip":    true, // FIXME: Unimplemented
		"CreateDoc.skip": true, // FIXME: Unimplemented
		"Compact.skip":   true, // FIXME: Unimplemented
	})
}
