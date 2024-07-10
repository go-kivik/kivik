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
	"database/sql/driver"
	"fmt"
	"log"
	"strings"

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
	dsn         string
	queryLogger *log.Logger
}

func tx(e *errsql.Event) string {
	if e.InTransaction {
		return " (tx)"
	}
	return ""
}

func args(a []driver.NamedValue) string {
	result := make([]string, 0, len(a))
	for _, arg := range a {
		name := arg.Name
		if name == "" {
			name = fmt.Sprintf("$%d", arg.Ordinal)
		}
		result = append(result, fmt.Sprintf("%s=%v", name, arg.Value))
	}
	return strings.Join(result, ", ")
}

func (c *connector) Connect() (*sql.DB, error) {
	db, err := sql.Open("sqlite", c.dsn)
	if err != nil {
		return nil, err
	}
	if c.queryLogger != nil {
		drv := errsql.NewWithHooks(db.Driver(), &errsql.Hooks{
			ErrorHook: func(e *errsql.Event, err error) error {
				return fmt.Errorf("[%s.%s] %s", e.Entity, e.Method, err)
			},
			BeforePrepare: func(e *errsql.Event, query string) (string, error) {
				c.queryLogger.Printf("[%s.%s]%s Preparing query:\n%s\n", e.Entity, e.Method, tx(e), query)
				return query, nil
			},
			BeforePreparedQueryContext: func(e *errsql.Event, a []driver.NamedValue) ([]driver.NamedValue, error) {
				c.queryLogger.Printf("[%s.%s]%s arguments:\n%s\n", e.Entity, e.Method, tx(e), args(a))
				return a, nil
			},
			BeforeQueryContext: func(e *errsql.Event, query string, a []driver.NamedValue) (string, []driver.NamedValue, error) {
				c.queryLogger.Printf("[%s.%s]%s Query:\n%s\narguments:\n%s\n", e.Entity, e.Method, tx(e), query, args(a))
				return query, a, nil
			},
		})
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

type optionQueryLog struct {
	*log.Logger
}

func (o optionQueryLog) Apply(target interface{}) {
	if cn, ok := target.(*connector); ok {
		cn.queryLogger = o.Logger
	}
}

// OptionQueryLogger enables query logging for the SQLite driver. Query logs are
// sent to the logger (see [OptionLogger]) at DEBUG level.
func OptionQueryLogger(logger *log.Logger) kivik.Option {
	return optionQueryLog{Logger: logger}
}
