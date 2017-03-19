package kivik

import (
	"io"

	"github.com/flimzy/kivik/driver"
)

// DBUpdate represents a single DB Update event.
type DBUpdate struct {
	DBName string
	Seq    string
	Type   string
}

// DBUpdateFeed provides access to database updates.
type DBUpdateFeed struct {
	feed  <-chan *driver.DBUpdate
	close func() error
}

// Next returns the next DBUpdate from the feed. This function will block
// until an event is received. If an error occurs, it will be returned and
// the feed closed. If the feed was closed normally, io.EOF will be returned
// when there are no more events in the buffer.
func (f *DBUpdateFeed) Next() (*DBUpdate, error) {
	event, ok := <-f.feed
	if !ok {
		return nil, io.EOF
	}
	if event.Error != nil {
		_ = f.Close()
		return nil, event.Error
	}
	return &DBUpdate{
		DBName: event.DBName,
		Seq:    event.Seq,
		Type:   event.Type,
	}, nil
}

// Close closes the feed. Any unread updates will still be accessible via
// Next().
func (f *DBUpdateFeed) Close() error {
	return f.close()
}

// DBUpdates begins polling for database updates.
func (c *Client) DBUpdates() (*DBUpdateFeed, error) {
	updater, ok := c.driverClient.(driver.DBUpdater)
	if !ok {
		return nil, ErrNotImplemented
	}
	updateChan, closeFunc, err := updater.DBUpdates()
	if err != nil {
		return nil, err
	}
	return &DBUpdateFeed{
		feed:  updateChan,
		close: closeFunc,
	}, nil
}
