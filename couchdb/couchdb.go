// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package couchdb

import (
	"net/http"
	"net/url"
	"sync"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
)

// couch represents the parent driver instance.
type couch struct{}

var _ driver.Driver = &couch{}

func init() {
	kivik.Register("couch", &couch{})
}

type client struct {
	*chttp.Client

	// schedulerDetected will be set once the scheduler has been detected.
	// It should only be accessed through the schedulerSupported() method.
	schedulerDetected *bool
	sdMU              sync.Mutex
}

var (
	_ driver.Client    = &client{}
	_ driver.DBUpdater = &client{}
)

func (d *couch) NewClient(dsn string, options driver.Options) (driver.Client, error) {
	httpClient := &http.Client{}
	options.Apply(httpClient)
	chttpClient, err := chttp.New(httpClient, dsn, options)
	if err != nil {
		return nil, err
	}
	return &client{
		Client: chttpClient,
	}, nil
}

func (c *client) DB(dbName string, _ driver.Options) (driver.DB, error) {
	if dbName == "" {
		return nil, missingArg("dbName")
	}
	return &db{
		client: c,
		dbName: url.PathEscape(dbName),
	}, nil
}
