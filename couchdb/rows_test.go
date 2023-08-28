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
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

const input = `
{
    "offset": 6,
    "rows": [
        {
            "id": "SpaghettiWithMeatballs",
            "key": "meatballs",
            "value": 1
        },
        {
            "id": "SpaghettiWithMeatballs",
            "key": "spaghetti",
            "value": 1
        },
        {
            "id": "SpaghettiWithMeatballs",
            "key": "tomato sauce",
            "value": 1
        }
    ],
    "total_rows": 3
}
`

var expectedKeys = []string{`"meatballs"`, `"spaghetti"`, `"tomato sauce"`}

func TestRowsIterator(t *testing.T) {
	rows := newRows(context.TODO(), io.NopCloser(strings.NewReader(input)))
	var count int
	for {
		row := &driver.Row{}
		err := rows.Next(row)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next() failed: %s", err)
		}
		if string(row.Key) != expectedKeys[count] {
			t.Errorf("Expected key #%d to be %s, got %s", count, expectedKeys[count], string(row.Key))
		}
		if count++; count > 10 {
			t.Fatalf("Ran too many iterations.")
		}
	}
	if count != 3 {
		t.Errorf("Expected 3 rows, got %d", count)
	}
	if rows.TotalRows() != 3 {
		t.Errorf("Expected TotalRows of 3, got %d", rows.TotalRows())
	}
	if rows.Offset() != 6 {
		t.Errorf("Expected Offset of 6, got %d", rows.Offset())
	}
	if err := rows.Next(&driver.Row{}); err != io.EOF {
		t.Errorf("Calling Next() after end returned unexpected error: %s", err)
	}
	if err := rows.Close(); err != nil {
		t.Errorf("Error closing rows iterator: %s", err)
	}
}

const multipleQueries = `{
    "results" : [
        {
            "offset": 0,
            "rows": [
                {
                    "id": "SpaghettiWithMeatballs",
                    "key": "meatballs",
                    "value": 1
                },
                {
                    "id": "SpaghettiWithMeatballs",
                    "key": "spaghetti",
                    "value": 1
                },
                {
                    "id": "SpaghettiWithMeatballs",
                    "key": "tomato sauce",
                    "value": 1
                }
            ],
            "total_rows": 3
        },
        {
            "offset" : 2,
            "rows" : [
                {
                    "id" : "Adukiandorangecasserole-microwave",
                    "key" : "Aduki and orange casserole - microwave",
                    "value" : [
                        null,
                        "Aduki and orange casserole - microwave"
                    ]
                },
                {
                    "id" : "Aioli-garlicmayonnaise",
                    "key" : "Aioli - garlic mayonnaise",
                    "value" : [
                        null,
                        "Aioli - garlic mayonnaise"
                    ]
                },
                {
                    "id" : "Alabamapeanutchicken",
                    "key" : "Alabama peanut chicken",
                    "value" : [
                        null,
                        "Alabama peanut chicken"
                    ]
                }
            ],
            "total_rows" : 2667
        }
    ]
}`

func TestMultiQueriesRowsIterator(t *testing.T) {
	rows := newMultiQueriesRows(context.TODO(), io.NopCloser(strings.NewReader(multipleQueries)))
	results := make([]interface{}, 0, 8)
	for {
		row := &driver.Row{}
		err := rows.Next(row)
		if err == driver.EOQ {
			results = append(results, map[string]interface{}{
				"EOQ":        true,
				"total_rows": rows.TotalRows(),
				"offset":     rows.Offset(),
			})
			continue
		}
		if err == io.EOF {
			results = append(results, map[string]interface{}{
				"EOF":        true,
				"total_rows": rows.TotalRows(),
				"offset":     rows.Offset(),
			})
			break
		}
		if err != nil {
			t.Fatalf("Next() failed: %s", err)
		}
		results = append(results, map[string]interface{}{
			"key": row.Key,
		})
	}
	if d := testy.DiffInterface(testy.Snapshot(t), results); d != nil {
		t.Error(d)
	}
	if err := rows.Next(&driver.Row{}); err != io.EOF {
		t.Errorf("Calling Next() after end returned unexpected error: %s", err)
	}
	if err := rows.Close(); err != nil {
		t.Errorf("Error closing rows iterator: %s", err)
	}
}

func TestRowsIteratorErrors(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		status int
		err    string
	}{
		{
			name:   "empty input",
			input:  "",
			status: http.StatusBadGateway,
			err:    "EOF",
		},
		{
			name:   "unexpected delimiter",
			input:  "[]",
			status: http.StatusBadGateway,
			err:    "Unexpected JSON delimiter: [",
		},
		{
			name:   "unexpected input",
			input:  `"foo"`,
			status: http.StatusBadGateway,
			err:    "Unexpected token string: foo",
		},
		{
			name:   "missing closing delimiter",
			input:  `{"rows":[{"id":"1","key":"1","value":1}`,
			status: http.StatusBadGateway,
			err:    "EOF",
		},
		{
			name:   "unexpected key",
			input:  `{"foo":"bar","rows":[]}`,
			status: http.StatusInternalServerError,
			err:    "EOF",
		},
		{
			name:   "unexpected key after valid row",
			input:  `{"rows":[{"id":"1","key":"1","value":1}],"foo":"bar"}`,
			status: http.StatusInternalServerError,
			err:    "EOF",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := newRows(context.TODO(), io.NopCloser(strings.NewReader(test.input)))
			for i := 0; i < 10; i++ {
				err := rows.Next(&driver.Row{})
				if err == nil {
					continue
				}
				testy.StatusError(t, test.err, test.status, err)
			}
		})
	}
}

const findInput = `
{"warning":"no matching index found, create an index to optimize query time",
"docs":[
{"id":"SpaghettiWithMeatballs","key":"meatballs","value":1},
{"id":"SpaghettiWithMeatballs","key":"spaghetti","value":1},
{"id":"SpaghettiWithMeatballs","key":"tomato sauce","value":1}
],
"bookmark": "nil"
}
`

type fullRows interface {
	driver.Rows
	driver.RowsWarner
	driver.Bookmarker
}

func TestFindRowsIterator(t *testing.T) {
	rows := newFindRows(context.TODO(), io.NopCloser(strings.NewReader(findInput))).(fullRows)
	var count int
	for {
		row := &driver.Row{}
		err := rows.Next(row)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next() failed: %s", err)
		}
		if count++; count > 10 {
			t.Fatalf("Ran too many iterations.")
		}
	}
	if count != 3 {
		t.Errorf("Expected 3 rows, got %d", count)
	}
	if err := rows.Next(&driver.Row{}); err != io.EOF {
		t.Errorf("Calling Next() after end returned unexpected error: %s", err)
	}
	if err := rows.Close(); err != nil {
		t.Errorf("Error closing rows iterator: %s", err)
	}
	if rows.Warning() != "no matching index found, create an index to optimize query time" {
		t.Errorf("Unexpected warning: %s", rows.Warning())
	}
	if rows.Bookmark() != "nil" {
		t.Errorf("Unexpected bookmark: %s", rows.Bookmark())
	}
}
