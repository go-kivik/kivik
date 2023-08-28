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
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
)

type session struct {
	Data    json.RawMessage
	Info    authInfo    `json:"info"`
	UserCtx userContext `json:"userCtx"`
}

type authInfo struct {
	AuthenticationMethod   string   `json:"authenticated"`
	AuthenticationDB       string   `json:"authentiation_db"`
	AuthenticationHandlers []string `json:"authentication_handlers"`
}

type userContext struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

func (s *session) UnmarshalJSON(data []byte) error {
	type alias session
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*s = session(a)
	s.Data = data
	return nil
}

func (c *client) Session(ctx context.Context) (*driver.Session, error) {
	s := &session{}
	err := c.DoJSON(ctx, http.MethodGet, "/_session", nil, s)
	return &driver.Session{
		RawResponse:            s.Data,
		Name:                   s.UserCtx.Name,
		Roles:                  s.UserCtx.Roles,
		AuthenticationMethod:   s.Info.AuthenticationMethod,
		AuthenticationDB:       s.Info.AuthenticationDB,
		AuthenticationHandlers: s.Info.AuthenticationHandlers,
	}, err
}
