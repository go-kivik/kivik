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
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-kivik/couchdb/v4/chttp"
	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

// couch represents the parent driver instance.
type couch struct{}

var _ driver.Driver = &couch{}

func init() {
	kivik.Register("couch", &couch{})
}

// Known vendor strings
const (
	VendorCouchDB  = "The Apache Software Foundation"
	VendorCloudant = "IBM Cloudant"
)

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

func (d *couch) NewClient(dsn string, options map[string]interface{}) (driver.Client, error) {
	var httpClient *http.Client
	if c, ok := options[OptionHTTPClient]; ok {
		if httpClient, ok = c.(*http.Client); !ok {
			return nil, &kivik.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("OptionHTTPClient is %T, must be *http.Client", c)}
		}
	}
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	chttpClient, err := chttp.New(httpClient, dsn, options)
	if err != nil {
		return nil, err
	}
	return &client{
		Client: chttpClient,
	}, nil
}

func (c *client) DB(dbName string, _ map[string]interface{}) (driver.DB, error) {
	if dbName == "" {
		return nil, missingArg("dbName")
	}
	return &db{
		client: c,
		dbName: url.PathEscape(dbName),
	}, nil
}
