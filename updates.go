package kivik

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v3/driver"
)

// DBUpdates provides access to database updates.
type DBUpdates struct {
	*iter
	updatesi driver.DBUpdates
}

type updatesIterator struct{ driver.DBUpdates }

var _ iterator = &updatesIterator{}

func (r *updatesIterator) Next(i interface{}) error { return r.DBUpdates.Next(i.(*driver.DBUpdate)) }

func newDBUpdates(ctx context.Context, updatesi driver.DBUpdates) *DBUpdates {
	return &DBUpdates{
		iter:     newIterator(ctx, &updatesIterator{updatesi}, &driver.DBUpdate{}),
		updatesi: updatesi,
	}
}

// DBName returns the database name for the current update.
func (f *DBUpdates) DBName() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).DBName
}

// Type returns the type of the current update.
func (f *DBUpdates) Type() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).Type
}

// Seq returns the update sequence of the current update.
func (f *DBUpdates) Seq() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).Seq
}

// DBUpdates begins polling for database updates.
func (c *Client) DBUpdates(ctx context.Context) (*DBUpdates, error) {
	updater, ok := c.driverClient.(driver.DBUpdater)
	if !ok {
		return nil, &Error{HTTPStatus: http.StatusNotImplemented, Message: "kivik: driver does not implement DBUpdater"}
	}
	updatesi, err := updater.DBUpdates(ctx)
	if err != nil {
		return nil, err
	}
	return newDBUpdates(context.Background(), updatesi), nil
}
