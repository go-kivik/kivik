package couchdb

import (
	"encoding/json"
	"runtime"

	"golang.org/x/net/context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

func (c *client) DBUpdates() (updateChan <-chan *driver.DBUpdate, closer func() error, err error) {
	resp, err := c.DoReq(context.Background(), kivik.MethodGet, "/_db_updates?feed=continuous", nil)
	if err != nil {
		return nil, nil, err
	}
	if err := chttp.ResponseError(resp.Response); err != nil {
		return nil, nil, err
	}
	feed := make(chan *driver.DBUpdate)
	done := make(chan struct{})
	closer = func() error {
		resp.Response.Body.Close()
		close(done)
		return nil
	}
	go func(feed chan<- *driver.DBUpdate, done <-chan struct{}) {
		dec := json.NewDecoder(resp.Response.Body)
		for dec.More() {
			var event *driver.DBUpdate
			err := dec.Decode(&event)
			event.Error = err
			select {
			case <-done:
				close(feed)
				runtime.Goexit()
			case feed <- event:
			}
		}
	}(feed, done)
	return feed, closer, nil
}
