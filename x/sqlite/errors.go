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
	"errors"
	"strings"

	"modernc.org/sqlite"
)

const (
	// https://www.sqlite.org/rescode.html
	codeSQLiteError           = 1
	codeSQLiteConstraintCheck = 275
)

func errIsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	sqliteErr := new(sqlite.Error)
	if errors.As(err, &sqliteErr) &&
		sqliteErr.Code() == codeSQLiteError &&
		strings.Contains(sqliteErr.Error(), "already exists") {
		return true
	}
	return false
}

func errIsNoSuchTable(err error) bool {
	if err == nil {
		return false
	}
	sqliteErr := new(sqlite.Error)
	if errors.As(err, &sqliteErr) &&
		sqliteErr.Code() == codeSQLiteError &&
		strings.Contains(sqliteErr.Error(), "no such table") {
		return true
	}
	return false
}

func errIsInvalidCollation(err error) bool {
	if err == nil {
		return false
	}
	sqliteErr := new(sqlite.Error)
	if errors.As(err, &sqliteErr) &&
		sqliteErr.Code() == codeSQLiteConstraintCheck &&
		strings.Contains(sqliteErr.Error(), " collation IN ") {
		return true
	}
	return false
}
