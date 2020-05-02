/*
Package kivik provides a generic interface to CouchDB or CouchDB-like databases.

The kivik package must be used in conjunction with a database driver. See
https://github.com/go-kivik/kivik/wiki/Kivik-database-drivers for a list.

The kivik driver system is modeled after the standard library's sql and
sql/driver packages, although the client API is completely different due to
the different  database models implemented by SQL and NoSQL databases such as
CouchDB.

Error Handling

Kivik returns errors that embed an HTTP status code. In most cases, this is the
HTTP status code returned by the server. The embedded HTTP status code may be
accessed easily using the StatusCode() method, or with a type assertion to
`interface { StatusCode() int }`. Example:

    if statusErr, ok := err.(interface{ StatusCode() int }); ok {
		status = statusErr.StatusCode()
	}

Any error that does not conform to this interface will be assumed to represent
a http.StatusInternalServerError status code.
*/
package kivik // import "github.com/go-kivik/kivik/v4"
