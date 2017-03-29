package driver

// DBUpdate represents a database update event.
type DBUpdate struct {
	DBName string `json:"db_name"`
	Type   string `json:"type"`
	Seq    string `json:"seq"`
	Error  error  `json:"-"`
}

// DBUpdater is an optional interface that may be implemented by a client to
// provide access to the DB Updates feed.
type DBUpdater interface {
	// DBUpdates must return a channel on which *DBUpdate events are sent,
	// and a function to close the connection.
	DBUpdates() (updateChan <-chan *DBUpdate, close func() error, err error)
}
