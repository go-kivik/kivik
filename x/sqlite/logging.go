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

package sqlite

import (
	"database/sql"
	"log"

	"gitlab.com/flimzy/errsql"

	"github.com/go-kivik/kivik/v4"
)

type optionLogger struct {
	*log.Logger
}

var _ kivik.Option = (*optionLogger)(nil)

func (o optionLogger) Apply(target interface{}) {
	if client, ok := target.(*client); ok {
		client.logger = o.Logger
	}
}

// OptionLogger is an option to set a custom logger for the SQLite driver. The
// logger will be used to log any errors that occur during asynchronous
// operations such as background index rebuilding.
func OptionLogger(logger *log.Logger) kivik.Option {
	return optionLogger{Logger: logger}
}

// connector is used temporarily on startup, to connect to a database, and
// to possibly switch to the errsql driver if query logging is enabled.
type connector struct {
	dsn          string
	queryLogging bool
}

func (c *connector) Connect() (*sql.DB, error) {
	db, err := sql.Open("sqlite", c.dsn)
	if err != nil {
		return nil, err
	}
	if c.queryLogging {
		drv := errsql.NewWithHooks(db.Driver(), &errsql.Hooks{})
		cn, err := drv.OpenConnector(c.dsn)
		if err != nil {
			return nil, err
		}
		db = sql.OpenDB(cn)
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}
	return db, nil
}

type optionQueryLog struct{}

func (o optionQueryLog) Apply(target interface{}) {
	if cn, ok := target.(*connector); ok {
		cn.queryLogging = true
	}
}

// OptionQueryLog enables query logging for the SQLite driver. Query logs are
// sent to the logger (see [OptionLogger]) at DEBUG level.
func OptionQueryLog() kivik.Option {
	return optionQueryLog{}
}
