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

// Package mockdb provides a full Kivik driver implementation, for mocking in
// tests.
package mockdb

//go:generate go run ./gen ./gen/templates
//go:generate gofmt -s -w clientexpectations_gen.go client_gen.go dbexpectations_gen.go db_gen.go clientmock_gen.go dbmock_gen.go
