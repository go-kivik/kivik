// Package kivik is an assemble-your-own CouchDB API, proxy, cache, and server..
package kivik

import (
	"fmt"

	"github.com/flimzy/kivik/driver"
)

// Server is a connection handle to a CouchDB-like server.
type Server struct {
	driver driver.Driver
	dsn    string
	conn   driver.Conn
}

// Open opens a connection to a server specified by its database driver name and
// a driver-specific data source name.
func Open(driverName, dataSourceName string) (*Server, error) {
	driversMu.RLock()
	driveri, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("kivik: unknown driver %q (forgotten import?)", driverName)
	}
	conn, err := driveri.Open(dataSourceName)
	if err != nil {
		return nil, err
	}
	return &Server{
		driver: driveri,
		dsn:    dataSourceName,
		conn:   conn,
	}, nil
}

// Version returns the reported server version
func (s *Server) Version() (string, error) {
	si, err := s.conn.ServerInfo()
	return si.Version(), err
}
