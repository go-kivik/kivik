package memory

import (
	"context"

	"github.com/flimzy/kivik/driver"
)

func (d *db) Security(_ context.Context) (*driver.Security, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) SetSecurity(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return notYetImplemented
}
