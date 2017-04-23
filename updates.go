package kivik

import (
	"github.com/flimzy/kivik/driver"
	"golang.org/x/net/context"
)

// DBUpdate represents a single DB Update event.
type DBUpdate struct {
	DBName string
	Seq    string
	Type   string
}

// DBUpdateFeed provides access to database updates.
type DBUpdateFeed struct {
	*iter
	updatesi driver.DBUpdates
}

// Next returns the next DBUpdate from the feed. This function will block
// until an event is received. If an error occurs, it will be returned and
// the feed closed. If the feed was closed normally, io.EOF will be returned
// when there are no more events in the buffer.
func (f *DBUpdateFeed) Next() bool {
	return f.iter.Next()
}

// Close closes the feed. Any unread updates will still be accessible via
// Next().
func (f *DBUpdateFeed) Close() error {
	return f.iter.Close()
}

// Err returns the error, if any, that was encountered during iteration. Err
// may be called after an explicit or implicit Close.
func (f *DBUpdateFeed) Err() error {
	return f.iter.Err()
}

type updatesIterator struct{ driver.DBUpdates }

var _ iterator = &updatesIterator{}

func (r *updatesIterator) Next(i interface{}) error { return r.DBUpdates.Next(i.(*driver.DBUpdate)) }

func newDBUpdates(ctx context.Context, updatesi driver.DBUpdates) *DBUpdateFeed {
	return &DBUpdateFeed{
		iter:     newIterator(ctx, &updatesIterator{updatesi}, &driver.DBUpdate{}),
		updatesi: updatesi,
	}
}

// DBName returns the database name for the current update.
func (f *DBUpdateFeed) DBName() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).DBName
}

// Type returns the type of the current update.
func (f *DBUpdateFeed) Type() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).Type
}

// Seq returns the update sequence of the current update.
func (f *DBUpdateFeed) Seq() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).Seq
}

// DBUpdates begins polling for database updates.
func (c *Client) DBUpdates() (*DBUpdateFeed, error) {
	updater, ok := c.driverClient.(driver.DBUpdater)
	if !ok {
		return nil, ErrNotImplemented
	}
	updatesi, err := updater.DBUpdates()
	if err != nil {
		return nil, err
	}
	return newDBUpdates(context.Background(), updatesi), nil
}
