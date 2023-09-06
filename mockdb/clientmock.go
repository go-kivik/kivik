package kivikmock

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
)

// Client allows configuring the mock kivik client.
type Client struct {
	ordered    bool
	dsn        string
	opened     int
	drv        *mockDriver
	expected   []expectation
	newdbcount int
}

// nextExpectation accepts the expected value actual, checks that this is a
// valid expectation, and if so, populates actual with the matching expectation.
// If the expectation is not expected, an error is returned.
func (c *Client) nextExpectation(actual expectation) error {
	c.drv.Lock()
	defer c.drv.Unlock()

	var expected expectation
	var fulfilled int
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if c.ordered {
			if reflect.TypeOf(actual).Elem().Name() == reflect.TypeOf(next).Elem().Name() {
				if meets(actual, next) {
					expected = next
					break
				}
				next.Unlock()
				return fmt.Errorf("Expectation not met:\nExpected: %s\n  Actual: %s",
					next, actual)
			}
			next.Unlock()
			return fmt.Errorf("call to %s was not expected. Next expectation is: %s", actual.method(false), next.method(false))
		}
		if meets(actual, next) {
			expected = next
			break
		}

		next.Unlock()
	}

	if expected == nil {
		if fulfilled == len(c.expected) {
			return fmt.Errorf("call to %s was not expected, all expectations already fulfilled", actual.method(false))
		}
		return fmt.Errorf("call to %s was not expected", actual.method(!c.ordered))
	}

	defer expected.Unlock()
	expected.fulfill()

	reflect.ValueOf(actual).Elem().Set(reflect.ValueOf(expected).Elem())
	return nil
}

func (c *Client) open() (*kivik.Client, *Client, error) {
	client, err := kivik.New("kivikmock", c.dsn)
	return client, c, err
}

// ExpectationsWereMet returns an error if any outstanding expectatios were
// not met.
func (c *Client) ExpectationsWereMet() error {
	c.drv.Lock()
	defer c.drv.Unlock()
	for _, e := range c.expected {
		e.Lock()
		fulfilled := e.fulfilled()
		e.Unlock()

		if !fulfilled {
			return fmt.Errorf("there is a remaining unmet expectation: %s", e)
		}
	}
	return nil
}

// MatchExpectationsInOrder sets whether expectations should occur in the
// precise order in which they were defined.
func (c *Client) MatchExpectationsInOrder(b bool) {
	c.ordered = b
}

// ExpectAuthenticate queues an expectation for an Authenticate() call.
func (c *Client) ExpectAuthenticate() *ExpectedAuthenticate {
	e := &ExpectedAuthenticate{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectCreateDB queues an expectation for a CreateDB() call.
func (c *Client) ExpectCreateDB() *ExpectedCreateDB {
	e := &ExpectedCreateDB{}
	c.expected = append(c.expected, e)
	return e
}

// NewDB creates a new mock DB object, which can be used along with ExpectDB()
// or ExpectCreateDB() calls to mock database actions.
func (c *Client) NewDB() *DB {
	c.newdbcount++
	return &DB{
		client: c,
		id:     c.newdbcount,
	}
}

// NewRows returns a new, empty set of rows, which can be returned by any of
// the row-returning expectations.
func NewRows() *Rows {
	return &Rows{}
}

// NewChanges returns a new, empty changes set, which can be returned by the
// DB.Changes() expectation.
func NewChanges() *Changes {
	return &Changes{}
}

// NewDBUpdates returns a new, empty update set, which can be returned by the
// DBUpdates() expectation.
func NewDBUpdates() *Updates {
	return &Updates{}
}

// Replication is a replication instance.
type Replication struct {
	id        string
	source    string
	target    string
	startTime time.Time
	endTime   time.Time
	state     string
	err       error
}

// NewReplication returns a new, empty Replication.
func (c *Client) NewReplication() *Replication {
	return &Replication{}
}

func (r *Replication) MarshalJSON() ([]byte, error) {
	type rep struct {
		ID        string     `json:"replication_id,omitempty"`
		Source    string     `json:"source,omitempty"`
		Target    string     `json:"target,omitempty"`
		StartTime *time.Time `json:"start_time,omitempty"`
		EndTime   *time.Time `json:"end_time,omitempty"`
		State     string     `json:"state,omitempty"`
		Err       string     `json:"error,omitempty"`
	}
	doc := &rep{
		ID:     r.id,
		Source: r.source,
		Target: r.target,
		State:  r.state,
	}
	if !r.startTime.IsZero() {
		doc.StartTime = &r.startTime
	}
	if !r.endTime.IsZero() {
		doc.EndTime = &r.endTime
	}
	if r.err != nil {
		doc.Err = r.err.Error()
	}
	return json.Marshal(doc)
}

func (r *Replication) ID(id string) *Replication {
	r.id = id
	return r
}

func (r *Replication) Source(s string) *Replication {
	r.source = s
	return r
}

func (r *Replication) Target(t string) *Replication {
	r.target = t
	return r
}

func (r *Replication) StartTime(t time.Time) *Replication {
	r.startTime = t
	return r
}

func (r *Replication) EndTime(t time.Time) *Replication {
	r.endTime = t
	return r
}

func (r *Replication) State(s kivik.ReplicationState) *Replication {
	r.state = string(s)
	return r
}

func (r *Replication) Err(e error) *Replication {
	r.err = e
	return r
}
