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

package kivik_test

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4"
)

func ExampleHTTPStatus() {
	client, err := kivik.New("couch", "http://example.com:5984/")
	if err != nil {
		panic(err)
	}
	row := client.DB("foo").Get(context.Background(), "my_doc_id")
	switch err := row.Err(); kivik.HTTPStatus(err) {
	case http.StatusNotFound:
		return
	case http.StatusUnauthorized:
		panic("Authentication required")
	case http.StatusForbidden:
		panic("You are not authorized")
	default:
		panic("Unexpected error: " + err.Error())
	}
}
