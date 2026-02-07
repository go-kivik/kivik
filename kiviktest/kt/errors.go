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

package kt

// CheckError compares the error's status code with that expected.
func (c *Context) CheckError(err error) (match bool, success bool) {
	c.T.Helper()
	return c.ContextCore.CheckError(c.T, err)
}

// IsExpected checks the error against the expected status, and returns true
// if they match.
func (c *Context) IsExpected(err error) bool {
	c.T.Helper()
	return c.ContextCore.IsExpected(c.T, err)
}

// IsSuccess is similar to IsExpected, except for its return value. This method
// returns true if the expected status == 0, regardless of the error.
func (c *Context) IsSuccess(err error) bool {
	c.T.Helper()
	return c.ContextCore.IsSuccess(c.T, err)
}

// IsExpectedSuccess combines IsExpected() and IsSuccess(), returning true only
// if there is no error, and no error was expected.
func (c *Context) IsExpectedSuccess(err error) bool {
	c.T.Helper()
	return c.ContextCore.IsExpectedSuccess(c.T, err)
}
