package kivikmock

import "sync"

// DB serves to create expectations for database actions to mock and test real
// database behavior.
type DB struct {
	name   string
	id     int
	client *Client
	count  int
	mu     sync.RWMutex
}

func (db *DB) expectations() int {
	return db.count
}

// ExpectClose queues an expectation for DB.Close() to be called.
func (db *DB) ExpectClose() *ExpectedDBClose {
	e := &ExpectedDBClose{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}
