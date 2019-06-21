// Package kivik provides a generic interface to CouchDB or CouchDB-like databases.
//
// The kivik package must be used in conjunction with a database driver. See
// https://github.com/go-kivik/kivik/wiki/Kivik-database-drivers for a list.
//
// The kivik driver system is modeled after the standard library's sql and
// sql/driver packages, although the client API is completely different due to
// the different  database models implemented by SQL and NoSQL databases such as
// CouchDB.
package kivik // import "github.com/go-kivik/kivik"
