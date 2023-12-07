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

package auth

import "context"

// A UserStore provides an AuthHandler with access to a user store for.
type UserStore interface {
	// Validate returns a user context object if the credentials are valid. An
	// error must be returned otherwise. A Not Found error must not be returned.
	// Not Found should be treated identically to Unauthorized.
	Validate(ctx context.Context, username, password string) (user *UserContext, err error)
	// UserCtx returns a user context object if the user exists. It is used by
	// AuthHandlers that don't validate the password (e.g. Cookie auth).
	UserCtx(ctx context.Context, username string) (user *UserContext, err error)
}
