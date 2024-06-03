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

package kivik

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

// Session represents an authentication session.
type Session struct {
	// Name is the name of the authenticated user.
	Name string
	// Roles is a list of roles the user belongs to.
	Roles []string
	// AuthenticationMethod is the authentication method that was used for this
	// session.
	AuthenticationMethod string
	// AuthenticationDB is the user database against which authentication was
	// performed.
	AuthenticationDB string
	// AuthenticationHandlers is a list of authentication handlers configured on
	// the server.
	AuthenticationHandlers []string
	// RawResponse is the raw JSON response sent by the server, useful for
	// custom backends which may provide additional fields.
	RawResponse json.RawMessage
}

// Ensure types are equal.
var _ = Session(driver.Session{})

// Session returns information about the currently authenticated user.
func (c *Client) Session(ctx context.Context) (*Session, error) {
	endQuery, err := c.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	if sessioner, ok := c.driverClient.(driver.Sessioner); ok {
		session, err := sessioner.Session(ctx)
		if err != nil {
			return nil, err
		}
		ses := Session(*session)
		return &ses, nil
	}
	return nil, &internal.Error{Status: http.StatusNotImplemented, Message: "kivik: driver does not support sessions"}
}
