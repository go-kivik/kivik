package memory

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

func (d *db) Security(_ context.Context) (*driver.Security, error) {
	if d.db.deleted {
		return nil, errors.Status(kivik.StatusNotFound, "missing")
	}
	// deep copy
	return &driver.Security{
		Admins: driver.Members{
			Names: d.db.security.Admins.Names,
			Roles: d.db.security.Admins.Roles,
		},
		Members: driver.Members{
			Names: d.db.security.Members.Names,
			Roles: d.db.security.Members.Roles,
		},
	}, nil
}

func (d *db) SetSecurity(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return notYetImplemented
}
